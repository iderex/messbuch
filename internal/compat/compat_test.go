package compat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every test here builds its own tree. A suite that reads this repository's own
// frozen corpus would report the state of the tree on the day it ran, and this
// package's whole subject is what happens when that state moves.

// legal is a record that validates against schema version 1.
//
// INVENTED. Nothing in it is transcribed from any publication and none of it
// may be cited.
const legal = `schema_version = 1
status = "active"
blinding = "not-stated"

[measurement]
quantity = "example-quantity"
normalization_status = "normalized"
definition_epoch = "example-epoch"

[measurement.published]
value = 1.5
unit = "example unit"
uncertainty_status = "reported"

[[measurement.published.uncertainty]]
component = "statistical"
plus = 0.1
minus = 0.1
coverage = "k=1"
interval_kind = "frequentist"

[measurement.normalized]
value = 1.5
unit = "s"

[[measurement.normalized.uncertainty]]
component = "statistical"
plus = 0.1
minus = 0.1
coverage = "k=1"
interval_kind = "frequentist"

[measurement.conversion]
factor = 1.0
exact = true
factor_source = "invented for a fixture, and no such authority exists"

[publication]
date = "1900"
data_taken_status = "none-in-source"

[method]
technique = "example-technique"

[group]
id = "example-group"

[source]
statement_kind = "primary-result"
directness = "primary"
confirmation = "unconfirmed"

[[source.identifier]]
scheme = "doi"
value = "10.0000/invented-for-a-fixture"

[source.locator]
page = "1"
`

// tree writes a repository holding the schema and one frozen corpus, and
// freezes it, so that every test starts from a tree the check passes over.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()

	source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
	if err != nil {
		t.Fatalf("cannot read the schema this tree carries: %v", err)
	}
	write(t, filepath.Join(root, "schema", "record-1.toml"), string(source))

	for name, content := range files {
		write(t, filepath.Join(root, filepath.FromSlash(Dir), "schema-1", name), content)
	}
	if _, err := FreezeAll(root); err != nil {
		t.Fatalf("cannot freeze the fixture tree: %v", err)
	}
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("cannot create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("cannot write %s: %v", path, err)
	}
}

func fixture(t *testing.T, root, name string) string {
	t.Helper()
	return filepath.Join(root, filepath.FromSlash(Dir), "schema-1", name)
}

func TestAFrozenCorpusThatStillReadsTheSameWayPasses(t *testing.T) {
	root := tree(t, map[string]string{"accepted.toml": legal})

	report, err := Check(root)
	if err != nil {
		t.Fatalf("a tree frozen a moment ago disagrees with itself: %v", err)
	}
	if len(report.Differences) > 0 {
		t.Fatalf("differences over a tree nothing changed:\n%s", strings.Join(report.Differences, "\n"))
	}
	if len(report.Versions) != 1 || report.Versions[0].Schema != 1 {
		t.Fatalf("read %d version(s), want schema version 1 alone", len(report.Versions))
	}
	if len(report.Versions[0].Files) != 1 {
		t.Fatalf("read %d file(s), want 1", len(report.Versions[0].Files))
	}
}

// The refusal this leg exists for. The bytes do not move, the schema does, and
// the corpus still validates: what changed is what the field means.
func TestAChangedInterpretationOfAnExistingFieldIsRefused(t *testing.T) {
	root := tree(t, map[string]string{"accepted.toml": legal})

	// The file is unchanged. Its reading moves because the value it carries is
	// read as something else, which is exactly the silent half of this issue:
	// nothing refuses the corpus and every number derived from it has moved.
	write(t, fixture(t, root, "accepted.toml"), strings.Replace(legal, "value = 1.5\nunit = \"s\"", "value = 1500.0\nunit = \"ms\"", 1))

	report, err := Check(root)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(report.Differences) == 0 {
		t.Fatal("the normalized value and unit of a frozen record moved and nothing said so")
	}
	if !strings.Contains(report.Differences[0], "reads differently now") {
		t.Errorf("the difference does not say what kind of change it is: %s", report.Differences[0])
	}
	if !strings.Contains(report.Differences[0], "accepted.toml") {
		t.Errorf("the difference does not name the file: %s", report.Differences[0])
	}
}

