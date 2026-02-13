package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/SemperSupra/supragoflow/internal/version"
)

func main() {
	var showVersion bool
	var asJSON bool

	flag.BoolVar(&showVersion, "version", false, "print version info and exit")
	flag.BoolVar(&asJSON, "json", false, "print version info as JSON (use with --version)")
	flag.Parse()

	if showVersion {
		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(version.Info())
			return
		}
		fmt.Printf("version=%s commit=%s date=%s builtBy=%s\n",
			version.Version, version.Commit, version.Date, version.BuiltBy)
		return
	}

	fmt.Println("supragoflow: hello (run with --version)")
}
