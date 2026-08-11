package gate

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/iderex/messbuch/internal/build"
)

// artifactUntrackedLeg refuses a build output that git is tracking.
//
// #27 left a fork open between a tracked artifact whose bytes a check rebuilds
// and compares, and an untracked one. The stamp closed it: an artifact has to
// carry the revision it was built from, and a tracked file is part of the
// commit that supplies that revision, so its bytes would have to contain an
// identifier that does not exist until after the bytes are fixed. The argument
// is written where the build is, in internal/build.
//
// That leaves this leg as the guard the fork's other branch would have needed,
// caught one step earlier. What it refuses is a build output that was
// committed, because the next thing that happens to a committed build output
// is that somebody edits it by hand when they are in a hurry, and a hand
// edited artifact is a corpus nobody validated wearing a stamp somebody
// trusts.
//
// It fails closed. A tree whose index cannot be read is a refusal rather than
// a tree with nothing tracked at those paths.
func artifactUntrackedLeg(root string) (string, error) {
	paths := build.Outputs()
	args := append([]string{"ls-files", "--"}, append(paths, build.Dir)...)

	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot list what git tracks under %s, so whether a build output is committed is unknown rather than answered: %w", build.Dir, err)
	}

	var tracked []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			tracked = append(tracked, line)
		}
	}
	if len(tracked) > 0 {
		return "", fmt.Errorf("%d build output(s) are tracked:\n  %s\n\nThe artifact is produced by `go run . build` and is not committed. It carries the revision it was built from, which a tracked file cannot hold, and a committed build output is the file somebody edits by hand.",
			len(tracked), strings.Join(tracked, "\n  "))
	}

	return fmt.Sprintf("%d output path(s) of `go run . build`, and %s/, none of them tracked", len(paths), build.Dir), nil
}

// limitsOfArtifactUntracked says where the leg stops, printed beside its
// result.
const limitsOfArtifactUntracked = `this reads the index and answers one question: whether a build output is
committed. It does not read the artifact, does not run the build, and does not
compare a built artifact against anything, so it says nothing about whether a
build is reproducible, which is #28. A build output committed at a path
internal/build does not name is not refused either, because the paths come
from the build rather than from a list here.`
