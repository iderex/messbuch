package gate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every tree here is written by the test. A suite reading this repository's own
// changelog would report that there are no releases today, and go on reporting
// it after the leg stopped working.

const changelogShape = `# Changelog

# The tool

## Unreleased

Nothing yet.

# The corpus

## Unreleased

Nothing yet.
`

// changelogTree makes a repository carrying a changelog and the given tags.
func changelogTree(t *testing.T, changelog string, tagNames ...string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ChangelogFile), changelog)

	gitIn(t, root, "init")
	// A tag needs a commit to point at. The fixture's own commit and tags are
	// made unsigned, and that is not the bypass this project refuses: what is
	// refused is an unsigned commit reaching a branch of this repository, and
	// nothing here leaves the temporary directory. A fixture that signed would
	// need a key, would prompt whoever ran the suite, and would fail on any
	// machine without one.
	writeFile(t, filepath.Join(root, "README.md"), "a tree with releases in it\n")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid",
		"-c", "commit.gpgsign=false", "commit", "-m", "fixture")
	for _, name := range tagNames {
		gitIn(t, root, "-c", "tag.gpgsign=false", "-c", "tag.forceSignAnnotated=false", "tag", name)
	}
	return root
}

func gitIn(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func TestChangelogPassesWithNoReleasesAndSaysSo(t *testing.T) {
	root := changelogTree(t, changelogShape)

	examined, err := changelogLeg(root)
	if err != nil {
		t.Fatalf("a repository with no release and a changelog with nothing in it was refused: %v", err)
	}
	if !strings.Contains(examined, "0 release(s) of the tool") || !strings.Contains(examined, "0 release(s) of the corpus") {
		t.Errorf("the leg does not say what it counted: %q", examined)
	}
}

// The refusal this leg exists for. A release happened and the document a reader
// would go to says nothing about it.
func TestATagWithNoEntryIsRefused(t *testing.T) {
	root := changelogTree(t, changelogShape, "corpus-v1.0.0")

	_, err := changelogLeg(root)
	if err == nil {
		t.Fatal("a corpus release with no changelog entry passed, so a reader holding a number from it is told nothing")
	}
	if !strings.Contains(err.Error(), "corpus-v1.0.0") {
		t.Errorf("the refusal does not name the tag: %v", err)
	}
	if !strings.Contains(err.Error(), "The corpus") {
		t.Errorf("the refusal does not say which stream is missing the entry: %v", err)
	}
}

// The two streams are separate, and a tool entry does not answer for a corpus
// release. This is the failure a single stream would have hidden.
func TestAnEntryUnderTheWrongStreamDoesNotAnswerForATag(t *testing.T) {
	changelog := strings.Replace(changelogShape, `# The tool

## Unreleased

Nothing yet.`, `# The tool

## Unreleased

Nothing yet.

## 1.0.0

An entry under the wrong heading.`, 1)

	root := changelogTree(t, changelog, "corpus-v1.0.0")

	_, err := changelogLeg(root)
	if err == nil {
		t.Fatal("an entry under the tool answered for a corpus release, so the two streams are one")
	}
	if !strings.Contains(err.Error(), "corpus-v1.0.0 is a tag") {
		t.Errorf("the refusal does not say the corpus tag has no entry: %v", err)
	}
	if !strings.Contains(err.Error(), "and no v1.0.0 tag exists") {
		t.Errorf("the refusal does not say the tool entry runs ahead of its tag: %v", err)
	}
}

// The other direction. A changelog running ahead of the tags tells a reader
// about a release they cannot fetch.
func TestAnEntryAheadOfItsTagIsRefused(t *testing.T) {
	changelog := strings.Replace(changelogShape, `# The corpus

## Unreleased

Nothing yet.`, `# The corpus

## Unreleased

Nothing yet.

## 2.0.0

A release that has not been tagged.`, 1)

	root := changelogTree(t, changelog)

	_, err := changelogLeg(root)
	if err == nil {
		t.Fatal("a changelog naming a release with no tag passed")
	}
	if !strings.Contains(err.Error(), "cannot fetch") {
		t.Errorf("the refusal does not say what is wrong with it: %v", err)
	}
}

// A tagged release that is written down passes, which is the state this leg
// wants rather than a state it merely tolerates.
func TestATaggedReleaseWithItsEntryPasses(t *testing.T) {
	changelog := strings.Replace(changelogShape, `# The corpus

## Unreleased

Nothing yet.`, `# The corpus

## Unreleased

Nothing yet.

## 1.0.0

The first series, and one corrected record naming its old and new value.`, 1)

	root := changelogTree(t, changelog, "corpus-v1.0.0")

	examined, err := changelogLeg(root)
	if err != nil {
		t.Fatalf("a release with its entry was refused: %v", err)
	}
	if !strings.Contains(examined, "1 release(s) of the corpus") {
		t.Errorf("the leg does not count the release it read: %q", examined)
	}
}

// It fails closed, and each of these would otherwise be a green line about
// nothing.
func TestChangelogFailsClosed(t *testing.T) {
	t.Run("no changelog at all", func(t *testing.T) {
		root := t.TempDir()
		gitIn(t, root, "init")
		if _, err := changelogLeg(root); err == nil {
			t.Fatal("a tree with no changelog passed")
		}
	})

	t.Run("a stream with no heading", func(t *testing.T) {
		root := changelogTree(t, "# Changelog\n\n# The tool\n\n## Unreleased\n\nNothing yet.\n")
		if _, err := changelogLeg(root); err == nil {
			t.Fatal("a changelog with no corpus stream passed, so a corpus release has nowhere to be written")
		}
	})

	t.Run("a stream with nowhere to write the next change", func(t *testing.T) {
		root := changelogTree(t, strings.Replace(changelogShape, "# The corpus\n\n## Unreleased", "# The corpus\n\n## 1.0.0", 1), "corpus-v1.0.0")
		if _, err := changelogLeg(root); err == nil {
			t.Fatal("a stream with no Unreleased section passed, so the next change has nowhere to go")
		}
	})

	t.Run("outside a repository", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ChangelogFile), changelogShape)
		if _, err := changelogLeg(root); err == nil {
			t.Fatal("a directory git cannot answer about was read as one with no releases")
		}
	})
}

// The real file is read once, and only for the property the leg would be
// useless without: that this repository's own changelog carries both streams.
// What it says is the review's business.
func TestThisRepositoryCarriesBothStreams(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", ChangelogFile))
	if err != nil {
		t.Fatalf("cannot read %s: %v", ChangelogFile, err)
	}
	for _, stream := range streams {
		if !strings.Contains(string(content), stream.Heading+"\n") {
			t.Errorf("%s carries no heading for %s", ChangelogFile, stream.Name)
		}
	}
}