// A record whose refusals move is the loud half, and it is frozen too: a field
// that stops being refused is a corpus that starts being accepted.
func TestAChangedRefusalIsRefused(t *testing.T) {
	root := tree(t, map[string]string{
		"refused.toml": strings.Replace(legal, `blinding = "not-stated"`, `blindng = "not-stated"`, 1),
	})

	frozen, err := frozenReadings(root, 1)
	if err != nil {
		t.Fatalf("read the frozen readings: %v", err)
	}
	if len(frozen["refused.toml"].Refusals) == 0 {
		t.Fatal("the fixture was frozen as accepted, so this test is not about what it says it is")
	}

	// The misspelling is corrected, so the file stops being refused.
	write(t, fixture(t, root, "refused.toml"), legal)

	report, err := Check(root)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(report.Differences) == 0 {
		t.Fatal("a frozen file stopped being refused and nothing said so")
	}
	if !strings.Contains(report.Differences[0], "is now accepted") {
		t.Errorf("the difference does not say the file is now accepted: %s", report.Differences[0])
	}
}

// The check compares in both directions, because one that only compares the
// files it finds in both places is silenced by deleting a fixture.
func TestAFrozenFileThatLeavesTheCorpusIsRefused(t *testing.T) {
	root := tree(t, map[string]string{"accepted.toml": legal})

	if err := os.Remove(fixture(t, root, "accepted.toml")); err != nil {
		t.Fatalf("cannot remove the fixture: %v", err)
	}
	// A second file, so the version directory is not empty and the refusal
	// under test is the missing one rather than the empty one.
	write(t, fixture(t, root, "other.toml"), legal)

	report, err := Check(root)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	joined := strings.Join(report.Differences, "\n")
	if !strings.Contains(joined, "no longer in the corpus") {
		t.Errorf("deleting a frozen fixture was not refused: %q", joined)
	}
	if !strings.Contains(joined, "has no frozen reading") {
		t.Errorf("adding a fixture without freezing it was not refused: %q", joined)
	}
}

// It fails closed. Each of these would otherwise be a green line about nothing.
func TestCheckFailsClosed(t *testing.T) {
	t.Run("no frozen corpus at all", func(t *testing.T) {
		root := t.TempDir()
		source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
		if err != nil {
			t.Fatalf("cannot read the schema: %v", err)
		}
		write(t, filepath.Join(root, "schema", "record-1.toml"), string(source))

		if _, err := Check(root); err == nil {
			t.Fatal("a tree with nothing frozen passed, so the leg would report green having examined nothing")
		}
	})

	t.Run("a frozen version holding no file", func(t *testing.T) {
		root := tree(t, map[string]string{"accepted.toml": legal})
		if err := os.Remove(fixture(t, root, "accepted.toml")); err != nil {
			t.Fatalf("cannot remove the fixture: %v", err)
		}
		if _, err := Check(root); err == nil {
			t.Fatal("a frozen version with no file passed, and it proves nothing about that version")
		}
	})

	t.Run("readings that do not parse", func(t *testing.T) {
		root := tree(t, map[string]string{"accepted.toml": legal})
		write(t, filepath.Join(root, filepath.FromSlash(Dir), "schema-1", ReadingsName), "not json\n")
		if _, err := Check(root); err == nil {
			t.Fatal("a readings file that does not parse was read as agreeing with everything")
		}
	})

	t.Run("a directory that does not name a version", func(t *testing.T) {
		root := tree(t, map[string]string{"accepted.toml": legal})
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(Dir), "schema-not-a-number"), 0o755); err != nil {
			t.Fatalf("cannot create the directory: %v", err)
		}
		if _, err := Check(root); err == nil {
			t.Fatal("a frozen directory naming no schema version passed, so nothing says what its files were written against")
		}
	})
}

// Freezing is what a person runs after an argued break, so it has to produce
// something Check then agrees with rather than a file somebody edits by hand.
func TestFreezingProducesReadingsTheCheckAgreesWith(t *testing.T) {
	root := tree(t, map[string]string{"accepted.toml": legal})

	written, err := FreezeAll(root)
	if err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if len(written) != 1 || !strings.HasSuffix(written[0], ReadingsName) {
		t.Fatalf("froze %v, want one readings file", written)
	}

	report, err := Check(root)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(report.Differences) > 0 {
		t.Fatalf("a freeze produced readings its own check disagrees with:\n%s", strings.Join(report.Differences, "\n"))
	}
}
