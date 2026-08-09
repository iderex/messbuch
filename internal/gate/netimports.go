package gate

import (
	"fmt"
	"sort"
	"strings"
)

// networkCapable is the list of import paths that count as network-capable.
//
// It is data rather than a condition inside the matcher below, so adding a
// name is a reviewable line in this table rather than an edit to code that
// decides things. docs/decisions/0009-offline-by-default.md asks for exactly
// that shape, and the reason is staleness: a list that is hard to extend stops
// being extended, and a check that has gone stale reads as enforcement while
// enforcing less than it claims.
//
// Matching is on the exact import path and never on a prefix. A prefix rule
// looks tighter and is wrong in a way that costs somebody an afternoon:
// net/url parses a string and opens nothing, so a rule reading "net/" would
// refuse a validator for reading the source URL out of a record.
var networkCapable = []struct {
	Path string
	Why  string
}{
	{"net", "dials, listens and resolves; every standard-library package that opens a socket reaches it"},
	{"net/http", "client and server over TCP"},
	{"net/rpc", "serves and calls over a connection"},
	{"net/smtp", "sends mail over a connection"},
	{"crypto/tls", "opens and terminates TLS connections"},
}

// permittedSuffix is the one package allowed to reach anything in the table
// above, relative to the module path.
//
// The name is fixed by docs/decisions/0002-language-and-toolchain.md rather
// than here, and the check string the branch protection matches is fixed by
// docs/decisions/0009-offline-by-default.md.
const permittedSuffix = "/internal/net"

// noNetworkImportsLeg refuses a package outside internal/net that transitively
// reaches a network-capable API.
//
// The failure it prevents is the offline guarantee decaying into a promise. A
// package that cannot import anything network-capable cannot open a socket
// whatever its code says and whatever flags are set, which is the whole reason
// docs/decisions/0009-offline-by-default.md chose a structural boundary over a
// runtime one. Without this leg, that record describes an intention.
//
// It fails closed. An import graph that cannot be computed is a refusal, and
// an empty package set is treated as a graph that was never read rather than
// as a clean tree, because the two are indistinguishable in the output alone.
func noNetworkImportsLeg(root string) (string, error) {
	module, err := modulePath(root)
	if err != nil {
		return "", err
	}
	reaches, err := importGraph(root)
	if err != nil {
		return "", err
	}
	own := ownPackages(module, reaches)
	if len(own) == 0 {
		return "", fmt.Errorf("no package of module %s appears in the import graph, so this leg examined nothing", module)
	}

	permitted := module + permittedSuffix
	// The permitted package's own dependencies are permitted too. Without
	// this, a helper that internal/net reaches would be refused for being
	// reached, and the record permits exactly "net and its own dependencies".
	allowed := map[string]bool{}
	for pkg, deps := range reaches {
		if !inSubtree(pkg, permitted) {
			continue
		}
		allowed[pkg] = true
		for dep := range deps {
			allowed[dep] = true
		}
	}

	var refused []string
	for _, pkg := range own {
		if allowed[pkg] {
			continue
		}
		if hits := networkHits(reaches[pkg]); len(hits) > 0 {
			refused = append(refused, fmt.Sprintf("%s reaches %s", pkg, strings.Join(hits, ", ")))
		}
	}
	if len(refused) > 0 {
		return "", fmt.Errorf("%d package(s) outside %s reach a network-capable API:\n  %s",
			len(refused), permitted, strings.Join(refused, "\n  "))
	}

	return fmt.Sprintf("%d package(s) of %s against %d network-capable path(s); only %s may reach one",
		len(own), module, len(networkCapable), permitted), nil
}

// limitsOfNoNetworkImports is printed beside the leg's result rather than
// living only in the decision record, because a boundary whose limits are one
// document away gets quoted without them.
const limitsOfNoNetworkImports = `this refuses a package that can reach a socket through the import graph.
It does not refuse a package that shells out to a program that opens one, and
it does not refuse a dependency that opens a socket from inside code the graph
shows as reached for another reason. A network-capable path absent from the
table in internal/gate/netimports.go is not refused either.`

