package gate

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// CoverageFloorFile is where the threshold and the surface are written.
//
// A number in a file with a written reason rather than whatever the current
// value happens to be, which is the difference between a bar and a
// description.
const CoverageFloorFile = "coverage-floor.toml"

// coverageFloor is the whole of that file. A key it does not carry is refused
// rather than skipped, for the same reason the schema loader refuses one: a
// setting nobody reads looks exactly like a setting somebody applied.
type coverageFloor struct {
	Threshold float64  `toml:"threshold"`
	Reason    string   `toml:"reason"`
	Surface   []string `toml:"surface"`
}

// coverageFloorLeg refuses a surface whose coverage has fallen below the
// stated floor.
//
// The bar is not a percentage over the repository. A number over everything
// goes up while the code that decides refusals stays untested, so the surface
// is named and the naming is guarded in the direction that drifts: a package
// holding a refusal site and missing from the list reds this leg.
//
// It fails closed. A missing or unreadable floor file, a surface entry naming
// no package, a test run that will not start, a profile that will not parse and
// a profile reporting no statement at all are all refusals rather than a bar
// that quietly measured nothing.
func coverageFloorLeg(root string) (string, error) {
	floor, err := readCoverageFloor(root)
	if err != nil {
		return "", err
	}
	if err := surfaceCoversEveryRefuser(root, floor.Surface); err != nil {
		return "", err
	}

	covered, total, err := surfaceCoverage(root, floor.Surface)
	if err != nil {
		return "", err
	}
	percent := 100 * float64(covered) / float64(total)
	if percent < floor.Threshold {
		return "", fmt.Errorf("%.1f%% of %d statement(s) on the surface are executed, and %s requires %.1f%%.\nThe surface is %s.\nRaising the bar's own number to make this green is the move that file exists to make visible.",
			percent, total, CoverageFloorFile, floor.Threshold, strings.Join(floor.Surface, ", "))
	}
	return fmt.Sprintf("%.1f%% of %d statement(s) over %s, against a floor of %.1f%% in %s",
		percent, total, strings.Join(floor.Surface, ", "), floor.Threshold, CoverageFloorFile), nil
}

// limitsOfCoverageFloor is printed beside the leg's result, because the known
// limit of the measure belongs in the check's own output and not only in a
// document.
const limitsOfCoverageFloor = `executing a line is not testing it. A suite that reaches every statement on
this surface while asserting nothing scores a hundred per cent here, which is
why #51 sits beside this leg and not instead of it. The refusal sites carry a
stricter obligation of their own under the refusal-sites leg, and this number
covers what is left rather than the part that matters most. The surface is the
list in coverage-floor.toml: a package holding a refusal site and missing from
it reds this leg, and a package producing numbers rather than refusals is
outside what that guard can find.`

// readCoverageFloor reads the file and refuses a floor that could not do its
// job.
func readCoverageFloor(root string) (*coverageFloor, error) {
	rel := CoverageFloorFile
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, so there is no bar to measure against: %w", rel, err)
	}
	var floor coverageFloor
	md, err := toml.Decode(string(b), &floor)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rel, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		var keys []string
		for _, k := range undecoded {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("%s: %d key(s) this leg does not read: %s", rel, len(keys), strings.Join(keys, ", "))
	}
	if floor.Threshold <= 0 || floor.Threshold > 100 {
		return nil, fmt.Errorf("%s: a threshold of %v is not a percentage, and a bar nothing can fall below is not a bar", rel, floor.Threshold)
	}
	if strings.TrimSpace(floor.Reason) == "" {
		return nil, fmt.Errorf("%s: the threshold carries no reason, and a number with no argument behind it is the current value written down", rel)
	}
	if len(floor.Surface) == 0 {
		return nil, fmt.Errorf("%s: the surface is empty, so this leg would measure nothing and pass", rel)
	}
	for _, pkg := range floor.Surface {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(pkg))); err != nil {
			return nil, fmt.Errorf("%s names %s and there is no such package in this tree, so the bar covers less than it says: %w", rel, pkg, err)
		}
	}
	return &floor, nil
}

