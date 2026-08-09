package gate

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// refusalSurface is the package whose refusals are accounted for.
//
// One path rather than a list, because there is one place that refuses a
// record today. A second surface is a second entry here and a line in the
// leg's own output, not a silent widening.
const refusalSurface = "internal/validate"

// refusalConstructors are the function names every refusal in that package is
// made through. A refusal built any other way is invisible to this accounting,
// which is why the package has exactly one constructor and the checker's
// helper calls it.
var refusalConstructors = []string{"newRefusal", "refuse"}

// A refusalSite is one place in the source that can refuse.
//
// The accounting is per site and not per refusal name, and the difference is
// the whole point. A second branch inside an existing refusal adds no name to
// the catalogue, so a check counting names would stay green while the new
// branch had never been executed by anything.
type refusalSite struct {
	Name string
	File string
	Line int
}

func (s refusalSite) String() string {
	return fmt.Sprintf("%s:%d refuses %s", s.File, s.Line, s.Name)
}

// refusalSitesLeg refuses a refusal site that no fixture reaches.
//
// The site list is derived from the source and the reached set is derived from
// running the fixtures under coverage. Neither comes from a list somebody
// maintains, because a list maintained by hand drifts the first time a branch
// is added, and the drift is invisible: an unproven refusal looks exactly like
// a proven one from outside.
//
// It fails closed. A source tree it cannot parse, a test run it cannot start,
// a coverage profile it cannot read and a surface with no site in it at all
// are refusals rather than an empty set read as a complete one.
func refusalSitesLeg(root string) (string, error) {
	sites, err := refusalSitesIn(root)
	if err != nil {
		return "", err
	}
	if len(sites) == 0 {
		return "", fmt.Errorf("no refusal site under %s, so this leg examined nothing; a surface that refuses nothing is a finding rather than a clean result", refusalSurface)
	}

	covered, err := coveredLines(root)
	if err != nil {
		return "", err
	}

	var unreached []string
	for _, site := range sites {
		if !reached(covered, site) {
			unreached = append(unreached, site.String())
		}
	}
	if len(unreached) > 0 {
		sort.Strings(unreached)
		return "", fmt.Errorf("%d of %d refusal site(s) are executed by no fixture:\n  %s\n\nA refusal nobody has watched fire is a refusal nobody knows is wired up. Reach it with a fixture, or take the branch out.",
			len(unreached), len(sites), strings.Join(unreached, "\n  "))
	}
	return fmt.Sprintf("%d refusal site(s) under %s, every one executed by a fixture", len(sites), refusalSurface), nil
}

// limitsOfRefusalSites is printed beside the leg's result.
const limitsOfRefusalSites = `this says a refusal site was EXECUTED by a fixture, and never that the fixture
asserted anything about what came out of it. Executing a line is not testing
it, and a suite that reaches every site while asserting nothing scores
perfectly here. What the fixtures assert is the suite's own business and #51 is
where the quality of the assertions is measured. The surface is the one package
named in internal/gate/refusalsites.go; a refusal that grows somewhere else is
not accounted for until that package is named too.`

// refusalSitesIn reads the package source and returns every call to a refusal
// constructor whose first argument is a string literal.
//
// Test files are excluded. A refusal a test constructs is not a refusal the
// validator can produce, and counting one would let the suite discharge an
// obligation by writing the answer down.
func refusalSitesIn(root string) ([]refusalSite, error) {
	return refusalSitesInDir(filepath.Join(root, filepath.FromSlash(refusalSurface)), root)
}

// refusalSitesInDir is the same reading over one directory, so that the
// coverage floor can ask the same question of every package under internal/
// without a second parser that agrees until it does not.
func refusalSitesInDir(dir, root string) ([]refusalSite, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s, so the refusal sites are unknown: %w", filepath.ToSlash(dir), err)
	}

	var sites []refusalSite
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			rel, relErr := filepath.Rel(root, name)
			if relErr != nil {
				rel = name
			}
			rel = filepath.ToSlash(rel)
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					return true
				}
				if !isRefusalConstructor(call.Fun) {
					return true
				}
				literal, ok := call.Args[0].(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}
				id, err := strconv.Unquote(literal.Value)
				if err != nil {
					return true
				}
				sites = append(sites, refusalSite{
					Name: id,
					File: rel,
					Line: fset.Position(call.Lparen).Line,
				})
				return true
			})
		}
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].File != sites[j].File {
			return sites[i].File < sites[j].File
		}
		return sites[i].Line < sites[j].Line
	})
	return sites, nil
}

