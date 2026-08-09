package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tree writes files into a fresh directory and returns its path. The contents
// are written from escapes rather than from raw literals wherever a specific
// byte is the point, because a literal would be normalised by this
// repository's own text attributes on the way into git and the byte the
// fixture exists to prove would be gone.
//
// A tree gets a module and one already-formatted package unless the caller
// wrote its own, because the lint half runs go vet over the tree and a
// directory with no module would fail for a reason no fixture here is about.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if _, ours := files["go.mod"]; !ours {
		files["go.mod"] = "module messbuch.example/fixture\n\ngo 1.26.0\n"
		files["clean.go"] = "package fixture\n"
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func read(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A trailing space is invisible in every editor and shows up in every diff.
// The fix has to remove it, and the check has to name the file.
func TestFormatLegTrailingSpaceIsRefusedAndTheFixRemovesIt(t *testing.T) {
	root := tree(t, map[string]string{
		"docs/note.md": "A line with one trailing space\x20\nand a clean one\n",
	})

	if _, err := formatAndLintLeg(root); err == nil {
		t.Fatal("a trailing space passed")
	} else if !strings.Contains(err.Error(), "docs/note.md") {
		t.Errorf("the refusal does not name the file: %v", err)
	}

	changed, err := Reformat(root)
	if err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if len(changed) != 1 || changed[0] != "docs/note.md" {
		t.Fatalf("the fix changed %v, want docs/note.md alone", changed)
	}
	if got, want := read(t, root, "docs/note.md"), "A line with one trailing space\nand a clean one\n"; got != want {
		t.Errorf("after the fix the file is %q, want %q", got, want)
	}
}

// A file with no final newline makes the next line added to it show up as two
// changed lines instead of one.
func TestFormatLegMissingFinalNewlineIsRefusedAndTheFixAddsIt(t *testing.T) {
	root := tree(t, map[string]string{"docs/note.md": "one line and no newline"})

	if _, err := formatAndLintLeg(root); err == nil {
		t.Fatal("a file with no final newline passed")
	}
	if _, err := Reformat(root); err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if got, want := read(t, root, "docs/note.md"), "one line and no newline\n"; got != want {
		t.Errorf("after the fix the file is %q, want %q", got, want)
	}
}

// The Go half. gofmt's own opinion, reached through the package gofmt calls,
// so the check cannot drift into a second opinion about what formatted means.
func TestFormatLegUnformattedGoIsRefusedAndTheFixFormatsIt(t *testing.T) {
	root := tree(t, map[string]string{
		"go.mod":  "module messbuch.example/fixture\n\ngo 1.26.0\n",
		"bad.go":  "package fixture\n\nfunc  Wrong( ) int {\nreturn 1\n}\n",
		"fine.go": "package fixture\n\n// Right is already what gofmt writes.\nfunc Right() int { return 2 }\n",
	})

	if _, err := formatAndLintLeg(root); err == nil {
		t.Fatal("unformatted Go passed")
	} else if !strings.Contains(err.Error(), "bad.go") {
		t.Errorf("the refusal does not name the file: %v", err)
	} else if strings.Contains(err.Error(), "fine.go") {
		t.Errorf("the refusal names a file that was already formatted: %v", err)
	}

	changed, err := Reformat(root)
	if err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if len(changed) != 1 || changed[0] != "bad.go" {
		t.Fatalf("the fix changed %v, want bad.go alone", changed)
	}
	if got := read(t, root, "bad.go"); !strings.Contains(got, "func Wrong() int {") {
		t.Errorf("the fix did not produce gofmt's bytes: %q", got)
	}
}

// The refusal carries the command that repairs it. A check that says a file is
// wrong without saying what produces the right bytes leaves every contributor
// guessing at the difference between their formatter and this one.
func TestFormatLegRefusalNamesTheFixCommand(t *testing.T) {
	root := tree(t, map[string]string{"docs/note.md": "trailing\x20\n"})
	_, err := formatAndLintLeg(root)
	if err == nil {
		t.Fatal("the fixture passed")
	}
	if !strings.Contains(err.Error(), "go run . fmt") {
		t.Errorf("the refusal does not name the fix command: %v", err)
	}
}

// The check and the fix agree. This is the property the whole shape exists
// for: a tree the fix has been run on passes, with nothing left over for a
// second run to find.
func TestFormatLegTreeTheFixHasRunOnPasses(t *testing.T) {
	root := tree(t, map[string]string{
		"go.mod":       "module messbuch.example/fixture\n\ngo 1.26.0\n",
		"bad.go":       "package fixture\n\nfunc  Wrong( ) int {\n\treturn 1\n}\n",
		"docs/note.md": "trailing\x20\x20\nno newline at the end",
	})

	if _, err := Reformat(root); err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if _, err := formatAndLintLeg(root); err != nil {
		t.Fatalf("a tree the fix ran on was refused: %v", err)
	}
	again, err := Reformat(root)
	if err != nil {
		t.Fatalf("the second fix returned %v", err)
	}
	if len(again) != 0 {
		t.Errorf("a second run of the fix still changed %v", again)
	}
}

// The line-ending trap #17 asks to be settled. A carriage return before a line
// feed is a fact about somebody's checkout, and a formatter that reported
// every file in a clean clone as wrong would drown a real defect in noise.
// .gitattributes is what settles the stored bytes; this leg does not.
func TestFormatLegCarriageReturnIsNotADefect(t *testing.T) {
	root := tree(t, map[string]string{"docs/note.md": "one line\r\nanother\r\n"})
	if _, err := formatAndLintLeg(root); err != nil {
		t.Fatalf("a checkout with CRLF was refused as misformatted: %v", err)
	}
	if got, want := read(t, root, "docs/note.md"), "one line\r\nanother\r\n"; got != want {
		t.Errorf("the file was rewritten to %q", got)
	}
}

// A trailing space before a carriage return is still a trailing space, and it
// is the one a naive rule misses.
func TestFormatLegTrailingSpaceBeforeACarriageReturnIsRefused(t *testing.T) {
	root := tree(t, map[string]string{"docs/note.md": "one line\x20\r\n"})
	if _, err := formatAndLintLeg(root); err == nil {
		t.Fatal("a trailing space before a carriage return passed")
	}
	if _, err := Reformat(root); err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if got, want := read(t, root, "docs/note.md"), "one line\r\n"; got != want {
		t.Errorf("after the fix the file is %q, want %q", got, want)
	}
}

// Source that does not parse is the build leg's refusal. Two refusals for one
// defect is one refusal too many, and the second one says the wrong thing.
func TestFormatLegGoFileThatDoesNotParseIsNotItsRefusal(t *testing.T) {
	root := tree(t, map[string]string{
		"go.mod":  "module messbuch.example/fixture\n\ngo 1.26.0\n",
		"open.go": "package fixture\n\nfunc Open() {\n",
	})
	if _, err := formatAndLintLeg(root); err != nil && strings.Contains(err.Error(), "not formatted") {
		t.Errorf("a parse failure was reported as a formatting defect: %v", err)
	}
}

// Fixture bytes are the thing being proved elsewhere in this package, and a
// formatter rewriting a fixture deletes the property it carries.
func TestFormatLegFilesUnderTestdataAreNotFormatted(t *testing.T) {
	root := tree(t, map[string]string{
		"testdata/broken/note.md": "a fixture with a trailing space\x20\n",
		"docs/note.md":            "clean\n",
	})
	if _, err := formatAndLintLeg(root); err != nil {
		t.Fatalf("a fixture under testdata was refused: %v", err)
	}
	changed, err := Reformat(root)
	if err != nil {
		t.Fatalf("the fix returned %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("the fix rewrote %v under testdata", changed)
	}
}

// LICENSE and DCO are copies of texts that are not this project's to
// normalise, and a check that rewrote one would be changing a legal
// instrument.
func TestFormatLegLicenseAndSignOffTextAreLeftAlone(t *testing.T) {
	root := tree(t, map[string]string{
		"LICENSE":      "a license line with a trailing space\x20\n",
		"DCO":          "a certificate line with a trailing space\x20\n",
		"docs/note.md": "clean\n",
	})
	if _, err := formatAndLintLeg(root); err != nil {
		t.Fatalf("LICENSE or DCO was refused: %v", err)
	}
}

// Fails closed. A tree with nothing to read is a leg that examined nothing,
// and that is not the same statement as a clean tree.
func TestFormatLegTreeWithNothingToFormatIsARefusal(t *testing.T) {
	if _, err := formatAndLintLeg(t.TempDir()); err == nil {
		t.Fatal("a tree with no source and no prose passed")
	}
}

// The tree this repository actually is. A leg green only on its own fixtures
// says nothing about the mainline.
func TestTheTreeIsFormattedAsWritten(t *testing.T) {
	examined, err := formatAndLintLeg("../..")
	if err != nil {
		t.Fatalf("this repository is not formatted as it writes: %v", err)
	}
	if !strings.Contains(examined, "go vet") {
		t.Errorf("the result does not say the lint half ran: %q", examined)
	}
}