// modulePath reads the module path of the tree rooted at root.
//
// A tree whose module path cannot be read is a refusal: the permitted package
// is named relative to it, so without it there is nothing to compare against.
func modulePath(root string) (string, error) {
	out, err := runGo(root, "list", "-m")
	if err != nil {
		return "", fmt.Errorf("cannot read the module path of %s, so the import graph has no subject: %w", root, err)
	}
	got := lines(out)
	if len(got) == 0 {
		return "", fmt.Errorf("go list -m printed nothing in %s, so the import graph has no subject", root)
	}
	return got[0], nil
}

// importGraph returns, for every package of the tree, the whole set of
// packages it reaches.
//
// The reach is read off go list's own Deps field rather than walked here, so
// the transitive closure this leg judges is the one the compiler would build
// and not a second implementation of it. -test brings in the packages built
// for tests, so a test file importing a network-capable package is refused on
// the same terms as a source file: "only a test" is exactly where somebody
// reaches for a live server.
func importGraph(root string) (map[string]map[string]bool, error) {
	out, err := runGo(root, "list", "-test", "-f", "{{.ImportPath}}|{{join .Deps \" \"}}", "./...")
	if err != nil {
		return nil, fmt.Errorf("cannot compute the import graph of %s, so this leg refuses rather than passing a graph it never read: %w", root, err)
	}
	reaches := map[string]map[string]bool{}
	for _, l := range lines(out) {
		path, deps, ok := strings.Cut(l, "|")
		if !ok {
			return nil, fmt.Errorf("go list printed a line this leg cannot read: %q", l)
		}
		path = stripTestVariant(path)
		if reaches[path] == nil {
			reaches[path] = map[string]bool{}
		}
		for _, dep := range strings.Fields(deps) {
			reaches[path][stripTestVariant(dep)] = true
		}
	}
	if len(reaches) == 0 {
		return nil, fmt.Errorf("the import graph of %s is empty, which this leg reads as a graph it failed to compute rather than as a tree with no imports", root)
	}
	return reaches, nil
}

// stripTestVariant folds the packages go list builds for a test back onto the
// package they belong to.
//
// With -test, a package p appears three more times: "p [p.test]" for the build
// that includes its own test files, "p_test [p.test]" for the external test
// package, and "p.test" for the generated binary. The first two carry the
// imports a test file wrote, and attributing them to p is what makes a test
// that imports a network-capable package refusable on the same terms as a
// source file. The binary is left as its own node and dropped in ownPackages,
// since everything it imports is already attributed through the other two.
func stripTestVariant(path string) string {
	if i := strings.Index(path, " ["); i >= 0 {
		return strings.TrimSuffix(path[:i], "_test")
	}
	return path
}

// ownPackages lists the packages of the module under test, sorted, so a
// refusal names them in a fixed order.
//
// The generated test binary is not one of them. Its import path is the
// package's with ".test" appended, a shape a real package cannot have here
// because the last element of a package path in this tree is a directory name.
func ownPackages(module string, reaches map[string]map[string]bool) []string {
	var own []string
	for pkg := range reaches {
		if inSubtree(pkg, module) && !strings.HasSuffix(pkg, ".test") {
			own = append(own, pkg)
		}
	}
	sort.Strings(own)
	return own
}

// inSubtree reports whether pkg is root or sits under it.
func inSubtree(pkg, root string) bool {
	return pkg == root || strings.HasPrefix(pkg, root+"/")
}

// networkHits returns every network-capable path in reached, in the table's
// declared order so a refusal reads the same on every run.
//
// All of them rather than the first. The first is nearly always net, since
// every socket in the standard library reaches it, and a refusal saying only
// that sends the reader looking for an import of net that nobody wrote.
func networkHits(reached map[string]bool) []string {
	var hits []string
	for _, entry := range networkCapable {
		if reached[entry.Path] {
			hits = append(hits, entry.Path)
		}
	}
	return hits
}
