package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SemperSupra/supragoflow/internal/version"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string, out io.Writer) error {
	var showVersion bool
	var asJSON bool

	fs := flag.NewFlagSet("supragoflow", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&showVersion, "version", false, "print version info and exit")
	fs.BoolVar(&asJSON, "json", false, "print version info as JSON (use with --version)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if showVersion {
		if asJSON {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(version.Info())
		}
		_, err := fmt.Fprintf(out, "version=%s commit=%s date=%s builtBy=%s\n",
			version.Version, version.Commit, version.Date, version.BuiltBy)
		return err
	}

	_, err := fmt.Fprintln(out, "supragoflow: hello (run with --version)")
	return err
}
