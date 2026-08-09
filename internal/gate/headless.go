package gate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// OnlineHarness is the directory holding the tests that need the outside
// world, and the tag that keeps them out of every ordinary build.
//
// Named here so that the leg that excludes it and the line the gate prints
// about it read the same constant, rather than two strings that agree until
// one is edited.
const (
	OnlineHarness    = "test/online"
	OnlineTag        = "online"
	OnlineInvocation = "go test -tags online ./test/online/..."
)

// limitsOfTheTestLeg is what the gate prints beside its test result, so that a
// green run says which suite it did not cover.
//
// The harness is excluded rather than merely unmentioned. A configuration that
// never named it would be byte for byte the configuration of a gate that had
// never heard of it, and those two states are what
// docs/decisions/0010-headless-tests.md exists to keep apart. The exclusion is
// written and printed rather than achieved by the build constraint alone.
var limitsOfTheTestLeg = fmt.Sprintf(`%s is excluded from this leg and was not run. Its files carry the build
constraint %q, so none of them is compiled into an ordinary run, and what it
tests needs the outside world. It is invoked deliberately and reports the
number of tests it ran, zero included:

    %s`, OnlineHarness, OnlineTag, OnlineInvocation)

// displayNames are the environment variables that name a display.
//
// Data rather than a condition inside the matcher, for the same reason the
// network table is: a name added here is a reviewable line.
var displayNames = []string{"DISPLAY", "WAYLAND_DISPLAY", "XAUTHORITY"}

// privilegedTools are programs a test cannot run without asking for rights it
// should not have, or that prompt the person at the machine.
//
// The list is short and unambiguous on purpose. A generic name would refuse a
// test for mentioning a word, and a refusal nobody believes is a refusal
// people route around.
var privilegedTools = []string{
	"sudo", "doas", "pkexec", "su",
	"runas", "runas.exe", "netsh", "netsh.exe", "sc.exe",
	"setcap", "chown", "mount", "systemctl", "launchctl",
}

// homeReaders are the standard-library calls that resolve a directory belonging
// to whoever is logged in rather than to the repository.
var homeReaders = []string{"os.UserHomeDir", "os.UserConfigDir", "os.UserCacheDir"}

// absolutePath matches a string literal that names a place on a machine rather
// than a place in this repository.
//
// A leading double slash is excluded because "//go:build online" is a line of
// Go source, not a path, and a fixture carrying one is ordinary. A literal
// containing a space is excluded for the same reason: it is prose.
var absolutePath = regexp.MustCompile(`^(/[^/ ]|~/|[A-Za-z]:[\\/])`)

// headlessLeg refuses a test that needs something the gate does not have.
//
// The rule is that every test in the gate runs with no display, no elevation,
// no network and no external service. A sentence in a document is not that
// rule; this is. The failure it prevents is the one that cannot be repaired
// later: a test that quietly acquires a dependency on its environment looks
// exactly like one that did not until it is run somewhere without it, and
// finding them afterwards is an audit rather than a decision.
//
// It reads the test files rather than the running tests, so what it can say is
// bounded by what source shows. That bound is printed beside its result.
//
// test/online is not scanned. It is the harness for the tests that do need the
// outside world, it is kept out of every ordinary build by its own constraint,
// and the gate names it as not run rather than passing over it in silence.
func headlessLeg(root string) (string, error) {
	files, err := testFiles(root)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no test file under %s outside %s, so this leg examined nothing", root, OnlineHarness)
	}

	var found []string
	for _, rel := range files {
		breaches, err := headlessBreaches(filepath.Join(root, rel), rel)
		if err != nil {
			return "", err
		}
		found = append(found, breaches...)
	}
	if len(found) > 0 {
		sort.Strings(found)
		return "", fmt.Errorf("%d test(s) reach for something the gate does not have:\n  %s",
			len(found), strings.Join(found, "\n  "))
	}

	return fmt.Sprintf("%d test file(s) reach no display, no elevation, no network and no path outside this repository; %s is excluded and was not run",
		len(files), OnlineHarness), nil
}

// limitsOfHeadless says where the leg stops, printed beside its result.
const limitsOfHeadless = `this reads what the test source shows: its imports, the strings it names and
the standard-library calls that resolve somebody's home directory. It cannot
see what a program a test starts goes on to do, it cannot see a capability
reached through a dependency that names none of these, and it does not run the
tests to find out. A tool absent from the list in internal/gate/headless.go is
not refused either.`

// headlessBreaches parses one test file and returns one line per condition it
// breaks, naming the test and what it reached for.
//
// Naming the test rather than the file, because a file carries many tests and
// the person reading the refusal has to know which one to open. A breach
// outside any function is reported against the file, which is where an import
// sits.
func headlessBreaches(path, rel string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		// Source that does not parse is the build leg's refusal. Two
		// refusals for one defect is one refusal too many.
		return nil, nil
	}

	var breaches []string
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		for _, entry := range networkCapable {
			if p == entry.Path {
				breaches = append(breaches, fmt.Sprintf("%s imports %s, which is the network", rel, p))
			}
		}
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		where := fmt.Sprintf("%s in %s", fn.Name.Name, rel)
		ast.Inspect(fn, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.BasicLit:
				if node.Kind != token.STRING {
					return true
				}
				s, err := strconv.Unquote(node.Value)
				if err != nil {
					return true
				}
				breaches = append(breaches, literalBreaches(where, s)...)
			case *ast.SelectorExpr:
				call := selectorName(node)
				for _, reader := range homeReaders {
					if call == reader {
						breaches = append(breaches, fmt.Sprintf("%s calls %s, which is a directory outside this repository", where, reader))
					}
				}
			}
			return true
		})
	}
	return breaches, nil
}

// literalBreaches reports what one string literal in a test reaches for.
func literalBreaches(where, s string) []string {
	var out []string
	for _, name := range displayNames {
		if s == name {
			out = append(out, fmt.Sprintf("%s names %s, which is a display", where, name))
		}
	}
	for _, tool := range privilegedTools {
		if s == tool || strings.HasSuffix(s, "/"+tool) || strings.HasSuffix(s, `\`+tool) {
			out = append(out, fmt.Sprintf("%s names %s, which needs privileges the gate does not have", where, tool))
		}
	}
	if absolutePath.MatchString(s) {
		out = append(out, fmt.Sprintf("%s names %q, which is a path outside this repository", where, s))
	}
	return out
}

// selectorName renders a selector as "pkg.Name" where the receiver is a plain
// identifier, and the empty string otherwise.
func selectorName(sel *ast.SelectorExpr) string {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name + "." + sel.Sel.Name
}

// testFiles lists the test files under root, relative to it, in a fixed order,
// with the online harness and every fixture directory left out.
func testFiles(root string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		slashed := filepath.ToSlash(rel)
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == "testdata" || slashed == OnlineHarness) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), "_test.go") {
			found = append(found, slashed)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s, so the set of test files is unknown: %w", root, err)
	}
	sort.Strings(found)
	return found, nil
}
