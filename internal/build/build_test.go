package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iderex/messbuch/internal/schema"
)

// The corpus every test here builds from is written by the test rather than
// read out of this repository. A suite that builds the real record/ proves the
// state of the tree on the day it ran: it passes today because the corpus is
// empty, and it would go on passing after somebody broke the walk.

// wellFormed is one record that validates against schema version 1.
//
// It is a fixture and its numbers are invented. Nothing here is transcribed
// from a publication and nothing here may be cited.
const wellFormed = `
schema_version = 1
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

// corpus writes a tree holding the schema and the given records, and returns
// its root.
func corpus(t *testing.T, records map[string]string) string {
	t.Helper()
	root := t.TempDir()

	source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
	if err != nil {
		t.Fatalf("cannot read the schema this tree carries: %v", err)
	}
	write(t, filepath.Join(root, "schema", "record-1.toml"), string(source))

	for rel, content := range records {
		write(t, filepath.Join(root, filepath.FromSlash(rel)), content)
	}
	if err := os.MkdirAll(filepath.Join(root, "record"), 0o755); err != nil {
		t.Fatalf("cannot create record/: %v", err)
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

func load(t *testing.T, root string) *schema.Set {
	t.Helper()
	set, err := schema.Load(root)
	if err != nil {
		t.Fatalf("cannot load the schema: %v", err)
	}
	return set
}

// provenance is what every test builds against, so that no test depends on the
// checkout it happens to run in.
func provenance() Provenance {
	return Provenance{
		CorpusVersion:  "2.3.0",
		CorpusRevision: strings.Repeat("a", 40),
		CorpusState:    StateClean,
		ToolVersion:    Unreleased,
		ToolRevision:   strings.Repeat("a", 40),
		ToolState:      StateClean,
	}
}

func TestBuildCarriesTheRecordsAndTheStamp(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})

	artifact, err := Build(root, load(t, root), provenance())
	if err != nil {
		t.Fatalf("the corpus validates, so the build should not have refused: %v", err)
	}

	if len(artifact.Records) != 1 {
		t.Fatalf("built %d record(s), want 1", len(artifact.Records))
	}
	if got, want := artifact.Records[0].Path, "record/example-quantity/1900-example-01.toml"; got != want {
		t.Errorf("record path is %q, want %q", got, want)
	}
	if artifact.Stamp.SelectedCount != 1 {
		t.Errorf("selected_count is %d, want 1", artifact.Stamp.SelectedCount)
	}
	if artifact.Stamp.Command != "build" {
		t.Errorf("command is %q, want build", artifact.Stamp.Command)
	}
	if artifact.Stamp.StampVersion != StampVersion {
		t.Errorf("stamp_version is %d, want %d", artifact.Stamp.StampVersion, StampVersion)
	}
	if artifact.Stamp.CorpusVersion != "2.3.0" || artifact.Stamp.CorpusRevision != strings.Repeat("a", 40) {
		t.Errorf("the stamp does not carry the provenance it was built with: %+v", artifact.Stamp)
	}

	// The record is carried whole. A field the CSV drops has to be in the
	// authority, or the authority is a second lossy view.
	measurement, ok := artifact.Records[0].Fields["measurement"].(map[string]any)
	if !ok {
		t.Fatalf("the measurement block did not survive the build: %#v", artifact.Records[0].Fields)
	}
	published, ok := measurement["published"].(map[string]any)
	if !ok {
		t.Fatalf("the published block did not survive the build: %#v", measurement)
	}
	if _, ok := published["uncertainty"]; !ok {
		t.Error("the uncertainty components did not survive the build, so the lossless format is not lossless")
	}
}

// A clean tree carries no dirty digest and a dirty one carries a digest that
// depends on what was built. The field exists to tell two builds off two edits
// apart, so a constant would satisfy its presence and not its purpose.
func TestTheDirtyDigestIsPresentOnlyWhenDirtyAndDependsOnTheRecords(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	set := load(t, root)

	clean, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if clean.Stamp.CorpusDirtyDigest != "" {
		t.Errorf("a clean build carries a dirty digest: %q", clean.Stamp.CorpusDirtyDigest)
	}

	prov := provenance()
	prov.CorpusState = StateDirty
	first, err := Build(root, set, prov)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if first.Stamp.CorpusDirtyDigest == "" {
		t.Fatal("a dirty build carries no digest, so two builds off two edits are indistinguishable")
	}

	write(t, filepath.Join(root, "record", "example-quantity", "1901-example-01.toml"), wellFormed)
	second, err := Build(root, set, prov)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if second.Stamp.CorpusDirtyDigest == first.Stamp.CorpusDirtyDigest {
		t.Error("the digest did not move when the records did, so it fingerprints nothing")
	}
}

// The refusal that matters most: an artifact is data wearing a stamp, and a
// stamp on records nothing accepted is worse than no artifact.
func TestBuildRefusesACorpusThatDoesNotValidate(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": strings.Replace(wellFormed, `status = "active"`, `statuss = "active"`, 1),
	})

	_, err := Build(root, load(t, root), provenance())
	if err == nil {
		t.Fatal("built an artifact out of a record the validator refuses")
	}
	if !strings.Contains(err.Error(), "unknown-field") {
		t.Errorf("the refusal does not say what was wrong with the corpus: %v", err)
	}
}

// A directory the schema excludes is counted rather than silently passed over,
// so a build says what it did not read beside what it did.
func TestAnExcludedDirectoryIsCountedAndNotBuilt(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
		"record/_example/1900-example-01.toml":         "this is not a record and nothing reads it\n",
	})

	artifact, err := Build(root, load(t, root), provenance())
	if err != nil {
		t.Fatalf("an excluded directory is not a defect: %v", err)
	}
	if len(artifact.Records) != 1 {
		t.Fatalf("built %d record(s), want 1", len(artifact.Records))
	}
	if got := artifact.Stamp.Excluded[ExcludedDirectory]; got != 1 {
		t.Errorf("excluded count is %d, want 1; a build that reads less than the whole corpus has to say so", got)
	}
}

func TestRecordsComeOutInPathOrder(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1902-example-01.toml": wellFormed,
		"record/example-quantity/1900-example-01.toml": wellFormed,
		"record/another-quantity/1901-example-01.toml": wellFormed,
	})

	artifact, err := Build(root, load(t, root), provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := []string{
		"record/another-quantity/1901-example-01.toml",
		"record/example-quantity/1900-example-01.toml",
		"record/example-quantity/1902-example-01.toml",
	}
	for i, record := range artifact.Records {
		if record.Path != want[i] {
			t.Fatalf("record %d is %s, want %s; an order that comes from the walk is an artifact that changes without the corpus changing", i, record.Path, want[i])
		}
	}
}

// A corpus that cannot be read is not an empty corpus, and the difference is
// the whole reason a green line means anything.
func TestBuildFailsClosedWhenTheCorpusCannotBeRead(t *testing.T) {
	root := t.TempDir()
	source, err := os.ReadFile(filepath.Join("..", "..", "schema", "record-1.toml"))
	if err != nil {
		t.Fatalf("cannot read the schema this tree carries: %v", err)
	}
	write(t, filepath.Join(root, "schema", "record-1.toml"), string(source))

	if _, err := Build(root, load(t, root), provenance()); err == nil {
		t.Fatal("a tree with no record directory built an empty artifact instead of refusing")
	}
}

func TestOutputsAreUnderTheBuildDirectory(t *testing.T) {
	outputs := Outputs()
	if len(outputs) != 3 {
		t.Fatalf("the build writes %d path(s), want 3", len(outputs))
	}
	for _, path := range outputs {
		if !strings.HasPrefix(path, Dir+"/") {
			t.Errorf("%s is not under %s/, so the leg that refuses a tracked artifact would not reach it", path, Dir)
		}
	}
}

func TestWriteAllWritesEveryOutput(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	set := load(t, root)

	artifact, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	written, err := WriteAll(root, artifact, set.Any())
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if len(written) != len(Outputs()) {
		t.Fatalf("wrote %d file(s), want %d", len(written), len(Outputs()))
	}
	for _, rel := range written {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was reported written and is not there: %v", rel, err)
		}
	}
}
