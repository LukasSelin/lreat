// Command owl renders lreat's ontology as an OWL 2 document in functional-style
// syntax. With no flags it writes lreat.ofn beside this file, which is the
// copy checked in; -o - prints to standard output instead, and -lint reports
// what gowl thinks of the result.
//
//	go run ./owl            # regenerate owl/lreat.ofn
//	go run ./owl -o - | less
//	go run ./owl -lint
//
// The document is derived, not authored: everything in it is read off the Go
// declarations in core/ontology, so the way to change it is to change those
// and run this again.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gowl/lint"
	"gowl/owl"
)

func main() {
	out := flag.String("o", "", "where to write the document; - for standard output")
	check := flag.Bool("lint", false, "report lint findings instead of writing")
	flag.Parse()

	o, err := Build()
	if err != nil {
		fail(err)
	}
	o.Sort()

	if *check {
		// The roots are orphans by construction — :Act and :Thing are under
		// nothing — and gowl says so at info. Only a warning or worse is a
		// reason to fail.
		var bad int
		for _, f := range lint.Run(o, lint.Default()) {
			fmt.Println(f)
			if f.Severity >= lint.Warning {
				bad++
			}
		}
		fmt.Fprintf(os.Stderr, "%d axioms; profiles: %v\n", o.Len(), owl.Profiles(o))
		if bad > 0 {
			os.Exit(1)
		}
		return
	}

	if *out == "-" {
		if err := o.WriteFunctional(os.Stdout); err != nil {
			fail(err)
		}
		return
	}
	path := *out
	if path == "" {
		path = defaultPath()
	}
	f, err := os.Create(path)
	if err != nil {
		fail(err)
	}
	if err := o.WriteFunctional(f); err != nil {
		f.Close()
		fail(err)
	}
	if err := f.Close(); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s: %d axioms, %d classes\n", path, o.Len(), len(o.Classes()))
}

// defaultPath is lreat.ofn beside this command, wherever it was run from.
func defaultPath() string {
	if wd, err := os.Getwd(); err == nil {
		if filepath.Base(wd) == "owl" {
			return "lreat.ofn"
		}
		return filepath.Join("owl", "lreat.ofn")
	}
	return "lreat.ofn"
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "owl:", err)
	os.Exit(1)
}
