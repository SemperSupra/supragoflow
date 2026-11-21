package lease

import ("errors";"sync";"testing";"time")
func req()Request{return Request{Venue:"gha",Workset:"synthetic",Target:"dummy",Method:"read-only-ssh",Nonce:"one",Generation:1,TTL:time.Minute}}
func TestLifecycleAndDeny(t *testing.T){
 now:=time.Unix(100,0);s:=New();r:=req()
 if _,err:=s.Issue(r,now,true);err!=nil{t.Fatal(err)}
 if _,err:=s.Issue(r,now,true);!errors.Is(err,ErrDenied){t.Fatal("replay allowed")}
 if _,err:=s.Check("one",1,now.Add(time.Second));err!=nil{t.Fatal(err)}
 if _,err:=s.Check("one",2,now);!errors.Is(err,ErrDenied){t.Fatal("wrong generation allowed")}
 if _,err:=s.Renew("one",1,now,time.Minute,false);!errors.Is(err,ErrDenied){t.Fatal("unadmitted renewal")}
 if _,err:=s.Renew("one",1,now,time.Minute,true);err!=nil{t.Fatal(err)}
 r.Generation=2;r.Nonce="two"
 if _,err:=s.Issue(r,now,true);err!=nil{t.Fatal(err)}
 if _,err:=s.Check("one",1,now);!errors.Is(err,ErrDenied){t.Fatal("old generation active")}
 s.Close("two");if _,err:=s.Check("two",2,now);!errors.Is(err,ErrDenied){t.Fatal("closed lease active")}
}
func TestSecurityMatrix(t *testing.T){
 now:=time.Unix(100,0)
 tests:=[]func(*Request){func(r *Request){r.Venue="unknown"},func(r *Request){r.Target=""},func(r *Request){r.Method="arbitrary-shell"},func(r *Request){r.TTL=16*time.Minute},func(r *Request){r.Generation=0},func(r *Request){r.Nonce=""}}
 for i,mut:=range tests {s:=New();r:=req();mut(&r);if _,err:=s.Issue(r,now,true);!errors.Is(err,ErrDenied){t.Fatalf("case %d admitted",i)}}
 s:=New();if _,err:=s.Issue(req(),now,false);!errors.Is(err,ErrDenied){t.Fatal("unadmitted")}
 if _,err:=s.Check("missing",1,now);!errors.Is(err,ErrDenied){t.Fatal("missing")}
 if _,err:=s.Issue(req(),now,true);err!=nil{t.Fatal(err)}
 if _,err:=s.Check("one",1,now.Add(time.Minute));!errors.Is(err,ErrDenied){t.Fatal("expired")}
 if _,err:=s.Renew("one",1,now.Add(time.Minute),time.Minute,true);!errors.Is(err,ErrDenied){t.Fatal("resurrected")}
}
func TestConcurrentReplay(t *testing.T){
 s:=New();now:=time.Unix(100,0);var wg sync.WaitGroup;results:=make(chan bool,32)
 for i:=0;i<32;i++{wg.Add(1);go func(){defer wg.Done();_,err:=s.Issue(req(),now,true);results<-err==nil}()}
 wg.Wait();close(results);count:=0;for ok:=range results{if ok{count++}};if count!=1{t.Fatalf("issued %d times",count)}
}
