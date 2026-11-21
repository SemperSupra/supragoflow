// Package lease provides a synthetic, provider-free contract for disposable workcells.
// It does not authenticate callers, mint credentials, or grant production authority.
package lease

import (
 "errors"
 "sync"
 "time"
)
var ErrDenied = errors.New("lease denied")
type Request struct { Venue, Workset, Target, Method, Nonce string; Generation uint64; TTL time.Duration }
type Lease struct { Request Request; Issued, Expires time.Time; State string }
type Store struct { mu sync.Mutex; leases map[string]Lease; fence map[string]uint64; used map[string]bool }
func New() *Store { return &Store{leases:map[string]Lease{}, fence:map[string]uint64{}, used:map[string]bool{}} }
func key(r Request) string { return r.Venue+"\x00"+r.Workset+"\x00"+r.Target }
func (s *Store) Issue(r Request, now time.Time, admitted bool) (Lease,error) {
 s.mu.Lock(); defer s.mu.Unlock()
 k:=key(r)
 if !admitted || r.Venue=="" || r.Workset=="" || r.Target=="" || r.Method=="" || r.Nonce=="" || r.Generation==0 || r.TTL<=0 || r.TTL>15*time.Minute || !allowed(r.Venue,r.Method) || s.used[r.Nonce] || r.Generation<=s.fence[k] { return Lease{},ErrDenied }
 s.fence[k]=r.Generation; s.used[r.Nonce]=true
 for id,old:=range s.leases { if key(old.Request)==k && old.State=="ACTIVE" { old.State="FENCED";s.leases[id]=old } }
 l:=Lease{Request:r,Issued:now,Expires:now.Add(r.TTL),State:"ACTIVE"}; s.leases[r.Nonce]=l;return l,nil
}
func allowed(venue,method string)bool{
 switch venue {case "gha","colab","kaggle":return method=="read-only-ssh"||method=="read-only-oci";default:return false}
}
func (s *Store) Check(nonce string, generation uint64, now time.Time)(Lease,error){
 s.mu.Lock();defer s.mu.Unlock()
 l,ok:=s.leases[nonce];if !ok||l.State!="ACTIVE"||l.Request.Generation!=generation||s.fence[key(l.Request)]!=generation{return Lease{},ErrDenied}
 if !now.Before(l.Expires){l.State="EXPIRED";s.leases[nonce]=l;return Lease{},ErrDenied}
 return l,nil
}
func(s *Store) Close(nonce string){s.mu.Lock();defer s.mu.Unlock();if l,ok:=s.leases[nonce];ok {l.State="CLOSED";s.leases[nonce]=l}}
func(s *Store) Renew(nonce string,generation uint64,now time.Time,ttl time.Duration,admitted bool)(Lease,error){
 s.mu.Lock();defer s.mu.Unlock()
 l,ok:=s.leases[nonce]
 if !admitted||!ok||l.State!="ACTIVE"||l.Request.Generation!=generation||s.fence[key(l.Request)]!=generation||!now.Before(l.Expires)||ttl<=0||ttl>15*time.Minute{return Lease{},ErrDenied}
 l.Expires=now.Add(ttl);s.leases[nonce]=l;return l,nil
}