// isRefusalConstructor reports whether a call names one of the functions every
// refusal is made through, written plainly or as a method on the checker.
func isRefusalConstructor(fun ast.Expr) bool {
	var name string
	switch f := fun.(type) {
	case *ast.Ident:
		name = f.Name
	case *ast.SelectorExpr:
		name = f.Sel.Name
	default:
		return false
	}
	for _, known := range refusalConstructors {
		if name == known {
			return true
		}
	}
	return false
}

// A coveredBlock is one basic block the coverage run reported as executed.
type coveredBlock struct {
	start, end int
}

// coveredLines runs the surface's own suite under coverage and returns the
// blocks that were executed, keyed by the source file they are in.
func coveredLines(root string) (map[string][]coveredBlock, error) {
	dir, err := os.MkdirTemp("", "messbuch-cover")
	if err != nil {
		return nil, fmt.Errorf("cannot make a place for the coverage profile: %w", err)
	}
	defer os.RemoveAll(dir)
	profile := filepath.Join(dir, "cover.out")

	cmd := exec.Command("go", "test", "-count=1", "-covermode=set", "-coverprofile="+profile, "./"+refusalSurface+"/...")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("the suite behind the refusal surface did not pass, so nothing here says which sites are reached:\n%s", strings.TrimSpace(string(out)))
	}

	f, err := os.Open(profile)
	if err != nil {
		return nil, fmt.Errorf("cannot read the coverage profile, so the reached set is unknown: %w", err)
	}
	defer f.Close()

	covered := map[string][]coveredBlock{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		name, block, count, err := parseProfileLine(line)
		if err != nil {
			return nil, fmt.Errorf("cannot read the coverage profile: %w", err)
		}
		if count == 0 {
			continue
		}
		covered[name] = append(covered[name], block)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read the coverage profile: %w", err)
	}
	if len(covered) == 0 {
		return nil, fmt.Errorf("the coverage profile reports no executed block at all, which is a broken measurement rather than an unproven surface")
	}
	return covered, nil
}

// parseProfileLine reads one line of a coverage profile:
//
//	<import path>/<file>:<startLine>.<startCol>,<endLine>.<endCol> <statements> <count>
func parseProfileLine(line string) (file string, block coveredBlock, count int, err error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", block, 0, fmt.Errorf("a line with %d fields rather than 3: %q", len(fields), line)
	}
	count, err = strconv.Atoi(fields[2])
	if err != nil {
		return "", block, 0, fmt.Errorf("a count that is not a number in %q", line)
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon < 0 {
		return "", block, 0, fmt.Errorf("no position in %q", line)
	}
	file = fields[0][:colon]
	span := fields[0][colon+1:]
	comma := strings.Index(span, ",")
	if comma < 0 {
		return "", block, 0, fmt.Errorf("no block span in %q", line)
	}
	if block.start, err = lineOf(span[:comma]); err != nil {
		return "", block, 0, fmt.Errorf("%w in %q", err, line)
	}
	if block.end, err = lineOf(span[comma+1:]); err != nil {
		return "", block, 0, fmt.Errorf("%w in %q", err, line)
	}
	return file, block, count, nil
}

// lineOf reads the line number out of a <line>.<column> pair.
func lineOf(pair string) (int, error) {
	dot := strings.Index(pair, ".")
	if dot < 0 {
		return 0, fmt.Errorf("no line and column in %q", pair)
	}
	return strconv.Atoi(pair[:dot])
}

// reached reports whether an executed block covers the site's line.
//
// The profile names a file by its import path and the site by the path it was
// parsed from, so the two are matched on the suffix that is the file itself.
// That is exact enough here because the surface is one package: two files of
// one name in one package cannot exist.
func reached(covered map[string][]coveredBlock, site refusalSite) bool {
	base := filepath.Base(site.File)
	for name, blocks := range covered {
		if filepath.Base(name) != base {
			continue
		}
		for _, block := range blocks {
			if block.start <= site.Line && site.Line <= block.end {
				return true
			}
		}
	}
	return false
}
