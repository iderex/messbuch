// Package gate holds the local gate: the legs it runs, in order, and the
// runner that stops at the first failure.
//
// The runner exists so that the answer to "what did this run check" comes out
// of the run itself. Two properties are what the rest of the milestone is
// allowed to rely on. A leg that fails ends the run, so nothing after a
// failure reports a result nobody looked at. And a leg that is not built yet
// prints that it was not run, together with what is owed and where, so a run
// covering four of six legs cannot be mistaken for one covering six.
package gate

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// A Leg is one step of the gate.
//
// Exactly one of Run and Owed is set. Owed set means the leg is declared but
// not built: the runner prints it as not run, with the text of Owed as the
// reason, and the run continues. Run set means the leg examines something, and
// what it examined is the string it returns.
type Leg struct {
	// ID is the short name the run prints and a failure names.
	ID string

	// Subject is one line saying what the leg looks at.
	Subject string

	// Owed, when not empty, says why the leg is not built and who owes it.
	Owed string

	// Run examines the tree rooted at root. Its first return is what it
	// examined, phrased so a reader can tell a leg that looked at nothing
	// from one that looked at something and was content.
	Run func(root string) (examined string, err error)
}

// ErrLegFailed wraps the error a leg returned, so a caller can tell a failing
// leg from the runner itself failing.
type ErrLegFailed struct {
	ID  string
	Err error
}

func (e *ErrLegFailed) Error() string {
	return fmt.Sprintf("leg %s failed: %v", e.ID, e.Err)
}

func (e *ErrLegFailed) Unwrap() error { return e.Err }

// Run executes legs in order against root, writing a line per leg to w. It
// returns at the first failing leg without running the rest.
//
// A leg carrying neither Run nor Owed is a programming mistake in this package
// rather than a finding about the tree, and it is refused before any leg runs
// so that the refusal cannot be read as a verdict on the repository.
func Run(w io.Writer, legs []Leg, root string) error {
	if len(legs) == 0 {
		return errors.New("the gate declares no legs, so this run examined nothing")
	}
	for _, leg := range legs {
		if leg.Run == nil && leg.Owed == "" {
			return fmt.Errorf("leg %s declares neither an implementation nor what is owed", leg.ID)
		}
		if leg.Run != nil && leg.Owed != "" {
			return fmt.Errorf("leg %s is both built and owed", leg.ID)
		}
	}

	var ran, notRun int
	for _, leg := range legs {
		if leg.Owed != "" {
			notRun++
			fmt.Fprintf(w, "not run  %-18s %s\n", leg.ID, leg.Subject)
			for _, line := range strings.Split(leg.Owed, "\n") {
				fmt.Fprintf(w, "         %-18s %s\n", "", line)
			}
			continue
		}

		examined, err := leg.Run(root)
		if err != nil {
			fmt.Fprintf(w, "FAILED   %-18s %s\n", leg.ID, leg.Subject)
			fmt.Fprintf(w, "\n%d leg(s) ran, %d not run, and the run stopped at %s.\n", ran, notRun, leg.ID)
			return &ErrLegFailed{ID: leg.ID, Err: err}
		}
		ran++
		fmt.Fprintf(w, "ok       %-18s %s\n", leg.ID, examined)
	}

	fmt.Fprintf(w, "\n%d leg(s) ran and %d were not run. A leg that was not run examined nothing.\n", ran, notRun)
	return nil
}
