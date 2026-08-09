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

	"github.com/iderex/messbuch/internal/gate"
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
	go run . help       print this

The gate prints its own legs, in the order it runs them, with what each one
examined. A leg that is not built yet prints that it was not run and what is
owed, so a run that covered less than everything cannot be read as one that
covered it and found nothing.
`)
}
