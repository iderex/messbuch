package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pinnedGoMod = `module github.com/iderex/messbuch

go 1.26.0

toolchain go1.26.5
`

func TestPinnedToolchainReadsTheLine(t *testing.T) {
	got, err := pinnedToolchain([]byte(pinnedGoMod))
	if err != nil {
		t.Fatalf("pinnedToolchain returned %v", err)
	}
	if got != "go1.26.5" {
		t.Errorf("pin is %q, want %q", got, "go1.26.5")
	}
}

// The near miss. A go.mod carrying a go line and no toolchain line is the
// shape somebody produces by running go mod init and stopping, and it is the
// one this leg has to refuse: the go line is the language version and pins no
// release at all.
func TestPinnedToolchainRefusesAGoModWithOnlyAGoLine(t *testing.T) {
	only := "module github.com/iderex/messbuch\n\ngo 1.26.0\n"
	if _, err := pinnedToolchain([]byte(only)); err == nil {
		t.Fatal("a go.mod with no toolchain line was accepted")
	}
}

// The word appearing inside another directive is not a pin. Without the field
// count this matched, and a require line naming a module called toolchain
// would have set the pin.
func TestPinnedToolchainRefusesAToolchainWordInAnotherDirective(t *testing.T) {
	confusing := "module github.com/iderex/messbuch\n\ngo 1.26.0\n\nrequire example.com/toolchain v1.0.0\n"
	if _, err := pinnedToolchain([]byte(confusing)); err == nil {
		t.Fatal("a require line was read as a toolchain pin")
	}
}

func TestCheckToolchainAcceptsTheMatch(t *testing.T) {
	if err := checkToolchain("go1.26.5", "go1.26.5"); err != nil {
		t.Errorf("a matching pin was refused: %v", err)
	}
}

// The refusal this leg exists for, and the near miss beside it: a patch
// release apart is exactly the difference somebody would call close enough,
// and it is the difference a reproduced number depends on.
func TestCheckToolchainRefusesAMismatch(t *testing.T) {
	for _, running := range []string{"go1.26.4", "go1.27.0", "go1.26", "devel"} {
		err := checkToolchain("go1.26.5", running)
		if err == nil {
			t.Errorf("running %q against pin go1.26.5 was accepted", running)
			continue
		}
		if !strings.Contains(err.Error(), running) {
			t.Errorf("the refusal for %q does not name what was running: %v", running, err)
		}
	}
}

// Fail closed. A go.mod that cannot be read is a refusal, not a pass, because
// a missing pin and a satisfied pin must not print the same way.
func TestToolchainLegRefusesATreeWithNoGoMod(t *testing.T) {
	if _, err := toolchainLeg(t.TempDir()); err == nil {
		t.Fatal("a tree with no go.mod passed the toolchain leg")
	}
}

// The leg reads the file it names rather than a copy of it.
func TestToolchainLegReadsGoModFromTheTree(t *testing.T) {
	root := t.TempDir()
	wrong := strings.Replace(pinnedGoMod, "go1.26.5", "go0.0.1", 1)
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(wrong), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := toolchainLeg(root); err == nil {
		t.Fatal("a go.mod pinning a release nothing is running passed")
	}
}
