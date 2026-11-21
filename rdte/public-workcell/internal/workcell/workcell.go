// Package workcell is a deliberately SYNTHETIC RDTE workcell transaction.
// It never invokes SSH, Hyper-V, a package manager, provider API, or a service.
// Do not use it as a host installer or as an authorization system.
package workcell

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
)
const marker = "PUBLIC_SYNTHETIC_RDTE_V1"
const payload = "public synthetic workcell payload v1\n"
type Plan struct {
 Schema int `json:"schema"`
 Root string `json:"root"`
 Before string `json:"before_sha256"`
 After string `json:"after_sha256"`
 Digest string `json:"digest"`
}
type Journal struct {
 Phase string `json:"phase"`
 Digest string `json:"digest"`
 Before string `json:"before_sha256"`
 Previous string `json:"previous_content"`
}
type Observation struct {
 Schema int `json:"schema"`
 State string `json:"state"`
 PayloadSHA string `json:"payload_sha256,omitempty"`
 Synthetic bool `json:"synthetic"`
}
func sum(b []byte) string { v:=sha256.Sum256(b);return hex.EncodeToString(v[:]) }
func rootCheck(root string) (string,error) {
 if root=="" {return "",errors.New("root required")}
 abs,e:=filepath.Abs(root);if e!=nil{return "",e}
 resolvedTemp,e:=filepath.EvalSymlinks(os.TempDir());if e!=nil{return "",e}
 // Only children of the system temp directory are allowed; root itself is forbidden.
 rel,e:=filepath.Rel(resolvedTemp,abs)
 if e!=nil || rel=="." || rel==".." || strings.HasPrefix(rel,".."+string(os.PathSeparator)) {return "",errors.New("root must be unique child of system temp directory")}
 if len(strings.Split(rel,string(os.PathSeparator)))!=1 {return "",errors.New("nested/ambiguous temp roots forbidden")}
 st,e:=os.Lstat(abs);if e!=nil{return "",e}
 if !st.IsDir() || st.Mode()&os.ModeSymlink!=0 {return "",errors.New("root is not a real directory")}
 b,e:=os.ReadFile(filepath.Join(abs,".rdte-owner"));if e!=nil{return "",errors.New("no synthetic ownership marker")}
 if string(b)!=marker {return "",errors.New("wrong owner marker")}
 for _,name:=range []string{"payload.txt","journal.json",".rdte-lock"} {
   st,e:=os.Lstat(filepath.Join(abs,name))
   if e==nil && (!st.Mode().IsRegular() || st.Mode()&os.ModeSymlink!=0) {return "",errors.New("symlink/special file rejected")}
   if e!=nil && !os.IsNotExist(e) {return "",e}
 }
 return abs,nil
}
func Init(root string) error {
 if root=="" {return errors.New("root required")}
 abs,e:=filepath.Abs(root);if e!=nil{return e}
 temp,e:=filepath.EvalSymlinks(os.TempDir());if e!=nil{return e}
 rel,e:=filepath.Rel(temp,abs)
 if e!=nil || rel=="." || rel==".." || strings.HasPrefix(rel,".."+string(os.PathSeparator)) || strings.ContainsRune(rel,os.PathSeparator) {return errors.New("root must be direct child of system temp")}
 if e=os.Mkdir(abs,0700);e!=nil {return fmt.Errorf("refuse preexisting root: %w",e)}
 if e=os.WriteFile(filepath.Join(abs,".rdte-owner"),[]byte(marker),0600);e!=nil {os.Remove(abs);return e}
 return nil
}
func atomic(root,name string,b []byte) error {
 f,e:=os.CreateTemp(root,".rdte-staging-*");if e!=nil{return e}
 defer os.Remove(f.Name())
 if e=f.Chmod(0600);e!=nil{f.Close();return e}
 if _,e=f.Write(b);e!=nil{f.Close();return e}
 if e=f.Sync();e!=nil{f.Close();return e}
 if e=f.Close();e!=nil{return e}
 return os.Rename(f.Name(),filepath.Join(root,name))
}
func lock(root string) (func(),error){
 f,e:=os.OpenFile(filepath.Join(root,".rdte-lock"),os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
 if e!=nil{return nil,fmt.Errorf("already held or interrupted; inspect before retry: %w",e)}
 f.Close()
 return func(){os.Remove(filepath.Join(root,".rdte-lock"))},nil
}
func Observe(root string)(Observation,error){
 abs,e:=rootCheck(root);if e!=nil{return Observation{},e}
 o:=Observation{Schema:1,State:"ABSENT",Synthetic:true}
 b,e:=os.ReadFile(filepath.Join(abs,"payload.txt"))
 if os.IsNotExist(e){return o,nil};if e!=nil{return o,e}
 o.State="PRESENT";o.PayloadSHA=sum(b);return o,nil
}
func MakePlan(root string)(Plan,error){
 abs,e:=rootCheck(root);if e!=nil{return Plan{},e}
 o,e:=Observe(abs);if e!=nil{return Plan{},e}
 p:=Plan{Schema:1,Root:abs,Before:o.PayloadSHA,After:sum([]byte(payload))}
 b,_:=json.Marshal(p);p.Digest=sum(b);return p,nil
}
func readJournal(root string)(Journal,error){
 var j Journal
 b,e:=os.ReadFile(filepath.Join(root,"journal.json"));if e!=nil{return j,e}
 e=json.Unmarshal(b,&j);return j,e
}
func Apply(root string,p Plan,approve,failpoint string)error{
 abs,e:=rootCheck(root);if e!=nil{return e}
 if approve==""||p.Digest!=approve{return errors.New("exact synthetic plan digest required")}
 unlock,e:=lock(abs);if e!=nil{return e};defer unlock()
 j,jerr:=readJournal(abs)
 if jerr==nil {
   if j.Digest!=approve{return errors.New("unrelated incomplete journal; reconcile")}
   if j.Phase=="applied" {return Verify(abs)}
   return errors.New("incomplete transaction; explicit rollback first")
 }
 if !os.IsNotExist(jerr){return jerr}
 current,e:=MakePlan(abs);if e!=nil{return e}
 if current!=p{return errors.New("stale/tampered plan")}
 before,e:=os.ReadFile(filepath.Join(abs,"payload.txt"))
 if e!=nil&&!os.IsNotExist(e){return e}
 j=Journal{Phase:"prepared",Digest:p.Digest,Before:p.Before,Previous:string(before)}
 jb,_:=json.Marshal(j);if e=atomic(abs,"journal.json",jb);e!=nil{return e}
 if failpoint=="before-write"{return errors.New("synthetic prewrite fault")}
 if e=atomic(abs,"payload.txt",[]byte(payload));e!=nil{return e}
 if failpoint=="after-write"{return errors.New("synthetic postwrite fault")}
 j.Phase="applied";jb,_=json.Marshal(j);return atomic(abs,"journal.json",jb)
}
func Verify(root string)error{
 o,e:=Observe(root);if e!=nil{return e}
 if o.State!="PRESENT"||o.PayloadSHA!=sum([]byte(payload)){return errors.New("synthetic payload hash mismatch")}
 j,e:=readJournal(root);if e!=nil{return e}
 if j.Phase!="applied" {return errors.New("not applied/verified")}
 return nil
}
func Rollback(root,approved string)error{
 abs,e:=rootCheck(root);if e!=nil{return e}
 unlock,e:=lock(abs);if e!=nil{return e};defer unlock()
 j,e:=readJournal(abs);if e!=nil{return e}
 if j.Digest!=approved{return errors.New("wrong journal digest")}
 if j.Phase=="rolled_back"{return nil}
 if j.Phase!="prepared"&&j.Phase!="applied"{return errors.New("unrecognized transaction phase")}
 o,e:=Observe(abs);if e!=nil{return e}
 if o.State=="PRESENT"&&o.PayloadSHA!=sum([]byte(payload))&&o.PayloadSHA!=j.Before{return errors.New("refuse to overwrite user modification")}
 if j.Before==""{
    if o.State=="PRESENT"{if e=os.Remove(filepath.Join(abs,"payload.txt"));e!=nil{return e}}
 }else{
    if sum([]byte(j.Previous))!=j.Before{return errors.New("backup corrupt")}
    if e=atomic(abs,"payload.txt",[]byte(j.Previous));e!=nil{return e}
 }
 j.Phase="rolled_back";b,_:=json.Marshal(j);return atomic(abs,"journal.json",b)
}
func Destroy(root,approved string)error{
 abs,e:=rootCheck(root);if e!=nil{return e}
 j,e:=readJournal(abs);if e==nil{
  if j.Digest!=approved||j.Phase!="rolled_back"{return errors.New("rollback before destroy")}
 }else if !os.IsNotExist(e){return e}
 entries,e:=os.ReadDir(abs);if e!=nil{return e}
 for _,f:=range entries{
  if f.Name()!=".rdte-owner"&&f.Name()!="journal.json" {return errors.New("foreign or unretired payload; no destroy")}
 }
 for _,f:=range entries{if e=os.Remove(filepath.Join(abs,f.Name()));e!=nil{return e}}
 return os.Remove(abs)
}
