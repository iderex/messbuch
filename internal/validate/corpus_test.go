package validate

import (
	"os"
	"path/filepath"
	"testing"
)

// tree writes a corpus under a temporary root and returns the root.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestStructureCorpusAcceptsATreeOfWellFormedRecords(t *testing.T) {
	root := tree(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
		"record/example-quantity/1901-example-01.toml": wellFormed,
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.Records != 2 || len(report.Refusals) != 0 {
		t.Fatalf("read %d record(s) with %d refusal(s): %v", report.Records, len(report.Refusals), report.Refusals)
	}
}

// The path is the record's identity, so a file in the wrong place claims an
// identity nothing can read. This is the one refusal a walk earns rather than
// a file's contents.
func TestStructureCorpusRefusesAFileInTheWrongPlace(t *testing.T) {
	root := tree(t, map[string]string{
		"record/example-quantity/notes.md": "not a record\n",
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Refusals) != 1 || report.Refusals[0].Site != "record-in-the-wrong-place" {
		t.Fatalf("expected one path refusal, got %v", report.Refusals)
	}
	if report.Records != 0 {
		t.Fatalf("a file that is not at a record's path was read as a record")
	}
}

// The near miss for the path rule: the same name one character from legal.
// Two digits where the schema's pattern wants four is the mistake a
// transcriber working through one decade makes.
func TestStructureCorpusRefusesATwoDigitYearInAFileName(t *testing.T) {
	root := tree(t, map[string]string{
		"record/example-quantity/00-example-01.toml": wellFormed,
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Refusals) != 1 || report.Refusals[0].Site != "record-in-the-wrong-place" {
		t.Fatalf("expected the path refusal alone, got %v", report.Refusals)
	}
}

func TestStructureCorpusAcceptsThePathTheSchemaFixes(t *testing.T) {
	root := tree(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Refusals) != 0 {
		t.Fatalf("a legal path was refused: %v", report.Refusals)
	}
}

// A directory whose name begins with the excluded prefix is not a quantity,
// and the count of what was not read is reported rather than left to be
// inferred from silence.
func TestStructureCorpusCountsWhatItDidNotRead(t *testing.T) {
	root := tree(t, map[string]string{
		"record/_example/1900-example-01.toml":         "anything at all, and not TOML\n",
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.Records != 1 {
		t.Fatalf("expected one record, got %d", report.Records)
	}
	if report.Skipped != 1 || len(report.Excluded) != 1 || report.Excluded[0] != "record/_example" {
		t.Fatalf("the excluded directory is not reported: %d skipped, %v", report.Skipped, report.Excluded)
	}
	if len(report.Refusals) != 0 {
		t.Fatalf("an excluded file was refused: %v", report.Refusals)
	}
}

// Fails closed. A tree with no record directory is a corpus that is unknown
// rather than a corpus that is empty and clean.
func TestStructureCorpusRefusesATreeWithNoRecordDirectory(t *testing.T) {
	if _, err := Corpus(t.TempDir(), load(t)); err == nil {
		t.Fatalf("a tree with no record directory was read as an empty corpus")
	}
}

// An empty corpus is a fact rather than a failure. There is no record in this
// repository yet, and refusing on that would red every run until the first
// series lands.
func TestStructureCorpusReadsAnEmptyCorpusAsEmpty(t *testing.T) {
	root := tree(t, map[string]string{"record/.keep": ""})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatalf("an empty corpus is not a failure: %v", err)
	}
	if report.Records != 0 {
		t.Fatalf("expected no records, got %d", report.Records)
	}
}

func TestStructureCorpusReportsEveryBadRecordRatherThanTheFirst(t *testing.T) {
	broken := replace(t, `quantity = "example-quantity"`, "quantity = 4")
	root := tree(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": broken,
		"record/example-quantity/1901-example-01.toml": broken,
	})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Refusals) != 2 {
		t.Fatalf("expected both files reported, got %v", report.Refusals)
	}
	if report.Refusals[0].File > report.Refusals[1].File {
		t.Fatalf("refusals are not in a fixed order: %v", report.Refusals)
	}
}

// corpusFixtureReachesThePathRefusal is the accounting's evidence that the one
// site no record's contents can produce is reached by a fixture here.
func corpusFixtureReachesThePathRefusal(t *testing.T) bool {
	t.Helper()
	root := tree(t, map[string]string{"record/example-quantity/notes.md": "not a record\n"})
	report, err := Corpus(root, load(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range report.Refusals {
		if r.Site == "record-in-the-wrong-place" {
			return true
		}
	}
	return false
}
