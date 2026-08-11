package gate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ChangelogFile is where both release streams are written.
const ChangelogFile = "CHANGELOG.md"

// The two streams, and the tag prefix each one releases under.
//
// Two rather than one, and the reason is not a preference: the corpus
// versioning record fixes that the tool version moves on its own schedule and
// that neither release implies the other. A single stream here would tie a data
// correction to a tool release it has nothing to do with.
var streams = []struct {
	Name    string
	Heading string
	Match   string
	Prefix  string
}{
	{"the tool", "# The tool", "v[0-9]*", "v"},
	{"the corpus", "# The corpus", "corpus-v[0-9]*", "corpus-v"},
}

// unreleasedHeading is the section every stream has to carry, so that a change
// landing today has somewhere to be written rather than somewhere to be
// invented.
const unreleasedHeading = "## Unreleased"

// versionHeading matches a released version's heading inside a stream.
var versionHeading = regexp.MustCompile(`(?m)^## ([0-9]+\.[0-9]+\.[0-9]+)\b`)

// changelogLeg refuses a release nobody wrote down and an entry nobody
// released.
//
// The failure it prevents has a name on this board: somebody publishes a number
// computed from a record, the record is corrected, and the correction reaches no
// document they will ever read. A tag with no entry is exactly that, and it is
// invisible from inside the tag.
//
// The other direction is refused too, because a changelog that runs ahead of the
// tags is a reader being told about a release they cannot fetch.
//
// It fails closed. A missing changelog, a changelog with a stream heading
// missing, and a tag list that cannot be read are all refusals rather than a
// pass over nothing.
func changelogLeg(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, ChangelogFile))
	if err != nil {
		return "", fmt.Errorf("cannot read %s, so nothing here says what a release changed: %w", ChangelogFile, err)
	}
	text := string(content)

	tagged, err := tags(root)
	if err != nil {
		return "", err
	}

	var problems []string
	var counted []string
	for _, stream := range streams {
		section, ok := streamSection(text, stream.Heading)
		if !ok {
			return "", fmt.Errorf("%s carries no %q heading, so %s has nowhere to record a release", ChangelogFile, stream.Heading, stream.Name)
		}
		if !strings.Contains(section, unreleasedHeading) {
			return "", fmt.Errorf("%s carries no %q under %q, so a change landing today has nowhere to be written", ChangelogFile, unreleasedHeading, stream.Heading)
		}

		written := map[string]bool{}
		for _, match := range versionHeading.FindAllStringSubmatch(section, -1) {
			written[match[1]] = true
		}
		released := releasedVersions(tagged, stream.Prefix)

		for _, version := range sortedKeys(released) {
			if !written[version] {
				problems = append(problems, fmt.Sprintf("%s%s is a tag and %s has no entry for it under %q", stream.Prefix, version, ChangelogFile, stream.Heading))
			}
		}
		for _, version := range sortedKeys(written) {
			if !released[version] {
				problems = append(problems, fmt.Sprintf("%s names %s under %q and no %s%s tag exists, so a reader is told about a release they cannot fetch", ChangelogFile, version, stream.Heading, stream.Prefix, version))
			}
		}
		counted = append(counted, fmt.Sprintf("%d release(s) of %s", len(released), stream.Name))
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return "", fmt.Errorf("%d release(s) and entries disagree:\n  %s\n\nA corrected record is the entry that matters most here, because somebody may have published a number computed from the old value.",
			len(problems), strings.Join(problems, "\n  "))
	}
	return fmt.Sprintf("%s against the tags: %s, each with an entry and nothing written ahead of a tag", ChangelogFile, strings.Join(counted, " and ")), nil
}

// streamSection returns the part of the changelog under one stream heading.
func streamSection(text, heading string) (string, bool) {
	start := strings.Index(text, heading+"\n")
	if start < 0 {
		return "", false
	}
	rest := text[start+len(heading):]
	for _, other := range streams {
		if next := strings.Index(rest, "\n"+other.Heading+"\n"); next >= 0 {
			rest = rest[:next]
		}
	}
	return rest, true
}

// releasedVersions is the set of versions a stream's tags name.
func releasedVersions(tagged []string, prefix string) map[string]bool {
	out := map[string]bool{}
	for _, tag := range tagged {
		if !strings.HasPrefix(tag, prefix) {
			continue
		}
		version := strings.TrimPrefix(tag, prefix)
		// The corpus prefix contains the tool prefix, so a corpus tag would
		// otherwise be read as a tool release of a version beginning with a
		// letter. A version is three numbers and nothing else.
		if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(version) {
			continue
		}
		out[version] = true
	}
	return out
}

// tags is every tag in the repository.
//
// Read from git rather than from a file, because a tag is what a citation names
// and a list of them in the tree would be a second answer that can disagree with
// the first.
func tags(root string) ([]string, error) {
	cmd := exec.Command("git", "tag", "--list")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot list the tags, so which releases exist is unknown rather than none: %w", err)
	}
	var found []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			found = append(found, line)
		}
	}
	return found, nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// limitsOfChangelog says where the leg stops, printed beside its result.
const limitsOfChangelog = `this compares the release tags against the headings in CHANGELOG.md, in both
directions, and refuses a stream with nowhere to write the next change. It does
not read what an entry says: whether a corrected record names the old value and
the new, and whether a breaking change is described as one, are judgements no
reading of this tree makes. It also cannot see a release that was published
without a tag, because a tag is what it counts, and the versioning record
already says publishing from anything other than a tag is not a release. On a
clone fetched without tags it reads an empty tag list, which is the shape of a
shallow checkout rather than of a repository with no releases.`
