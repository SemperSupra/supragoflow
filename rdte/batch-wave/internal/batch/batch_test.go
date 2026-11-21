package batch

import (
 "encoding/json"
 "reflect"
 "strings"
 "testing"
)
func fixture()Manifest{
 sha:=strings.Repeat("a",40)
 return Manifest{
  Schema:"batch-wave/v1",Wave:"rep-01",Authority:"public-safe-synthetic-issue",
  Cards:[]Card{
   {ID:"a-host-census",Workset:"example-1",Branch:"agent/a-host-census",Commit:sha,Actor:"local-actor-a",Venue:"local-readonly",DataClass:"private",Effect:"observe",Reads:[]string{"docs/AGENTS.md"},Writes:[]string{"receipts/host-a.json"},Resources:[]string{"host/a-readonly"},Steps:[]string{"Read pinned host settings","Record sanitized census"},Oracle:"Public-only example: read-only receipt created",Rollback:"No mutation; no rollback",Receipt:"issue-1-receipt",MaxMinutes:10},
   {ID:"b-source-tests",Workset:"example-2",Branch:"agent/b-source-tests",Commit:sha,Actor:"web-agent-b",Venue:"public-gha",DataClass:"public-safe",Effect:"synthetic",Reads:[]string{"docs/AGENTS.md"},Writes:[]string{"rdte/tests/b.txt"},Resources:[]string{"runner/b"},Steps:[]string{"Run negative tests","Publish sanitized PASS/FAIL"},Oracle:"Negative test report and exact source hash",Rollback:"Discard ephemeral runner",Receipt:"issue-2-receipt",MaxMinutes:12},
  },
 }
}
func TestIndependentWave(t *testing.T){
 m:=fixture();p,e:=Validate(m);if e!=nil{t.Fatal(e)}
 if p.Count!=2||len(p.Parallel)!=1||!reflect.DeepEqual(p.Parallel[0],[]string{"a-host-census","b-source-tests"}){t.Fatalf("wrong batching %+v",p)}
 b,e:=CardBaton(m,"a-host-census");if e!=nil||b.Card.ID!="a-host-census"||b.PlanDigest!=p.Digest{t.Fatalf("bad bounded baton %+v %v",b,e)}
 b2,e:=json.Marshal(m);if e!=nil{t.Fatal(e)}
 if _,e:=ReadManifest(b2);e!=nil{t.Fatal(e)}
}
func TestCanonicalHashIndependentOfCardOrder(t *testing.T){
 m:=fixture();p,_:=Validate(m)
 m.Cards[0],m.Cards[1]=m.Cards[1],m.Cards[0]
 q,e:=Validate(m);if e!=nil{t.Fatal(e)}
 if q.Digest!=p.Digest {t.Fatal("permutation changed digest")}
}
func TestNegativeCases(t *testing.T){
 variants:=[]struct{name string;change func(*Manifest)}{
  {"duplicate-id",func(m *Manifest){m.Cards[1].ID=m.Cards[0].ID}},
  {"shared-branch",func(m *Manifest){m.Cards[1].Branch=m.Cards[0].Branch}},
  {"write-overlap",func(m *Manifest){m.Cards[1].Writes=[]string{"receipts/host-a.json"}}},
  {"read-write-race",func(m *Manifest){m.Cards[1].Writes=[]string{"docs/AGENTS.md"}}},
  {"resource-overlap",func(m *Manifest){m.Cards[1].Resources=m.Cards[0].Resources}},
  {"missing-dependency",func(m *Manifest){m.Cards[1].Depends=[]string{"ghost"}}},
  {"self-dependency",func(m *Manifest){m.Cards[1].Depends=[]string{m.Cards[1].ID}}},
  {"double-dependency",func(m *Manifest){m.Cards[1].Depends=[]string{m.Cards[0].ID,m.Cards[0].ID}}},
  {"cycle",func(m *Manifest){m.Cards[0].Depends=[]string{"b-source-tests"};m.Cards[1].Depends=[]string{"a-host-census"}}},
  {"private-cloud",func(m *Manifest){m.Cards[1].DataClass="private"}},
  {"unapproved-cloud-effect",func(m *Manifest){m.Cards[1].Effect="admitted";m.Cards[1].Admission="owner-grant"}},
  {"readonly-apply",func(m *Manifest){m.Cards[0].Effect="admitted";m.Cards[0].Admission="owner-grant"}},
  {"missing-receipt",func(m *Manifest){m.Cards[0].Receipt=""}},
  {"unpinned-commit",func(m *Manifest){m.Cards[1].Commit="main"}},
  {"oversized-step-set",func(m *Manifest){m.Cards[1].Steps=[]string{"1","2","3","4","5","6","7"}}},
  {"unbounded-job",func(m *Manifest){m.Cards[1].MaxMinutes=720}},
  {"escape-path",func(m *Manifest){m.Cards[1].Writes=[]string{"../escape"}}},
  {"unknown-venue",func(m *Manifest){m.Cards[1].Venue="unbounded-host"}},
  {"duplicate-write",func(m *Manifest){m.Cards[1].Writes=[]string{"rdte/tests/b.txt","rdte/tests/b.txt"}}},
  {"sensitive-field-in-step",func(m *Manifest){m.Cards[1].Steps=[]string{"fetch SSH key\n"} }},
 }
 for _,v:=range variants{t.Run(v.name,func(t *testing.T){m:=fixture();v.change(&m);if _,e:=Validate(m);e==nil{t.Fatal("invalid assignment accepted")}})}
}
func TestExplicitDependencyMakesSeparateWaves(t *testing.T){
 m:=fixture()
 m.Cards[1].Resources=m.Cards[0].Resources
 m.Cards[1].Depends=[]string{m.Cards[0].ID}
 p,e:=Validate(m);if e!=nil{t.Fatal(e)}
 if len(p.Parallel)!=2 || p.Parallel[0][0]!="a-host-census"||p.Parallel[1][0]!="b-source-tests" {t.Fatalf("wrong dependency %+v",p)}
}
func TestStrictJson(t *testing.T){
 m:=fixture();b,_:=json.Marshal(m)
 cases:=[]string{
  strings.Replace(string(b),`"schema":`,`"schema":"batch-wave/v1","schema":`,1),
  strings.Replace(string(b),`"actor":`,`"ssh_private_key":"forbidden","actor":`,1),
  string(b)+" {}",
  strings.Repeat(" ",MaxManifest+1),
 }
 for i,c:=range cases{if _,e:=ReadManifest([]byte(c));e==nil{t.Fatalf("malformed case %d accepted",i)}}
}
