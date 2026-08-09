package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A floor file that is one edit away from every fixture below.
const floorFixture = `
threshold = 85.0
reason = "Invented for a fixture."
surface = ["internal/validate"]
`

func floorAt(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "validate"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, CoverageFloorFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCoverageFloorReadsThisRepositorysOwnFile(t *testing.T) {
	floor, err := readCoverageFloor(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if floor.Threshold <= 0 || len(floor.Surface) == 0 || strings.TrimSpace(floor.Reason) == "" {
		t.Fatalf("the floor this repository ships carries no bar: %#v", floor)
	}
}

func TestCoverageFloorAcceptsTheFixture(t *testing.T) {
	if _, err := readCoverageFloor(floorAt(t, floorFixture)); err != nil {
		t.Fatal(err)
	}
}

// A setting nobody reads looks exactly like a setting somebody applied.
func TestCoverageFloorRefusesAKeyItDoesNotRead(t *testing.T) {
	_, err := readCoverageFloor(floorAt(t, floorFixture+"\nexclude = [\"internal/validate\"]\n"))
	if err == nil || !strings.Contains(err.Error(), "exclude") {
		t.Fatalf("a key nothing reads was accepted: %v", err)
	}
}

// A number with no argument behind it is the current value written down, which
// is the thing this file exists not to be.
func TestCoverageFloorRefusesAThresholdWithNoReason(t *testing.T) {
	_, err := readCoverageFloor(floorAt(t, strings.Replace(floorFixture, `reason = "Invented for a fixture."`, `reason = "  "`, 1)))
	if err == nil || !strings.Contains(err.Error(), "reason") {
		t.Fatalf("a threshold with no reason was accepted: %v", err)
	}
}

func TestCoverageFloorRefusesAThresholdThatIsNotAPercentage(t *testing.T) {
	for _, value := range []string{"0.0", "-1.0", "101.0"} {
		body := strings.Replace(floorFixture, "threshold = 85.0", "threshold = "+value, 1)
		if _, err := readCoverageFloor(floorAt(t, body)); err == nil {
			t.Fatalf("a threshold of %s was accepted", value)
		}
	}
}

func TestCoverageFloorRefusesAnEmptySurface(t *testing.T) {
	body := strings.Replace(floorFixture, `surface = ["internal/validate"]`, "surface = []", 1)
	if _, err := readCoverageFloor(floorAt(t, body)); err == nil {
		t.Fatalf("a bar over nothing was accepted, and it would pass forever")
	}
}

func TestCoverageFloorRefusesASurfaceEntryNamingNoPackage(t *testing.T) {
	body := strings.Replace(floorFixture, `surface = ["internal/validate"]`, `surface = ["internal/nowhere"]`, 1)
	_, err := readCoverageFloor(floorAt(t, body))
	if err == nil || !strings.Contains(err.Error(), "internal/nowhere") {
		t.Fatalf("a surface entry naming no package was accepted: %v", err)
	}
}

func TestCoverageFloorRefusesAMissingFile(t *testing.T) {
	if _, err := readCoverageFloor(t.TempDir()); err == nil {
		t.Fatalf("a tree with no floor file was read as a tree with a bar")
	}
}

// The half that makes the boundary mechanical: a package that can refuse a
// record and is not on the surface reds the leg, so adding a refuser somewhere
// the bar does not reach cannot happen quietly.
func TestCoverageFloorRefusesARefuserOutsideTheSurface(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "elsewhere")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := "package elsewhere\n\nfunc newRefusal(site string) string { return site }\n\nfunc a() string { return newRefusal(\"somewhere-else\") }\n"
	if err := os.WriteFile(filepath.Join(dir, "elsewhere.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	err := surfaceCoversEveryRefuser(root, []string{"internal/validate"})
	if err == nil || !strings.Contains(err.Error(), "internal/elsewhere") {
		t.Fatalf("a refuser outside the surface was accepted: %v", err)
	}
}

func TestCoverageFloorAcceptsThisRepositorysSurface(t *testing.T) {
	root := filepath.Join("..", "..")
	floor, err := readCoverageFloor(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := surfaceCoversEveryRefuser(root, floor.Surface); err != nil {
		t.Fatalf("this repository has a refuser the bar does not reach: %v", err)
	}
}

func TestCoverageFloorReadsAProfileLine(t *testing.T) {
	file, position, statements, count, err := parseProfileStatements("github.com/iderex/messbuch/internal/validate/validate.go:270.31,273.4 2 1")
	if err != nil {
		t.Fatal(err)
	}
	if file != "github.com/iderex/messbuch/internal/validate/validate.go" {
		t.Fatalf("file read as %q", file)
	}
	if position != "github.com/iderex/messbuch/internal/validate/validate.go:270.31,273.4" {
		t.Fatalf("position read as %q", position)
	}
	if statements != 2 || count != 1 {
		t.Fatalf("read %d statement(s) with count %d", statements, count)
	}
}

func TestCoverageFloorRefusesAProfileLineItCannotRead(t *testing.T) {
	for _, line := range []string{"one field", "a.go:1.1,2.2 x 1", "a.go:1.1,2.2 2 x", "a.go 2 1"} {
		if _, _, _, _, err := parseProfileStatements(line); err == nil {
			t.Fatalf("a profile line this leg cannot read was accepted: %q", line)
		}
	}
}

func TestCoverageFloorCountsOnlyTheSurface(t *testing.T) {
	surface := []string{"internal/validate"}
	if !onSurface("github.com/iderex/messbuch/internal/validate/validate.go", surface) {
		t.Fatalf("a file on the surface was read as off it")
	}
	if onSurface("github.com/iderex/messbuch/internal/gate/gate.go", surface) {
		t.Fatalf("a file off the surface was counted")
	}
	if onSurface("github.com/iderex/messbuch/internal/validate/sub/x.go", surface) {
		t.Fatalf("a file under a subdirectory of the surface package was counted as that package")
	}
}

func TestTheGateDeclaresTheCoverageFloorLeg(t *testing.T) {
	only, err := Only(Legs(), "coverage-floor")
	if err != nil {
		t.Fatal(err)
	}
	if only[0].Run == nil {
		t.Fatalf("the leg is declared and not built")
	}
	if !strings.Contains(only[0].Limits, "Executing a line is not testing it") && !strings.Contains(only[0].Limits, "executing a line is not testing it") {
		t.Fatalf("the known limit of the measure has to be in the check's own output: %q", only[0].Limits)
	}
}
