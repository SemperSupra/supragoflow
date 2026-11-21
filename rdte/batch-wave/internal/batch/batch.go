// Package batch statically validates small, conflict-free GitHub DLE work waves.
// It does not dispatch work, authenticate a caller, or authorize any effect.
package batch

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "path"
 "regexp"
 "sort"
 "strings"
)

const MaxManifest = 65536
var idRE=regexp.MustCompile("^[a-z][a-z0-9-]{1,39}$")
var shaRE=regexp.MustCompile("^[a-f0-9]{40}$")

type Manifest struct {
 Schema string `json:"schema"`
 Wave string `json:"wave"`
 Authority string `json:"authority"`
 Cards []Card `json:"cards"`
}
type Card struct {
 ID string `json:"id"`
 Workset string `json:"workset"`
 Branch string `json:"branch"`
 Commit string `json:"commit"`
 Actor string `json:"actor"`
 Venue string `json:"venue"`
 DataClass string `json:"data_class"`
 Effect string `json:"effect"`
 Admission string `json:"admission,omitempty"`
 Depends []string `json:"depends"`
 Reads []string `json:"reads"`
 Writes []string `json:"writes"`
 Resources []string `json:"resources"`
 Steps []string `json:"steps"`
 Oracle string `json:"oracle"`
 Rollback string `json:"rollback"`
 Receipt string `json:"receipt"`
 MaxMinutes int `json:"max_minutes"`
}
type Plan struct {
 Schema string `json:"schema"`
 Wave string `json:"wave"`
 Digest string `json:"digest_sha256"`
 Parallel [][]string `json:"parallel_waves"`
 Count int `json:"card_count"`
}
type Baton struct {
 Schema string `json:"schema"`
 Wave string `json:"wave"`
 Authority string `json:"authority"`
 PlanDigest string `json:"plan_digest_sha256"`
 Card Card `json:"card"`
}

