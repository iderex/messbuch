package gate

import (
	"path/filepath"
	"strings"
	"testing"
)

func fixture(name string) string { return filepath.Join("testdata", "netimports", name) }

// The pair the whole check rests on. Two trees that differ in one thing, which
// file the network-capable import sits in, and the leg has to tell them apart.
// The accepted half is not decoration: a leg that refused everything would
// pass the refused half on its own, and this is the near miss that catches it.
func TestTheImportIsRefusedOutsideNetAndAcceptedInsideIt(t *testing.T) {
	_, err := noNetworkImportsLeg(fixture("refused"))
	if err == nil {
		t.Fatal("a package outside internal/net imported net/http and the leg passed it")
	}
	for _, want := range []string{"internal/analysis", "net/http"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "internal/net reaches") {
		t.Errorf("the refusal names the permitted package as an offender: %v", err)
	}

	examined, err := noNetworkImportsLeg(fixture("accepted"))
	if err != nil {
		t.Fatalf("the same import inside internal/net was refused: %v", err)
	}
	if !strings.Contains(examined, "package(s)") {
		t.Errorf("the leg does not say how much it examined: %q", examined)
	}
}

// Two hops. A check reading only the direct imports of each package passes
// this tree, and that is the check somebody writes first.
func TestAReachTwoHopsAwayIsRefused(t *testing.T) {
	_, err := noNetworkImportsLeg(fixture("transitive"))
	if err == nil {
		t.Fatal("a package reaching net/http through one hop was passed")
	}
	if !strings.Contains(err.Error(), "internal/analysis") {
		t.Errorf("the refusal does not name the package that reaches, only the one that imports: %v", err)
	}
}

// A test file is source. A graph computed without the test packages passes
// this tree, and a test is exactly where somebody reaches for a live server
// because it is "only a test".
func TestANetworkCapableImportInATestFileIsRefused(t *testing.T) {
	_, err := noNetworkImportsLeg(fixture("test-file"))
	if err == nil {
		t.Fatal("a test file imported net/http and the leg passed it")
	}
	if !strings.Contains(err.Error(), "internal/analysis") {
		t.Errorf("the refusal does not name the package whose test reaches: %v", err)
	}
}

// The permitted package's own dependencies are permitted, which is what
// docs/decisions/0009-offline-by-default.md says and is the branch most easily
// written the other way. A helper reached only from internal/net is not an
// offender; the leak this check exists for is a package outside that reach.
func TestAHelperReachedOnlyFromNetIsAccepted(t *testing.T) {
	examined, err := noNetworkImportsLeg(fixture("net-dependency"))
	if err != nil {
		t.Fatalf("a helper reached only from internal/net was refused: %v", err)
	}
	if examined == "" {
		t.Error("the leg passed and said nothing about what it examined")
	}
}

// Fails closed. A tree with no module is a graph that was never computed, and
// the leg has to say so rather than report a clean set it never read.
func TestAGraphThatCannotBeComputedIsARefusal(t *testing.T) {
	if _, err := noNetworkImportsLeg(t.TempDir()); err == nil {
		t.Fatal("a directory with no module passed")
	}
}

// The other half of failing closed. A module whose package set comes back
// empty is indistinguishable in the output from a tree that was never walked,
// so it is refused rather than read as a tree with nothing to check.
func TestAModuleWithNoPackagesIsARefusal(t *testing.T) {
	if _, err := noNetworkImportsLeg(fixture("no-packages")); err == nil {
		t.Fatal("a module carrying no package passed")
	}
}

// The list is data, and data that carries no reason for each entry is a list
// nobody can review. This is the property that makes adding a name a
// reviewable line rather than an edit to a matcher.
func TestEveryNetworkCapablePathCarriesItsReason(t *testing.T) {
	if len(networkCapable) == 0 {
		t.Fatal("the table is empty, so the leg refuses nothing")
	}
	seen := map[string]bool{}
	for _, entry := range networkCapable {
		if entry.Path == "" || entry.Why == "" {
			t.Errorf("entry %q carries no path or no reason", entry.Path)
		}
		if seen[entry.Path] {
			t.Errorf("%q is in the table twice", entry.Path)
		}
		seen[entry.Path] = true
		if strings.HasSuffix(entry.Path, "/") {
			t.Errorf("%q reads as a prefix, and matching is on the exact path", entry.Path)
		}
	}
	if !seen["net"] {
		t.Error("net is absent, and it is the path every socket in the standard library reaches")
	}
}

// net/url parses and opens nothing. It is named here because it is what a
// prefix rule over "net/" would have refused, and a validator reading a source
// URL out of a record is ordinary work.
func TestPathsThatOnlyLookNetworkCapableAreNotInTheTable(t *testing.T) {
	for _, entry := range networkCapable {
		if entry.Path == "net/url" || entry.Path == "net/mail" {
			t.Errorf("%q is in the table and it opens nothing", entry.Path)
		}
	}
}

// The tree this repository actually is. A leg that passes only its own
// fixtures says nothing about the mainline.
func TestThisRepositoryHasNoNetworkCapableImport(t *testing.T) {
	examined, err := noNetworkImportsLeg("../..")
	if err != nil {
		t.Fatalf("this repository reaches a network-capable API: %v", err)
	}
	if !strings.Contains(examined, "internal/net") {
		t.Errorf("the result does not name the one permitted package: %q", examined)
	}
}
