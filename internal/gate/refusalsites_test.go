package gate

import (
	"path/filepath"
	"strings"
	"testing"
)

const refusalFixtures = "testdata/refusalsites"

func TestRefusalSitesAreReadOffTheSource(t *testing.T) {
	sites, err := refusalSitesIn(filepath.FromSlash(refusalFixtures))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, site := range sites {
		names = append(names, site.Name)
	}
	got := strings.Join(names, ",")
	if got != "first-site,second-site" {
		t.Fatalf("expected the two sites the fixture declares in source order, got %q", got)
	}
}

// A refusal a test constructs is not a refusal the validator can produce.
// Counting one would let the suite discharge its own obligation by writing the
// answer down.
func TestRefusalSitesIgnoresTestFiles(t *testing.T) {
	sites, err := refusalSitesIn(filepath.FromSlash(refusalFixtures))
	if err != nil {
		t.Fatal(err)
	}
	for _, site := range sites {
		if site.Name == "a-site-only-a-test-declares" {
			t.Fatalf("a site declared in a test file was counted: %s", site)
		}
	}
}

// A call whose first argument is not a literal names no site, so it is not one
// this accounting can hold anybody to. It is left out rather than counted
// under a name nobody wrote.
func TestRefusalSitesLeavesACallWithNoLiteralAlone(t *testing.T) {
	sites, err := refusalSitesIn(filepath.FromSlash(refusalFixtures))
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != 2 {
		t.Fatalf("expected two sites and got %d: %v", len(sites), sites)
	}
}

func TestRefusalSitesRefusesASourceTreeItCannotRead(t *testing.T) {
	if _, err := refusalSitesIn(t.TempDir()); err == nil {
		t.Fatalf("a tree with no surface in it was read as a surface with no sites")
	}
}

func TestRefusalSitesReadsACoverageProfileLine(t *testing.T) {
	file, block, count, err := parseProfileLine("github.com/iderex/messbuch/internal/validate/validate.go:270.31,273.4 2 1")
	if err != nil {
		t.Fatal(err)
	}
	if file != "github.com/iderex/messbuch/internal/validate/validate.go" {
		t.Fatalf("file read as %q", file)
	}
	if block.start != 270 || block.end != 273 || count != 1 {
		t.Fatalf("block read as %v with count %d", block, count)
	}
}

func TestRefusalSitesRefusesAProfileLineItCannotRead(t *testing.T) {
	for _, line := range []string{
		"one field",
		"a.go:1.1,2.2 2 notanumber",
		"a.go 2 1",
		"a.go:1.1 2 1",
		"a.go:one,2.2 2 1",
		"a.go:1.1,two 2 1",
	} {
		if _, _, _, err := parseProfileLine(line); err == nil {
			t.Fatalf("a profile line this leg cannot read was accepted: %q", line)
		}
	}
}

func TestRefusalSitesMatchesALineInsideAnExecutedBlock(t *testing.T) {
	covered := map[string][]coveredBlock{
		"github.com/iderex/messbuch/internal/validate/validate.go": {{start: 100, end: 110}},
	}
	inside := refusalSite{Name: "x", File: "internal/validate/validate.go", Line: 105}
	outside := refusalSite{Name: "x", File: "internal/validate/validate.go", Line: 111}
	elsewhere := refusalSite{Name: "x", File: "internal/validate/corpus.go", Line: 105}
	if !reached(covered, inside) {
		t.Fatalf("a line inside an executed block was read as unreached")
	}
	if reached(covered, outside) {
		t.Fatalf("a line outside every executed block was read as reached")
	}
	if reached(covered, elsewhere) {
		t.Fatalf("a line in another file was matched against this one's blocks")
	}
}

func TestTheGateDeclaresTheRefusalSitesLeg(t *testing.T) {
	only, err := Only(Legs(), "refusal-sites")
	if err != nil {
		t.Fatal(err)
	}
	if only[0].Run == nil {
		t.Fatalf("the leg is declared and not built")
	}
	if !strings.Contains(only[0].Limits, "EXECUTED") {
		t.Fatalf("the leg has to say that executing a line is not testing it: %q", only[0].Limits)
	}
}
