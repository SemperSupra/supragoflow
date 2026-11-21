package sshcanary
import(
 "strings"
 "sync"
 "sync/atomic"
 "testing"
 "time"
 "golang.org/x/crypto/ssh"
)
type fixture struct{lab *Lab; gate *Gate;ca,user,host,signer ssh.Signer;claim Claim}
func fresh(t *testing.T)*fixture{
 t.Helper()
 c,e:=NewClaim();if e!=nil{t.Fatal(e)}
 g,e:=NewGate(c);if e!=nil{t.Fatal(e)}
 ca,e:=FreshSigner();if e!=nil{t.Fatal(e)}
 user,e:=FreshSigner();if e!=nil{t.Fatal(e)}
 host,e:=FreshSigner();if e!=nil{t.Fatal(e)}
 signer,e:=Mint(c,user,ca,time.Now().Add(-time.Second),10*time.Minute);if e!=nil{t.Fatal(e)}
 l,e:=Start(g,ca,host);if e!=nil{t.Fatal(e)}
 t.Cleanup(l.Close)
 return &fixture{lab:l,gate:g,ca:ca,user:user,host:host,signer:signer,claim:c}
}
func execute(c *ssh.Client,cmd string)(string,error){
 s,e:=c.NewSession();if e!=nil{return "",e};defer s.Close()
 out,e:=s.CombinedOutput(cmd);return string(out),e
}
func TestLiveHandshakeAndOneUse(t *testing.T){
 f:=fresh(t)
 c,e:=f.lab.Dial(f.signer,f.lab.HostKey());if e!=nil{t.Fatal(e)};defer c.Close()
 out,e:=execute(c,"canary-health");if e!=nil||out!="PUBLIC_LOOPBACK_CANARY_OK\n"{t.Fatalf("positive failed %q %v",out,e)}
 if _,e=execute(c,"canary-health");e==nil{t.Fatal("replayed command was accepted")}
 if another,e:=f.lab.Dial(f.signer,f.lab.HostKey());e==nil{another.Close();t.Fatal("replayed cert accepted on new connection")}
}
func TestRevokeAfterHandshake(t *testing.T){
 f:=fresh(t)
 c,e:=f.lab.Dial(f.signer,f.lab.HostKey());if e!=nil{t.Fatal(e)};defer c.Close()
 f.gate.Revoke()
 if _,e:=execute(c,"canary-health");e==nil{t.Fatal("existing session acted after revocation")}
 if another,e:=f.lab.Dial(f.signer,f.lab.HostKey());e==nil{another.Close();t.Fatal("revoked signer reconnected")}
}
func TestNegativeTrustAndScope(t *testing.T){
 for _,kind:=range []string{"foreign-ca","wrong-nonce","wrong-generation","expired-cert","wrong-host","revoked-before-auth"}{
  t.Run(kind,func(t *testing.T){
   f:=fresh(t)
   signer,pin:=f.signer,f.lab.HostKey()
   var e error
   switch kind{
   case "foreign-ca":
     foreign,e:=FreshSigner();if e!=nil{t.Fatal(e)}
     signer,e=Mint(f.claim,f.user,foreign,time.Now().Add(-time.Second),10*time.Minute)
   case "wrong-nonce":
     c:=f.claim;c.Nonce=strings.Repeat("a",32)
     if c.Nonce==f.claim.Nonce{c.Nonce=strings.Repeat("b",32)}
     signer,e=Mint(c,f.user,f.ca,time.Now().Add(-time.Second),10*time.Minute)
   case "wrong-generation":
     c:=f.claim;c.Generation++
     signer,e=Mint(c,f.user,f.ca,time.Now().Add(-time.Second),10*time.Minute)
   case "expired-cert":
     signer,e=Mint(f.claim,f.user,f.ca,time.Now().Add(-30*time.Minute),10*time.Minute)
   case "wrong-host":
     bad,e2:=FreshSigner();if e2!=nil{t.Fatal(e2)};pin=bad.PublicKey()
   case "revoked-before-auth":f.gate.Revoke()
   }
   if e!=nil{t.Fatal(e)}
   if c,e:=f.lab.Dial(signer,pin);e==nil{c.Close();t.Fatal("invalid certificate/host admitted")}
  })
 }
}
func TestNoShellAndFixedActionOnly(t *testing.T){
 f:=fresh(t)
 c,e:=f.lab.Dial(f.signer,f.lab.HostKey());if e!=nil{t.Fatal(e)};defer c.Close()
 s,e:=c.NewSession();if e!=nil{t.Fatal(e)}
 if e=s.Shell();e==nil{t.Fatal("shell admitted")}
 s.Close()
 if _,e:=execute(c,"sh -c whoami");e==nil{t.Fatal("generic shell command admitted")}
 out,e:=execute(c,"canary-health");if e!=nil||out!="PUBLIC_LOOPBACK_CANARY_OK\n"{t.Fatalf("rejected legitimate canary after wrong command: %s %v",out,e)}
}
func TestOneWinnerConcurrentSessions(t *testing.T){
 f:=fresh(t)
 var allowed atomic.Int32
 var wg sync.WaitGroup
 for i:=0;i<8;i++{
  wg.Add(1)
  go func(){
   defer wg.Done()
   c,e:=f.lab.Dial(f.signer,f.lab.HostKey());if e!=nil{return}
   defer c.Close()
   out,e:=execute(c,"canary-health")
   if e==nil&&out=="PUBLIC_LOOPBACK_CANARY_OK\n"{allowed.Add(1)}
  }()
 }
 wg.Wait()
 if allowed.Load()!=1{t.Fatalf("expected exactly one accepted action, got %d",allowed.Load())}
}
func TestRejectInvalidTTLAndClaim(t *testing.T){
 f:=fresh(t)
 if _,e:=Mint(f.claim,f.user,f.ca,time.Now(),16*time.Minute);e==nil{t.Fatal("excessive TTL admitted")}
 c:=f.claim;c.Target="real-host"
 if _,e:=Mint(c,f.user,f.ca,time.Now(),time.Minute);e==nil{t.Fatal("real-host scope allowed by public canary")}
 if e:=Run();e!=nil{t.Fatal(e)}
}
