package gate

import (
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// proseExtensions are the tracked text files this leg reads beside the Go
// source.
//
// An extension list rather than "every text file", because the two files in
// this tree that are neither source nor prose of ours, LICENSE and DCO, are
// copies of texts whose bytes are not ours to normalise. Reformatting a
// license is a change to a legal instrument, and a check that did it silently
// would be worse than one that says which files it left alone.
var proseExtensions = []string{".md", ".yml", ".yaml", ".toml"}

// formatAndLintLeg refuses source that is not formatted, source the vet tool
// objects to, and prose carrying trailing whitespace or no final newline.
//
// The failure it prevents is a review spending its sentences on layout. The
// second failure it prevents is subtler and is why the fix is a command rather
// than a description: a check that says what it wants without producing it
// leaves every contributor guessing at the difference between their formatter
// and this one.
//
// The formatting question is decided by go/format, which is what gofmt itself
// calls, so "formatted" here means the same bytes gofmt would write and not a
// second opinion about them. `go run . fmt` applies exactly this, on the same
// code path, so there is no gap between what the check demands and what the
// fix produces.
//
// It fails closed. A tree whose files cannot be walked is a refusal rather
// than an empty set read as a clean one.
func formatAndLintLeg(root string) (string, error) {
	files, err := formattableFiles(root)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no Go source and no tracked prose under %s, so this leg examined nothing", root)
	}

	var wrong []string
	for _, rel := range files {
		changed, err := reformatFile(filepath.Join(root, rel), false)
		if err != nil {
			return "", err
		}
		if changed {
			wrong = append(wrong, rel)
		}
	}
	if len(wrong) > 0 {
		return "", fmt.Errorf("%d file(s) are not formatted as this repository writes them:\n  %s\n\n%s",
			len(wrong), strings.Join(wrong, "\n  "), theFixCommand)
	}

	if _, err := runGo(root, "vet", "./..."); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d file(s) formatted as written, and go vet is clean over every package", len(files)), nil
}

// theFixCommand is the one command a contributor runs, quoted in the refusal
// itself. A refusal that says a file is wrong without saying what produces the
// right bytes is a punishment rather than a message.
const theFixCommand = "Run this, and it produces exactly what this leg demands:\n\n    go run . fmt"

// limitsOfFormatAndLint says where this leg stops, printed beside its result.
const limitsOfFormatAndLint = `formatting here is what gofmt writes, plus trailing whitespace and a final
newline in tracked prose. It is not a prose style, a line length or a spelling.
The lint half is go vet and nothing else: no third-party analyser is in this
tree, so a defect vet does not name is not refused here. LICENSE and DCO are
read by neither half, because their bytes are copies of texts that are not
ours to normalise.`

// Reformat applies to root what formatAndLintLeg refuses, and returns the
// files it changed.
//
// One function serves both because the alternative is two implementations of
// "formatted" that agree until the day they do not, and the day they do not is
// the day a contributor's tree is red after they ran the fix.
func Reformat(root string) ([]string, error) {
	files, err := formattableFiles(root)
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, rel := range files {
		did, err := reformatFile(filepath.Join(root, rel), true)
		if err != nil {
			return nil, err
		}
		if did {
			changed = append(changed, rel)
		}
	}
	return changed, nil
}

// reformatFile decides what one file should contain and, when write is set,
// puts it there. It reports whether the file's bytes differ from that.
func reformatFile(path string, write bool) (bool, error) {
	before, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var after []byte
	if strings.EqualFold(filepath.Ext(path), ".go") {
		after, err = format.Source(before)
		if err != nil {
			// A file that does not parse is the build leg's refusal and
			// not this one's. Two refusals for one defect is one too many.
			return false, nil
		}
	} else {
		after = tidyProse(before)
	}

	if string(after) == string(before) {
		return false, nil
	}
	if write {
		if err := os.WriteFile(path, after, 0o644); err != nil {
			return false, fmt.Errorf("cannot write %s: %w", path, err)
		}
	}
	return true, nil
}

// tidyProse removes trailing whitespace from every line and gives the file a
// final newline.
//
// A carriage return before a line feed is left where it is. This repository's
// .gitattributes stores every tracked text file with LF, so a CR in a working
// copy is a fact about a checkout made before that attribute landed, not about
// the tree. Refusing it here would be the trap #17 warns about: a formatter
// that reports every file in a clean clone as wrong, so that a real defect and
// the noise become indistinguishable.
func tidyProse(b []byte) []byte {
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		cr := strings.HasSuffix(line, "\r")
		line = strings.TrimRight(line, " \t\r")
		if cr {
			line += "\r"
		}
		lines[i] = line
	}
	out := strings.Join(lines, "\n")
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return []byte(out)
}

// formattableFiles lists the Go source and tracked prose under root, relative
// to it, in a fixed order.
//
// Directories whose name starts with a dot are walked, because .github holds
// workflow files this leg reads, and .git is skipped by name because it is not
// the repository's text. testdata is skipped: the files under it are fixtures
// whose bytes are the thing being proved, and a formatter rewriting a fixture
// deletes the property it exists to carry.
func formattableFiles(root string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != root && (name == ".git" || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !isFormattable(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s, so the set of files to format is unknown: %w", root, err)
	}
	sort.Strings(found)
	return found, nil
}

// isFormattable reports whether a file name is one this leg reads.
func isFormattable(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".go" {
		return true
	}
	for _, prose := range proseExtensions {
		if ext == prose {
			return true
		}
	}
	return false
}
