// The one entry point of this repository.
//
//	go run . ci
//
// runs the whole local gate. Every workflow on the scaffolding milestone wraps
// this same command rather than restating its steps, so a green run here and a
// green run on a pull request are statements about one procedure and not two.
//
// The legs are not listed in this comment, in the readme or in any other
// document. The command prints them, together with what each one examined, and
// a document that repeated the list would drift against the code that decides
// it.
package main

import (
	"fmt"
	"os"

	"github.com/iderex/messbuch/internal/build"
	"github.com/iderex/messbuch/internal/compat"
	"github.com/iderex/messbuch/internal/gate"
	"github.com/iderex/messbuch/internal/schema"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(os.Stderr)
		os.Exit(2)
	}

	switch args[0] {
	case "ci":
		if len(args) > 2 {
			usage(os.Stderr)
			os.Exit(2)
		}
		legs := gate.Legs()
		if len(args) == 2 {
			var err error
			if legs, err = gate.Only(legs, args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}
		}
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the working directory: %v\n", err)
			os.Exit(1)
		}
		if err := gate.Run(os.Stdout, legs, wd); err != nil {
			fmt.Fprintf(os.Stderr, "\n%v\n", err)
			os.Exit(1)
		}
	case "fmt":
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the working directory: %v\n", err)
			os.Exit(1)
		}
		changed, err := gate.Reformat(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if len(changed) == 0 {
			fmt.Println("nothing to change; every file already carries the bytes the gate demands")
			return
		}
		for _, name := range changed {
			fmt.Println(name)
		}
		fmt.Printf("\n%d file(s) rewritten.\n", len(changed))
	case "schema":
		// The schema file is written for a program to read. This prints the
		// same authority as sentences, so that a contributor transcribing a
		// measurement never has to open TOML or Go to find out what a field
		// may hold. It renders the loaded schema and restates nothing, so it
		// cannot say something the validator does not obey.
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the working directory: %v\n", err)
			os.Exit(1)
		}
		set, err := schema.Load(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if err := schema.Print(os.Stdout, set); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	case "build":
		// The corpus as one thing a consumer can load, produced from the
		// tracked records every time rather than kept beside them. What it
		// writes is untracked, and the gate's artifact-untracked leg refuses a
		// build output that was committed; internal/build is where that choice
		// is argued.
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the working directory: %v\n", err)
			os.Exit(1)
		}
		set, err := schema.Load(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		prov, err := build.GitProvenance(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		artifact, err := build.Build(wd, set, prov)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		written, err := build.WriteAll(wd, artifact, set.Any())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		for _, name := range written {
			fmt.Println(name)
		}
		fmt.Printf("\n%d record(s) at revision %s, corpus version %s, tree %s.\n",
			artifact.Stamp.SelectedCount, artifact.Stamp.CorpusRevision,
			artifact.Stamp.CorpusVersion, artifact.Stamp.CorpusState)
		fmt.Printf("%s is the authority. %s drops %d field(s) and names them in its own header.\n",
			build.JSONName, build.CSVName, len(build.Dropped(set.Any())))
	case "freeze":
		// How a schema version is frozen, and the only supported way to write
		// a readings file. Hand-writing one would put a person in the position
		// of deciding what this tree reads, which is the thing the frozen
		// reading exists to observe rather than to assert.
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read the working directory: %v\n", err)
			os.Exit(1)
		}
		written, err := compat.FreezeAll(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		for _, name := range written {
			fmt.Println(name)
		}
		fmt.Printf("\n%d readings file(s) written. What changed in them is the change in interpretation, and it belongs in the pull request body.\n", len(written))
	case "refusals":
		// The list lives in the code that produces it and is printed rather
		// than written down, because a list of refusals in a document drifts
		// against the validator and the drift is invisible: a reader trusts
		// the document exactly where it has stopped being true.
		for _, site := range gate.Sites() {
			fmt.Printf("%-26s %s\n", site.ID, site.Refuses)
		}
	case "help", "-h", "--help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "unknown verb %q\n\n", args[0])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w *os.File) {
	fmt.Fprint(w, `messbuch

	go run . ci         run the local gate
	go run . ci <leg>   run one leg of it, named by the id the run prints
	go run . fmt        write the bytes the format-and-lint leg demands
	go run . schema     print the record schema as sentences
	go run . build      write the corpus artifact under build/
	go run . freeze     rewrite the frozen readings of every frozen schema version
	go run . refusals   print every refusal the corpus validator can produce
	go run . help       print this

The gate prints its own legs, in the order it runs them, with what each one
examined. A leg that is not built yet prints that it was not run and what is
owed, so a run that covered less than everything cannot be read as one that
covered it and found nothing.
`)
}
