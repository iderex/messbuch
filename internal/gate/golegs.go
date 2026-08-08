package gate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runGo runs one go subcommand in root with the module cache locked.
//
// GOFLAGS carries -mod=readonly so that a build never edits go.mod or go.sum
// to make itself work. That is the whole of "restores in locked mode" on this
// toolchain: a requirement missing from go.mod, or a sum missing from go.sum,
// is an error at the point of use rather than a file the build quietly
// rewrites and a drift nobody sees in the diff.
//
// GOWORK is emptied because a workspace file outside this repository would
// silently replace a pinned module with a directory on somebody's disk, and a
// gate that can be redirected by a file it does not track is not a gate.
func runGo(root string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=readonly", "GOWORK=off")

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(errOut.String())
		if detail == "" {
			detail = strings.TrimSpace(out.String())
		}
		return "", fmt.Errorf("go %s: %v\n%s", strings.Join(args, " "), err, detail)
	}
	return out.String(), nil
}

// lines splits command output into non-empty lines.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// modulesLeg refuses a dependency set that does not match what go.sum records.
//
// The failure it prevents is a silent version drift: a module resolved to
// different bytes than the ones somebody reviewed, with nothing in the diff to
// show it. go mod verify recomputes the hash of every module in the build list
// and compares it against go.sum, so a replaced or re-tagged module is a
// refusal here rather than a surprise in a released binary.
func modulesLeg(root string) (string, error) {
	if _, err := runGo(root, "mod", "verify"); err != nil {
		return "", err
	}
	out, err := runGo(root, "list", "-m", "all")
	if err != nil {
		return "", err
	}
	mods := lines(out)
	// The first line is this module itself, which go.sum does not record.
	return fmt.Sprintf("%d module(s) in the build list verify against go.sum", max(len(mods)-1, 0)), nil
}

// buildLeg refuses source that does not compile.
func buildLeg(root string) (string, error) {
	out, err := runGo(root, "list", "./...")
	if err != nil {
		return "", err
	}
	pkgs := lines(out)
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package under %s, so this leg examined nothing", root)
	}
	if _, err := runGo(root, "build", "./..."); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d package(s) build", len(pkgs)), nil
}

// testLeg refuses a failing test.
//
// -count=1 is not a preference. A cached result is a statement about a run
// that happened at some other time, and this leg exists to say what this run
// examined.
func testLeg(root string) (string, error) {
	out, err := runGo(root, "test", "-count=1", "./...")
	if err != nil {
		return "", err
	}
	var tested int
	for _, l := range lines(out) {
		if strings.HasPrefix(l, "ok ") || strings.HasPrefix(l, "?") {
			tested++
		}
	}
	return fmt.Sprintf("%d package(s) tested", tested), nil
}
