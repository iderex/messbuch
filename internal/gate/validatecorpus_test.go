package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyInto puts this repository's own schema directory under a temporary root,
// so a test can give the leg a corpus without giving it a second schema.
func copyInto(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("..", "..", "schema"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		body, err := os.ReadFile(filepath.Join("..", "..", "schema", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "schema", entry.Name()), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestValidateCorpusLegReadsThisRepository(t *testing.T) {
	examined, err := validateCorpusLeg(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("this repository does not pass its own structural leg: %v", err)
	}
	if !strings.Contains(examined, "record(s)") {
		t.Fatalf("the leg has to say how many records it read: %q", examined)
	}
}

func TestValidateCorpusLegRefusesARecordThatIsNotOne(t *testing.T) {
	root := t.TempDir()
	copyInto(t, root)
	if err := os.MkdirAll(filepath.Join(root, "record", "example-quantity"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "schema_version = 1\nstatus = \"active\"\n"
	if err := os.WriteFile(filepath.Join(root, "record", "example-quantity", "1900-example-01.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := validateCorpusLeg(root)
	if err == nil {
		t.Fatalf("a record carrying two fields of the set the schema requires was accepted")
	}
	if !strings.Contains(err.Error(), "go run . refusals") {
		t.Fatalf("the refusal has to name the command that lists what it can refuse: %v", err)
	}
}

// Fails closed. A tree with no schema is not a tree whose corpus is clean.
func TestValidateCorpusLegRefusesATreeWithNoSchema(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "record"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := validateCorpusLeg(root); err == nil {
		t.Fatalf("a corpus validated against no schema at all was reported clean")
	}
}

func TestTheGateDeclaresTheValidateCorpusLeg(t *testing.T) {
	only, err := Only(Legs(), "validate-corpus")
	if err != nil {
		t.Fatal(err)
	}
	if only[0].Run == nil {
		t.Fatalf("the leg is declared and not built")
	}
	if only[0].Limits == "" {
		t.Fatalf("an assurance with an unstated edge is worse than a narrower one that says where it stops")
	}
}

func TestEveryRefusalSiteCarriesALine(t *testing.T) {
	seen := map[string]bool{}
	for _, site := range Sites() {
		if site.ID == "" || site.Refuses == "" {
			t.Fatalf("a catalogue entry with nothing to print: %#v", site)
		}
		if seen[site.ID] {
			t.Fatalf("%s is in the catalogue twice", site.ID)
		}
		seen[site.ID] = true
	}
	if len(seen) == 0 {
		t.Fatalf("the catalogue is empty, so the command that prints it says nothing")
	}
}
