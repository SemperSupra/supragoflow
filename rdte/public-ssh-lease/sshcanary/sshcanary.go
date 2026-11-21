// Package sshcanary qualifies ephemeral SSH user certs on loopback only.
// It does not create an issuer, a credential file, or any external connection.
package sshcanary

import (
 "bytes"
 "crypto/ed25519"
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net"
 "regexp"
 "sync"
 "time"

 "golang.org/x/crypto/ssh"
)

var sha40=regexp.MustCompile("^[a-f0-9]{40}$")

type Claim struct {
 Workset string `json:"workset"`
 SourceSHA string `json:"source_sha"`
 Venue string `json:"venue"`
 Target string `json:"target"`
 Action string `json:"action"`
 Nonce string `json:"nonce"`
 Generation uint64 `json:"generation"`
}
func (c Claim) Validate()error{
 if c.Workset!="synthetic-rdte" || !sha40.MatchString(c.SourceSHA) ||
 c.Venue!="public-gha" || c.Target!="loopback-only" ||
 c.Action!="canary-health" || len(c.Nonce)!=32 || c.Generation==0{return errors.New("invalid public synthetic scope")}
 n,e:=hex.DecodeString(c.Nonce);if e!=nil||len(n)!=16{return errors.New("invalid synthetic nonce")}
 return nil
}
func NewClaim()(Claim,error){
 b:=make([]byte,16)
 if _,e:=rand.Read(b);e!=nil{return Claim{},e}
 return Claim{Workset:"synthetic-rdte",SourceSHA:"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",Venue:"public-gha",Target:"loopback-only",Action:"canary-health",Nonce:hex.EncodeToString(b),Generation:1},nil
}
type Gate struct {
 mu sync.Mutex
 expected Claim
 consumed bool
 revoked bool
}
func NewGate(c Claim)(*Gate,error){
 if e:=c.Validate();e!=nil{return nil,e}
 return &Gate{expected:c},nil
}
func (g *Gate) Check(cert *ssh.Certificate)error{
 if cert==nil || cert.CertType!=ssh.UserCert{return errors.New("user certificate required")}
 if cert.Serial!=g.expected.Generation || len(cert.ValidPrincipals)!=1 || cert.ValidPrincipals[0]!="public-canary"{
  return errors.New("wrong principal or generation")
 }
 now:=uint64(time.Now().Unix())
 if cert.ValidAfter>now || cert.ValidBefore<=now || cert.ValidBefore<cert.ValidAfter || cert.ValidBefore-cert.ValidAfter>900{
  return errors.New("certificate expired or TTL exceeded")
 }
 if cert.Permissions.Extensions["permit-canary-health"]!="1" || len(cert.Permissions.CriticalOptions)!=0{
  return errors.New("missing fixed action permit or unexpected critical option")
 }
 var signed Claim
 d:=json.NewDecoder(bytes.NewReader([]byte(cert.KeyId)));d.DisallowUnknownFields()
 if e:=d.Decode(&signed);e!=nil{return e}
 if _,e:=d.Token();e!=io.EOF{return errors.New("trailing claims")}
 if signed!=g.expected{return errors.New("mismatched signed scope")}
 g.mu.Lock();defer g.mu.Unlock()
 if g.consumed||g.revoked{return errors.New("lease consumed or revoked")}
 return nil
}
func (g *Gate) Consume()bool{
 g.mu.Lock();defer g.mu.Unlock()
 if g.consumed||g.revoked{return false}
 g.consumed=true
 return true
}
func (g *Gate) Revoke(){g.mu.Lock();g.revoked=true;g.mu.Unlock()}

func FreshSigner()(ssh.Signer,error){
 _,key,e:=ed25519.GenerateKey(rand.Reader);if e!=nil{return nil,e}
 return ssh.NewSignerFromKey(key)
}
func Mint(c Claim,user,ca ssh.Signer,start time.Time,ttl time.Duration)(ssh.Signer,error){
 if e:=c.Validate();e!=nil{return nil,e}
 if ttl<=0||ttl>15*time.Minute{return nil,errors.New("TTL out of bounds")}
 b,_:=json.Marshal(c)
 cert:=&ssh.Certificate{
  Key:user.PublicKey(),Serial:c.Generation,CertType:ssh.UserCert,
  KeyId:string(b),ValidPrincipals:[]string{"public-canary"},
  ValidAfter:uint64(start.Unix()),ValidBefore:uint64(start.Add(ttl).Unix()),
  Permissions:ssh.Permissions{Extensions:map[string]string{"permit-canary-health":"1"}},
 }
 if e:=cert.SignCert(rand.Reader,ca);e!=nil{return nil,e}
 return ssh.NewCertSigner(cert,user)
}

