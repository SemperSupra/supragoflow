// public-workcell-rdte intentionally exercises only a synthetic temporary
// filesystem. It is NOT a Hyper-V, SSH, OCI, or host package installer.
package main

import (
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "os"
 "path/filepath"
 "strings"

 "publicworkcellrdte/internal/workcell"
)
func main(){if err:=run(os.Args[1:]);err!=nil{fmt.Fprintln(os.Stderr,"rdte:",err);os.Exit(1)}}
func run(args []string)error{
 if len(args)==0{return errors.New("usage: workcell-rdte init|observe|plan|apply|verify|rollback|destroy|exercise [--root TEMP-DIR]")}
 if args[0]=="exercise"{return exercise()}
 stage:=args[0]
 fs:=flag.NewFlagSet(stage,flag.ContinueOnError)
 root:=fs.String("root","","unique direct child of system temp directory")
 planFile:=fs.String("plan","","saved signed synthetic plan JSON")
 approve:=fs.String("approve","","exact SHA256 of synthetic plan")
 if err:=fs.Parse(args[1:]);err!=nil{return err}
 if *root==""{return errors.New("--root required")}
 switch stage {
 case "init":return workcell.Init(*root)
 case "observe":o,e:=workcell.Observe(*root);if e!=nil{return e};return printJSON(o)
 case "plan":p,e:=workcell.MakePlan(*root);if e!=nil{return e};return printJSON(p)
 case "apply":
   if *planFile=="" {return errors.New("--plan JSON file required")}
   b,e:=os.ReadFile(*planFile);if e!=nil{return e}
   var p workcell.Plan
   d:=json.NewDecoder(strings.NewReader(string(b)));d.DisallowUnknownFields()
   if e=d.Decode(&p);e!=nil{return e}
   return workcell.Apply(*root,p,*approve,"")
 case "verify":return workcell.Verify(*root)
 case "rollback":return workcell.Rollback(*root,*approve)
 case "destroy":return workcell.Destroy(*root,*approve)
 default:return fmt.Errorf("invalid stage %q",stage)
 }
}
func printJSON(v any)error{
 b,e:=json.MarshalIndent(v,"","  ");if e!=nil{return e}
 fmt.Println(string(b));return nil
}
type receipt struct {
 Schema int `json:"schema"`
 Mode string `json:"mode"`
 Stages []string `json:"stages"`
 PlanSHA string `json:"plan_sha256"`
 DestroyVerified bool `json:"destroy_verified"`
 ProviderEffects string `json:"provider_effects"`
}
func exercise()error{
 id:=make([]byte,12)
 if _,e:=rand.Read(id);e!=nil{return e}
 root:=filepath.Join(os.TempDir(),"public-rdte-exercise-"+hex.EncodeToString(id))
 if e:=workcell.Init(root);e!=nil{return e}
 // Any unexpected error leaves the isolated root as evidence rather than
 // deleting an unknown state. No privileged side effects exist.
 p,e:=workcell.MakePlan(root);if e!=nil{return e}
 if e=workcell.Apply(root,p,p.Digest,"");e!=nil{return e}
 if e=workcell.Verify(root);e!=nil{return e}
 if e=workcell.Rollback(root,p.Digest);e!=nil{return e}
 if e=workcell.Destroy(root,p.Digest);e!=nil{return e}
 if _,e=os.Stat(root);!os.IsNotExist(e){return errors.New("destroy not verified")}
 return printJSON(receipt{Schema:1,Mode:"SYNTHETIC_FILESYSTEM_ONLY",Stages:[]string{"observe","plan","apply","verify","rollback","destroy"},PlanSHA:p.Digest,DestroyVerified:true,ProviderEffects:"NONE"})
}
