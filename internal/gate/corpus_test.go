package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, body string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const goodRecord = `quantity = "neutron-lifetime"
year = 1990

[value]
published = 887.6
unit = "s"
`

func TestDecodesAsTOMLAcceptsARecord(t *testing.T) {
	path := write(t, t.TempDir(), "record/x/1990-x-01.toml", goodRecord)
	if err := decodesAsTOML(path); err != nil {
		t.Errorf("a well formed file was refused: %v", err)
	}
}

// The refusal, with the message it owes a contributor: the file and the line.
// An unterminated string is the mistake a hand transcription actually makes,
// and it is one character away from the accepted file above.
func TestDecodesAsTOMLRefusesAnUnterminatedString(t *testing.T) {
	broken := strings.Replace(goodRecord, `unit = "s"`, `unit = "s`, 1)
	path := write(t, t.TempDir(), "record/x/1990-x-01.toml", broken)

	err := decodesAsTOML(path)
	if err == nil {
		t.Fatal("an unterminated string was accepted")
	}
	if !strings.Contains(err.Error(), "1990-x-01.toml") {
		t.Errorf("the refusal does not name the file: %v", err)
	}
	if !strings.Contains(err.Error(), "line ") {
		t.Errorf("the refusal does not name a line: %v", err)
	}
}

// The near miss on the other side. A duplicated key is legal-looking and is
// refused by the format itself, which is the reason a record is stored in one
// rather than in a format that would take the second value silently.
func TestDecodesAsTOMLRefusesADuplicatedKey(t *testing.T) {
	dup := strings.Replace(goodRecord, "year = 1990\n", "year = 1990\nyear = 1991\n", 1)
	path := write(t, t.TempDir(), "record/x/1990-x-01.toml", dup)
	if err := decodesAsTOML(path); err == nil {
		t.Fatal("a file setting one key twice was accepted")
	}
}

func TestTOMLFilesFindsThemInAFixedOrder(t *testing.T) {
	root := t.TempDir()
	write(t, root, "vocabulary/b.toml", "a = 1\n")
	write(t, root, "record/a/1900-a-01.toml", "a = 1\n")
	write(t, root, "README.md", "not a record\n")

	got, err := tomlFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"record/a/1900-a-01.toml", "vocabulary/b.toml"}
	if len(got) != len(want) {
		t.Fatalf("found %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("found %v, want %v", got, want)
		}
	}
}

// A dot directory is not the corpus. Without this, a run inside a checkout
// walks .git and reports on whatever a packfile happens to be named.
func TestTOMLFilesSkipsDotDirectories(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".git/objects/x.toml", "a = 1\n")
	write(t, root, "vocabulary/b.toml", "a = 1\n")

	got, err := tomlFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "vocabulary/b.toml" {
		t.Errorf("found %v, want only vocabulary/b.toml", got)
	}
}

// Fail closed on an empty set. A tree with no TOML file at all is a run that
// examined nothing, and that must not print like a clean corpus.
func TestCorpusLegRefusesATreeWithNoTOMLAtAll(t *testing.T) {
	if _, err := corpusDecodesLeg(t.TempDir()); err == nil {
		t.Fatal("a tree with no TOML file passed the corpus leg")
	}
}

func TestCorpusLegRefusesTheTreeThatCarriesABrokenFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, "record/a/1900-a-01.toml", goodRecord)
	write(t, root, "record/a/1901-a-01.toml", strings.Replace(goodRecord, `unit = "s"`, `unit = "s`, 1))

	if _, err := corpusDecodesLeg(root); err == nil {
		t.Fatal("a tree carrying a file that does not parse passed")
	}
}

func TestCorpusLegAcceptsATreeOfRecordsThatParse(t *testing.T) {
	root := t.TempDir()
	write(t, root, "record/a/1900-a-01.toml", goodRecord)
	write(t, root, "vocabulary/a.toml", "name = \"a\"\n")

	examined, err := corpusDecodesLeg(root)
	if err != nil {
		t.Fatalf("a tree of well formed files was refused: %v", err)
	}
	if !strings.Contains(examined, "2 TOML file(s)") {
		t.Errorf("the leg did not say how many files it read: %q", examined)
	}
}