type Lab struct {
 listener net.Listener
 gate *Gate
 hostKey ssh.PublicKey
 closed chan struct{}
}
func Start(g *Gate,ca,host ssh.Signer)(*Lab,error){
 if g==nil||ca==nil||host==nil{return nil,errors.New("missing local trust")}
 ln,e:=net.Listen("tcp4","127.0.0.1:0");if e!=nil{return nil,e}
 cfg:=&ssh.ServerConfig{PublicKeyCallback:func(meta ssh.ConnMetadata,pub ssh.PublicKey)(*ssh.Permissions,error){
  if meta.User()!="public-canary"{return nil,errors.New("wrong user")}
  cert,ok:=pub.(*ssh.Certificate);if !ok{return nil,errors.New("no user certificate")}
  checker:=ssh.CertChecker{IsUserAuthority:func(key ssh.PublicKey)bool{return bytes.Equal(key.Marshal(),ca.PublicKey().Marshal())}}
  if e:=checker.CheckCert(meta.User(),cert);e!=nil{return nil,e}
  if e:=g.Check(cert);e!=nil{return nil,e}
  return &ssh.Permissions{},nil
 }}
 cfg.AddHostKey(host)
 lab:=&Lab{listener:ln,gate:g,hostKey:host.PublicKey(),closed:make(chan struct{})}
 go func(){
  defer close(lab.closed)
  for {
   c,e:=ln.Accept();if e!=nil{return}
   go lab.handle(c,cfg)
  }
 }()
 return lab,nil
}
func(l *Lab)Address()string{return l.listener.Addr().String()}
func(l *Lab)HostKey()ssh.PublicKey{return l.hostKey}
func(l *Lab)Close(){l.listener.Close();<-l.closed}
func(l *Lab)handle(conn net.Conn,cfg *ssh.ServerConfig){
 defer conn.Close()
 s,channels,requests,e:=ssh.NewServerConn(conn,cfg);if e!=nil{return}
 defer s.Close()
 go ssh.DiscardRequests(requests)
 for ch:=range channels{
  if ch.ChannelType()!="session"{ch.Reject(ssh.UnknownChannelType,"fixed session only");continue}
  channel,reqs,e:=ch.Accept();if e!=nil{continue}
  go l.session(channel,reqs)
 }
}
func(l *Lab)session(channel ssh.Channel,reqs <-chan *ssh.Request){
 defer channel.Close()
 for req:=range reqs{
  if req.Type!="exec"{req.Reply(false,nil);continue}
  var body struct{Command string}
  if e:=ssh.Unmarshal(req.Payload,&body);e!=nil{req.Reply(false,nil);return}
  approved:=body.Command=="canary-health"&&l.gate.Consume()
  req.Reply(true,nil)
  status:=uint32(1)
  if approved {
   _,_ = io.WriteString(channel,"PUBLIC_LOOPBACK_CANARY_OK\n")
   status=0
  }
  channel.SendRequest("exit-status",false,ssh.Marshal(&struct{Status uint32}{status}))
  return
 }
}
func(l *Lab)Dial(cert ssh.Signer,pin ssh.PublicKey)(*ssh.Client,error){
 if cert==nil||pin==nil{return nil,errors.New("missing signer or host pin")}
 cfg:=&ssh.ClientConfig{User:"public-canary",Auth:[]ssh.AuthMethod{ssh.PublicKeys(cert)},HostKeyCallback:ssh.FixedHostKey(pin),Timeout:5*time.Second}
 return ssh.Dial("tcp",l.Address(),cfg)
}
func Run()error{
 claim,e:=NewClaim();if e!=nil{return e}
 gate,e:=NewGate(claim);if e!=nil{return e}
 ca,e:=FreshSigner();if e!=nil{return e}
 user,e:=FreshSigner();if e!=nil{return e}
 host,e:=FreshSigner();if e!=nil{return e}
 signer,e:=Mint(claim,user,ca,time.Now().Add(-time.Second),10*time.Minute);if e!=nil{return e}
 lab,e:=Start(gate,ca,host);if e!=nil{return e};defer lab.Close()
 client,e:=lab.Dial(signer,lab.HostKey());if e!=nil{return e};defer client.Close()
 one,e:=client.NewSession();if e!=nil{return e}
 out,e:=one.CombinedOutput("canary-health");one.Close()
 if e!=nil||string(out)!="PUBLIC_LOOPBACK_CANARY_OK\n"{return fmt.Errorf("positive control failed: %v",e)}
 two,e:=client.NewSession();if e!=nil{return e}
 _,e=two.CombinedOutput("canary-health");two.Close()
 if e==nil{return errors.New("replayed SSH cert command succeeded")}
 gate.Revoke()
 if _,e:=lab.Dial(signer,lab.HostKey());e==nil{return errors.New("revoked certificate accepted")}
 return nil
}
