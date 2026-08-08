package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// toolchainLeg refuses a run made by a toolchain other than the one go.mod
// pins.
//
// The failure it prevents is a number that is reproducible on one machine and
// not on another. The build entry point exists so that the toolchain version
// is a decision somebody took rather than whatever a runner happened to carry,
// and a pin nothing reads is a comment. This reads it, on every run, and it
// fails closed: a go.mod that cannot be read, or that carries no toolchain
// line, is a refusal rather than a pass.
func toolchainLeg(root string) (string, error) {
	path := filepath.Join(root, "go.mod")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read %s, so the pin cannot be checked: %w", path, err)
	}
	pinned, err := pinnedToolchain(b)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	running := runtime.Version()
	if err := checkToolchain(pinned, running); err != nil {
		return "", err
	}
	return fmt.Sprintf("go.mod pins %s and this run is %s", pinned, running), nil
}

// pinnedToolchain returns the argument of the toolchain line in go.mod.
//
// The line is read rather than the go line, because the two answer different
// questions: the go line is the language version the source may use, and the
// toolchain line is the release that builds it. Only the second is what a
// reproduced number depends on.
func pinnedToolchain(goMod []byte) (string, error) {
	for _, line := range strings.Split(string(goMod), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 2 && fields[0] == "toolchain" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("no toolchain line, so nothing pins the release that builds this module")
}

// checkToolchain compares the pin against the running release.
func checkToolchain(pinned, running string) error {
	if pinned != running {
		return fmt.Errorf("go.mod pins %s and this run is %s, so a number produced here is not the pinned toolchain's number", pinned, running)
	}
	return nil
}