func valid(s string, limit int) bool {
 return len(s)>0 && len(s)<=limit && s==strings.TrimSpace(s) && !strings.ContainsAny(s,"\r\n\x00")
}
func validPath(s string) bool {
 if !valid(s,200)|| strings.Contains(s,"\\") || strings.HasPrefix(s,"/") || strings.Contains(s,":") ||
    strings.HasPrefix(s,".") || path.Clean(s)!=s || s=="." {return false}
 parts:=strings.Split(s,"/")
 for _,p:=range parts {if p=="."||p==".."||p==""{return false}}
 return true
}
func uniqueStrings(items []string, check func(string)bool) error{
 seen:=map[string]bool{}
 for _,x:=range items{
  if !check(x){return fmt.Errorf("bad scoped value")}
  if seen[x]{return errors.New("duplicate scoped value")}
  seen[x]=true
 }
 return nil
}
func pathsCollide(a,b string) bool {
 return a==b||strings.HasPrefix(a,b+"/")||strings.HasPrefix(b,a+"/")
}
func listIntersect(a,b []string,folder bool) bool {
 for _,x:=range a {for _,y:=range b{
  if (folder&&pathsCollide(x,y))||(!folder&&x==y){return true}
 }}
 return false
}
func conflicts(a,b Card)bool {
 if listIntersect(a.Writes,b.Writes,true)||listIntersect(a.Writes,b.Reads,true)||listIntersect(a.Reads,b.Writes,true) {return true}
 if listIntersect(a.Resources,b.Resources,false){return true}
 return false
}
func reaches(from,to string, cards map[string]Card, visited map[string]bool)bool{
 if from==to{return true}
 if visited[from]{return false}
 visited[from]=true
 c,ok:=cards[from];if !ok{return false}
 for _,d:=range c.Depends {if reaches(d,to,cards,visited){return true}}
 return false
}
func canonical(m Manifest)Manifest{
 sort.Slice(m.Cards,func(i,j int)bool{return m.Cards[i].ID<m.Cards[j].ID})
 for i:=range m.Cards {
  c:=&m.Cards[i]
  sort.Strings(c.Depends);sort.Strings(c.Reads);sort.Strings(c.Writes);sort.Strings(c.Resources)
 }
 return m
}
func Validate(m Manifest)(Plan,error){
 empty:=Plan{}
 if m.Schema!="batch-wave/v1"||!idRE.MatchString(m.Wave)||!valid(m.Authority,200)||
   len(m.Cards)<1||len(m.Cards)>8 {return empty,errors.New("invalid wave envelope or size")}
 cards:=map[string]Card{}; branchOwner:=map[string]string{}
 for _,c:=range m.Cards {
  if !idRE.MatchString(c.ID)||!valid(c.Workset,200)||!valid(c.Branch,160)||!shaRE.MatchString(c.Commit)||
   !valid(c.Actor,120)||!valid(c.Oracle,220)||!valid(c.Rollback,220)||!valid(c.Receipt,200)||
   c.MaxMinutes<1||c.MaxMinutes>30||len(c.Steps)==0||len(c.Steps)>6 {
    return empty,fmt.Errorf("invalid or oversized card %q",c.ID)
  }
  if _,ok:=cards[c.ID];ok{return empty,fmt.Errorf("duplicate card ID %s",c.ID)}
  if prior,ok:=branchOwner[c.Branch];ok{return empty,fmt.Errorf("shared mutable branch %s / %s",prior,c.ID)}
  branchOwner[c.Branch]=c.ID
  if c.DataClass!="public-safe"&&c.DataClass!="private" {return empty,fmt.Errorf("%s invalid data class",c.ID)}
  if c.Venue!="public-gha"&&c.Venue!="local-readonly"&&c.Venue!="protected-disposable" {
   return empty,fmt.Errorf("%s unknown venue",c.ID)
  }
  if c.Effect!="source"&&c.Effect!="observe"&&c.Effect!="synthetic"&&c.Effect!="admitted" {
   return empty,fmt.Errorf("%s unknown effect",c.ID)
  }
  if c.Effect=="admitted"{
   if c.Venue!="protected-disposable"||!valid(c.Admission,200)||strings.Contains(strings.ToLower(c.Admission),"pending"){
    return empty,fmt.Errorf("%s effect not independently admitted",c.ID)
   }
  }else if c.Admission!=""{return empty,fmt.Errorf("%s needless admission identifier",c.ID)}
  if c.Venue=="public-gha"&&(c.DataClass!="public-safe"||(c.Effect!="source"&&c.Effect!="synthetic")){
   return empty,fmt.Errorf("%s private/effectful data forbidden on public GHA",c.ID)
  }
  if c.Venue=="local-readonly"&&(c.Effect!="source"&&c.Effect!="observe"){
   return empty,fmt.Errorf("%s local discovery must be read-only",c.ID)
  }
  for _,s:=range c.Steps{if !valid(s,180){return empty,fmt.Errorf("%s step too long or invalid",c.ID)}}
  if e:=uniqueStrings(c.Depends,func(x string)bool{return idRE.MatchString(x)&&x!=c.ID});e!=nil{return empty,fmt.Errorf("%s dependency: %w",c.ID,e)}
  for _,item:=range []struct{n string; a []string;f func(string)bool}{
   {"reads",c.Reads,validPath},{"writes",c.Writes,validPath},{"resources",c.Resources,func(s string)bool{return validPath(s)}},
  }{
   if e:=uniqueStrings(item.a,item.f);e!=nil{return empty,fmt.Errorf("%s %s: %w",c.ID,item.n,e)}
  }
  cards[c.ID]=c
 }
 for _,c:=range m.Cards {
  for _,dep:=range c.Depends{if _,ok:=cards[dep];!ok{return empty,fmt.Errorf("%s unknown dependency %s",c.ID,dep)}}
 }
 for i:=0;i<len(m.Cards);i++ {for j:=i+1;j<len(m.Cards);j++ {
  a,b:=m.Cards[i],m.Cards[j]
  if conflicts(a,b)&&!reaches(a.ID,b.ID,cards,map[string]bool{})&&!reaches(b.ID,a.ID,cards,map[string]bool{}){
   return empty,fmt.Errorf("%s and %s have undeclared read/write or resource conflict",a.ID,b.ID)
  }
 }}
 left:=map[string]bool{};for id:=range cards{left[id]=true}
 waves:=[][]string{}
 for len(left)>0 {
  ready:=[]string{}
  for id:=range left{
   yes:=true
   for _,dep:=range cards[id].Depends {if left[dep]{yes=false;break}}
   if yes{ready=append(ready,id)}
  }
  if len(ready)==0{return empty,errors.New("dependency cycle")}
  sort.Strings(ready)
  waves=append(waves,ready)
  for _,id:=range ready{delete(left,id)}
 }
 m=canonical(m)
 serialized,e:=json.Marshal(m);if e!=nil{return empty,e}
 h:=sha256.Sum256(serialized)
 return Plan{Schema:m.Schema,Wave:m.Wave,Digest:hex.EncodeToString(h[:]),Parallel:waves,Count:len(m.Cards)},nil
}
// ReadManifest enforces strict JSON shape and duplicate-key rejection. It does
// not read a credential store or trust an actor-supplied effect admission.
func ReadManifest(b []byte)(Manifest,error){
 var m Manifest
 if len(b)==0||len(b)>MaxManifest{return m,errors.New("manifest size invalid")}
 check:=json.NewDecoder(strings.NewReader(string(b)))
 if e:=rejectDuplicateKeys(check);e!=nil{return m,e}
 if _,e:=check.Token();e!=io.EOF{return m,errors.New("trailing JSON document or token")}
 dec:=json.NewDecoder(strings.NewReader(string(b)))
 dec.DisallowUnknownFields()
 if e:=dec.Decode(&m);e!=nil{return m,e}
 if _,e:=dec.Token();e!=io.EOF{return m,errors.New("trailing JSON document")}
 _,e:=Validate(m);return m,e
}
func rejectDuplicateKeys(d *json.Decoder) error{
 tok,e:=d.Token();if e!=nil{return e}
 delim,ok:=tok.(json.Delim);if !ok{return nil}
 switch delim {
 case '{':
  keys:=map[string]bool{}
  for d.More(){
   k,e:=d.Token();if e!=nil{return e}
   key,ok:=k.(string);if !ok{return errors.New("invalid object key")}
   if keys[key]{return errors.New("duplicate JSON key")}
   keys[key]=true
   if e=rejectDuplicateKeys(d);e!=nil{return e}
  }
  _,e=d.Token();return e
 case '[':
  for d.More(){if e=rejectDuplicateKeys(d);e!=nil{return e}}
  _,e=d.Token();return e
 default:
  return errors.New("unexpected JSON delimiter")
 }
}
func CardBaton(m Manifest,id string)(Baton,error){
 p,e:=Validate(m);if e!=nil{return Baton{},e}
 for _,c:=range m.Cards{
  if c.ID==id{
   return Baton{Schema:m.Schema,Wave:m.Wave,Authority:m.Authority,PlanDigest:p.Digest,Card:c},nil
  }
 }
 return Baton{},errors.New("card ID not found")
}
