package main
import(
 "encoding/json"
 "fmt"
 "os"
 "public-ssh-lease-rdte/sshcanary"
)
func main(){
 if e:=sshcanary.Run();e!=nil{
  fmt.Fprintln(os.Stderr,"PUBLIC_LOOPBACK_SSH_CERT_CANARY_FAILED:",e)
  os.Exit(1)
 }
 b,_:=json.Marshal(struct{
  Schema string `json:"schema"`
  Verdict string `json:"verdict"`
  Scope string `json:"scope"`
  ProviderEffects string `json:"provider_effects"`
  PersistentKeys bool `json:"persistent_keys"`
 }{"public-ssh-canary/v1","PASS","loopback-only","NONE",false})
 fmt.Println(string(b))
}
