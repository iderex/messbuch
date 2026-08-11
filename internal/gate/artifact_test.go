package gate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iderex/messbuch/internal/build"
)

// repo makes a git repository holding the given files, staged, and returns its
// root.
//
// Staged rather than committed, because the leg reads the index and a commit
// would add a signature and an identity to a fixture that needs neither.
func repo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()

	run(t, root, "init")
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("cannot create %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("cannot write %s: %v", rel, err)
		}
		run(t, root, "add", "--force", "--", rel)
	}
	return root
}

func run(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// The ordinary case: a tree where the build has run and nothing it wrote was
// added. The leg has to say what it examined rather than only that it passed.
func TestArtifactUntrackedPassesWhenNoOutputIsTracked(t *testing.T) {
	root := repo(t, map[string]string{"README.md": "a tree with no build output in it\n"})

	// The build's own output on disk is not the thing being refused, so it is
	// present here: an untracked artifact is the state this leg exists to
	// leave alone.
	if err := os.MkdirAll(filepath.Join(root, build.Dir), 0o755); err != nil {
		t.Fatalf("cannot create %s: %v", build.Dir, err)
	}
	if err := os.WriteFile(filepath.Join(root, build.Dir, build.JSONName), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("cannot write the artifact: %v", err)
	}

	examined, err := artifactUntrackedLeg(root)
	if err != nil {
		t.Fatalf("an untracked build output is the ordinary state and was refused: %v", err)
	}
	if !strings.Contains(examined, build.Dir) {
		t.Errorf("the leg does not say what it examined: %q", examined)
	}
}

// The refusal, and it bites on the file the fork in #27 would have created:
// the lossless artifact, committed.
func TestArtifactUntrackedRefusesATrackedArtifact(t *testing.T) {
	root := repo(t, map[string]string{
		build.Dir + "/" + build.JSONName: "{\"stamp\":{}}\n",
	})

	_, err := artifactUntrackedLeg(root)
	if err == nil {
		t.Fatal("a committed build output passed, so nothing stops one being edited by hand")
	}
	if !strings.Contains(err.Error(), build.JSONName) {
		t.Errorf("the refusal does not name the file it is about: %v", err)
	}
	if !strings.Contains(err.Error(), "go run . build") {
		t.Errorf("the refusal does not say what to do instead: %v", err)
	}
}

// Every output the build writes is reached, not only the first. A format added
// to internal/build arrives here without a line of this file moving, and the
// leg would otherwise cover whichever one somebody remembered.
func TestArtifactUntrackedReachesEveryOutput(t *testing.T) {
	for _, output := range build.Outputs() {
		root := repo(t, map[string]string{output: "tracked\n"})
		if _, err := artifactUntrackedLeg(root); err == nil {
			t.Errorf("%s is a build output and was tracked without being refused", output)
		}
	}
}

// It fails closed. A directory git cannot answer about is a refusal rather
// than a tree read as holding nothing.
func TestArtifactUntrackedFailsClosedOutsideARepository(t *testing.T) {
	if _, err := artifactUntrackedLeg(t.TempDir()); err == nil {
		t.Fatal("a directory that is not a repository was read as a tree tracking no build output")
	}
}
