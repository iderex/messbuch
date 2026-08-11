package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iderex/messbuch/internal/compat"
)

// The leg is exercised over trees this file writes rather than over this
// repository's own frozen corpus, for the reason the package it calls gives:
// the real corpus proves the state of the tree on the day the suite ran.

func compatTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()

	source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
	if err != nil {
		t.Fatalf("cannot read the schema this tree carries: %v", err)
	}
	writeFile(t, filepath.Join(root, "schema", "record-1.toml"), string(source))
	for name, content := range files {
		writeFile(t, filepath.Join(root, filepath.FromSlash(compat.Dir), "schema-1", name), content)
	}
	if _, err := compat.FreezeAll(root); err != nil {
		t.Fatalf("cannot freeze the fixture tree: %v", err)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("cannot create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("cannot write %s: %v", path, err)
	}
}

// A record that validates against schema version 1. INVENTED, and nothing in it
// may be cited.
const compatRecord = `schema_version = 1
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

func TestSchemaCompatibilityPassesAndSaysWhatItRead(t *testing.T) {
	root := compatTree(t, map[string]string{"accepted.toml": compatRecord})

	examined, err := schemaCompatibilityLeg(root)
	if err != nil {
		t.Fatalf("a tree frozen a moment ago was refused: %v", err)
	}
	if !strings.Contains(examined, "schema version 1") {
		t.Errorf("the leg does not say which versions it read: %q", examined)
	}
	if !strings.Contains(examined, "1 file(s)") {
		t.Errorf("the leg does not say how many files it read: %q", examined)
	}
}

func TestSchemaCompatibilityRefusesAMovedReading(t *testing.T) {
	root := compatTree(t, map[string]string{"accepted.toml": compatRecord})
	writeFile(t, filepath.Join(root, filepath.FromSlash(compat.Dir), "schema-1", "accepted.toml"),
		strings.Replace(compatRecord, "value = 1.5\nunit = \"s\"", "value = 1500.0\nunit = \"ms\"", 1))

	_, err := schemaCompatibilityLeg(root)
	if err == nil {
		t.Fatal("a frozen record started meaning something else and the leg passed")
	}
	if !strings.Contains(err.Error(), "accepted.toml") {
		t.Errorf("the refusal does not name the file: %v", err)
	}
	if !strings.Contains(err.Error(), "changelog") {
		t.Errorf("the refusal does not say what an argued break looks like: %v", err)
	}
}

func TestSchemaCompatibilityFailsClosedWithNothingFrozen(t *testing.T) {
	root := t.TempDir()
	source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
	if err != nil {
		t.Fatalf("cannot read the schema: %v", err)
	}
	writeFile(t, filepath.Join(root, "schema", "record-1.toml"), string(source))

	if _, err := schemaCompatibilityLeg(root); err == nil {
		t.Fatal("a tree with nothing frozen passed, so the leg reports green having examined nothing")
	}
}