// surfaceCoversEveryRefuser refuses a package that can refuse a record and is
// not on the surface.
//
// This is the half that makes the boundary mechanical. The list itself is
// written by a person, and a list a person maintains drifts the first time
// somebody adds a file; what cannot happen quietly is adding a refusal
// somewhere the bar does not reach.
func surfaceCoversEveryRefuser(root string, surface []string) error {
	listed := map[string]bool{}
	for _, pkg := range surface {
		listed[path.Clean(filepath.ToSlash(pkg))] = true
	}

	var missing []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		slashed := filepath.ToSlash(rel)
		if strings.Contains(slashed, "/testdata") {
			return filepath.SkipDir
		}
		if listed[slashed] {
			return nil
		}
		sites, err := refusalSitesInDir(p, root)
		if err != nil {
			return err
		}
		if len(sites) > 0 {
			missing = append(missing, fmt.Sprintf("%s (%d site(s))", slashed, len(sites)))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot walk internal/, so which packages can refuse is unknown: %w", err)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("%d package(s) can refuse a record and are not on the surface in %s:\n  %s\n\nA bar that does not reach the code deciding refusals is a number going up while the thing that matters rots.",
			len(missing), CoverageFloorFile, strings.Join(missing, "\n  "))
	}
	return nil
}

// surfaceCoverage runs the surface's own suites with coverage attributed to
// the surface, and returns the executed and total statement counts.
//
// The attribution is what makes the number mean anything here. Without it a
// package is credited only for what its own tests reach, and a reader of one
// package exercised entirely through another's fixtures scores near zero while
// being thoroughly executed.
func surfaceCoverage(root string, surface []string) (covered, total int, err error) {
	module, err := modulePath(root)
	if err != nil {
		return 0, 0, err
	}

	var pkgs, imports []string
	for _, pkg := range surface {
		pkgs = append(pkgs, "./"+filepath.ToSlash(pkg))
		imports = append(imports, module+"/"+filepath.ToSlash(pkg))
	}

	dir, err := os.MkdirTemp("", "messbuch-floor")
	if err != nil {
		return 0, 0, fmt.Errorf("cannot make a place for the coverage profile: %w", err)
	}
	defer os.RemoveAll(dir)
	profile := filepath.Join(dir, "cover.out")

	args := append([]string{"test", "-count=1", "-covermode=set",
		"-coverpkg=" + strings.Join(imports, ","), "-coverprofile=" + profile}, pkgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	if out, cmdErr := cmd.CombinedOutput(); cmdErr != nil {
		return 0, 0, fmt.Errorf("the suites over the surface did not pass, so no coverage number here means anything:\n%s", strings.TrimSpace(string(out)))
	}

	f, err := os.Open(profile)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read the coverage profile, so the number is unknown: %w", err)
	}
	defer f.Close()

	// One block appears once per test binary the profile was merged from, so
	// the blocks are folded by position before anything is counted. Summing the
	// lines as they arrive would count every statement as many times as there
	// are packages under test and read a block executed by one binary and not
	// the other as half covered.
	blocks := map[string]int{}
	statementsOf := map[string]int{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		name, position, statements, count, parseErr := parseProfileStatements(line)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("cannot read the coverage profile: %w", parseErr)
		}
		if !onSurface(name, surface) {
			continue
		}
		statementsOf[position] = statements
		if count > blocks[position] {
			blocks[position] = count
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return 0, 0, fmt.Errorf("cannot read the coverage profile: %w", scanErr)
	}
	for position, statements := range statementsOf {
		total += statements
		if blocks[position] > 0 {
			covered += statements
		}
	}
	if total == 0 {
		return 0, 0, fmt.Errorf("the coverage profile reports no statement on the surface at all, which is a broken measurement rather than a covered one")
	}
	return covered, total, nil
}

// onSurface reports whether a file named by a coverage profile sits in one of
// the surface's packages.
func onSurface(name string, surface []string) bool {
	dir := path.Dir(name)
	for _, pkg := range surface {
		if strings.HasSuffix(dir, "/"+path.Clean(filepath.ToSlash(pkg))) {
			return true
		}
	}
	return false
}

// parseProfileStatements reads the statement count and the execution count out
// of one profile line.
func parseProfileStatements(line string) (file, position string, statements, count int, err error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", "", 0, 0, fmt.Errorf("a line with %d fields rather than 3: %q", len(fields), line)
	}
	if statements, err = strconv.Atoi(fields[1]); err != nil {
		return "", "", 0, 0, fmt.Errorf("a statement count that is not a number in %q", line)
	}
	if count, err = strconv.Atoi(fields[2]); err != nil {
		return "", "", 0, 0, fmt.Errorf("an execution count that is not a number in %q", line)
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon < 0 {
		return "", "", 0, 0, fmt.Errorf("no position in %q", line)
	}
	return fields[0][:colon], fields[0], statements, count, nil
}
