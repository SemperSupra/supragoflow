package workcell

import (
 "crypto/rand"
 "encoding/hex"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func newRoot(t *testing.T) string{
 t.Helper()
 b:=make([]byte,12)
 if _,err:=rand.Read(b);err!=nil{t.Fatal(err)}
 root:=filepath.Join(os.TempDir(),"public-rdte-"+hex.EncodeToString(b))
 if err:=Init(root);err!=nil{t.Fatal(err)}
 t.Cleanup(func(){os.RemoveAll(root)})
 return root
}
func TestFullCycleAndReplay(t *testing.T){
 root:=newRoot(t)
 p,err:=MakePlan(root);if err!=nil{t.Fatal(err)}
 q,err:=MakePlan(root);if err!=nil{t.Fatal(err)}
 if p.Digest!=q.Digest{t.Fatal("non deterministic plan")}
 if err=Apply(root,p,p.Digest,"");err!=nil{t.Fatal(err)}
 if err=Verify(root);err!=nil{t.Fatal(err)}
 if err=Apply(root,p,p.Digest,"");err!=nil{t.Fatal("idempotent reapply:",err)}
 if err=Rollback(root,p.Digest);err!=nil{t.Fatal(err)}
 if err=Rollback(root,p.Digest);err!=nil{t.Fatal("idempotent rollback:",err)}
 o,err:=Observe(root);if err!=nil||o.State!="ABSENT"{t.Fatalf("rollback state: %+v, %v",o,err)}
 if err=Destroy(root,p.Digest);err!=nil{t.Fatal(err)}
 if _,err=os.Stat(root);!os.IsNotExist(err){t.Fatal("destroy left synthetic workspace")}
}
func TestWrongApprovalAndPlanTampering(t *testing.T){
 root:=newRoot(t)
 p,_:=MakePlan(root)
 if err:=Apply(root,p,strings.Repeat("0",64),"");err==nil{t.Fatal("wrong approval accepted")}
 p.After="fake"
 if err:=Apply(root,p,p.Digest,"");err==nil{t.Fatal("tampered plan accepted")}
 o,_:=Observe(root);if o.State!="ABSENT"{t.Fatal("denial wrote payload")}
}
func TestFaultBeforeAndAfterWriteAndRecovery(t *testing.T){
 for _,f:=range []string{"before-write","after-write"}{t.Run(f,func(t *testing.T){
  root:=newRoot(t);p,_:=MakePlan(root)
  if err:=Apply(root,p,p.Digest,f);err==nil{t.Fatal("fault failed to stop")}
  if err:=Verify(root);err==nil{t.Fatal("uncommitted journal verified")}
  if err:=Apply(root,p,p.Digest,"");err==nil{t.Fatal("incomplete state retried blindly")}
  if err:=Rollback(root,p.Digest);err!=nil{t.Fatal(err)}
  if err:=Destroy(root,p.Digest);err!=nil{t.Fatal(err)}
 })}
}
func TestUserEditAndForeignFileProtected(t *testing.T){
 root:=newRoot(t);p,_:=MakePlan(root)
 if err:=Apply(root,p,p.Digest,"");err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(root,"payload.txt"),[]byte("owned edit"),0600);err!=nil{t.Fatal(err)}
 if err:=Rollback(root,p.Digest);err==nil{t.Fatal("refused not to clobber user edit")}
 if err:=Destroy(root,p.Digest);err==nil{t.Fatal("destroy clobbered edited workcell")}
}
func TestForeignFilePreventsDestroy(t *testing.T){
 root:=newRoot(t)
 if err:=os.WriteFile(filepath.Join(root,"foreign.txt"),[]byte("retain"),0600);err!=nil{t.Fatal(err)}
 if err:=Destroy(root,"");err==nil{t.Fatal("destroy removed foreign file")}
 if _,err:=os.Stat(filepath.Join(root,"foreign.txt"));err!=nil{t.Fatal("foreign file lost")}
}
func TestRefusePreexistingOrNonTempRoot(t *testing.T){
 root:=newRoot(t)
 if err:=Init(root);err==nil{t.Fatal("preexisting root adopted")}
 if err:=Init(filepath.Join(filepath.Dir(filepath.Clean(os.TempDir())),"not-a-temporary-root"));err==nil{t.Fatal("non-temp root allowed")}
}
func TestRefuseSymlinkAsPayload(t *testing.T){
 root:=newRoot(t)
 dst:=filepath.Join(root,"payload.txt")
 if err:=os.Symlink("some-target",dst);err!=nil{t.Skip("symlinks unavailable to this Windows account:",err)}
 if _,err:=MakePlan(root);err==nil{t.Fatal("symlink accepted")}
}
func TestRejectJournalDigestMismatch(t *testing.T){
 root:=newRoot(t);p,_:=MakePlan(root)
 if err:=Apply(root,p,p.Digest,"");err!=nil{t.Fatal(err)}
 if err:=Rollback(root,strings.Repeat("a",64));err==nil{t.Fatal("mismatched journal accepted")}
}
