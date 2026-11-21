// batch-wave is a *static* manifest validator and bounded microbaton renderer.
// It cannot execute jobs, create cloud resources or validate actual permissions.
package main

import (
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "os"
 "publicbatchwave/internal/batch"
)
func main(){
 if e:=run(os.Args[1:]);e!=nil{fmt.Fprintln(os.Stderr,"batch-wave:",e);os.Exit(2)}
}
func run(args []string)error{
 if len(args)<2{return errors.New("usage: batch-wave plan FILE | baton FILE CARD_ID")}
 cmd,file:=args[0],args[1]
 if cmd!="plan"&&cmd!="baton"{return errors.New("unknown command")}
 if cmd=="baton"&&len(args)!=3{return errors.New("baton requires exactly one card id")}
 if cmd=="plan"&&len(args)!=2{return errors.New("plan requires one file")}
 f,e:=os.Open(file);if e!=nil{return e};defer f.Close()
 b,e:=io.ReadAll(io.LimitReader(f,batch.MaxManifest+1));if e!=nil{return e}
 m,e:=batch.ReadManifest(b);if e!=nil{return e}
 var v any
 if cmd=="plan"{v,e=batch.Validate(m)}else{v,e=batch.CardBaton(m,args[2])}
 if e!=nil{return e}
 encoder:=json.NewEncoder(os.Stdout);encoder.SetIndent("","  ")
 return encoder.Encode(v)
}
