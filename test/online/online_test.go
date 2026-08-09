//go:build online

package online

import (
	"fmt"
	"testing"
)

// An online test joins this register rather than being written as a plain Test
// function, so that the harness can say how many it ran.
//
// A count is the whole point. Under an ordinary test file, a harness with
// nothing in it prints the same line as a harness whose tests all passed, and
// the person reading the output has no way to tell them apart. That silence is
// what docs/decisions/0010-headless-tests.md refuses.
type onlineTest struct {
	Name string
	Run  func(t *testing.T)
}

// onlineTests is empty today. Nothing in this project reaches the outside world
// yet: the component permitted to do so does not exist, and the first entries
// here will be the ones that resolve a source identifier against a
// bibliographic service.
var onlineTests []onlineTest

// TestTheHarnessSaysWhatItRan runs the register and reports its size, zero
// included.
//
// It never asserts that the number is greater than zero. A harness that refused
// to be empty would have to carry a test that exists to keep it non-empty,
// which is a test about the harness rather than about the outside world.
// Reporting the number honestly is what this is for.
func TestTheHarnessSaysWhatItRan(t *testing.T) {
	for _, one := range onlineTests {
		t.Run(one.Name, func(t *testing.T) { one.Run(t) })
	}
	fmt.Printf("online harness: %d test(s) registered and run.\n", len(onlineTests))
	if len(onlineTests) == 0 {
		fmt.Println("Zero. This run reached the outside world for nothing, and it is not a pass for anything.")
	}
}
