package build

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// The convenience view's whole obligation is that it says what it dropped, so
// the first thing to check is that it drops what it claims to and that the two
// lists together are the schema rather than a hand-kept subset of it.
func TestColumnsAndDroppedPartitionTheSchema(t *testing.T) {
	root := corpus(t, map[string]string{})
	s := load(t, root).Any()

	columns := Columns(s)
	dropped := Dropped(s)
	if len(columns)+len(dropped) != len(s.Fields) {
		t.Fatalf("%d column(s) plus %d dropped field(s) is not the schema's %d field(s); a field that is in neither list is one nothing says was lost",
			len(columns), len(dropped), len(s.Fields))
	}

	seen := map[string]bool{}
	for _, name := range append(append([]string{}, columns...), dropped...) {
		if seen[name] {
			t.Errorf("%s is both a column and dropped", name)
		}
		seen[name] = true
	}

	// The uncertainty representation is the reason this format is called
	// lossy, so it is the one field whose place is worth naming outright.
	if !contains(dropped, "measurement.published.uncertainty") {
		t.Error("the uncertainty components are not in the dropped list, so either the CSV carries them or it loses them silently")
	}
	if !contains(columns, "measurement.published.value") {
		t.Error("the published value is not a column, so the convenience view carries no number")
	}
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

func TestTheConvenienceViewCarriesTheStampAndNamesWhatItDropped(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	set := load(t, root)
	artifact, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var out bytes.Buffer
	if err := WriteCSV(&out, artifact, set.Any()); err != nil {
		t.Fatalf("write: %v", err)
	}
	text := out.String()

	for _, want := range []string{
		"# corpus_revision: " + artifact.Stamp.CorpusRevision,
		"# corpus_version: 2.3.0",
		"# corpus_state: clean",
		"# command: build",
		"# selected_count: 1",
		"# dropped: ",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the convenience view does not carry %q, so a copy of it cannot say what it is", want)
		}
	}
	if !strings.Contains(text, "measurement.published.uncertainty") {
		t.Error("the dropped line does not name the uncertainty, which is the field this format exists to admit it cannot carry")
	}

	// Everything before the header row is a comment line, so a reader who
	// strips them reaches a table rather than half a stamp.
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "schema_version,") {
			break
		}
		if line != "" && !strings.HasPrefix(line, "#") {
			t.Fatalf("a line before the header is neither blank nor a comment: %q", line)
		}
	}
}

func TestTheConvenienceViewPutsTheValuesInTheRightCells(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	set := load(t, root)
	artifact, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var out bytes.Buffer
	if err := WriteCSV(&out, artifact, set.Any()); err != nil {
		t.Fatalf("write: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(out.String()))
	reader.Comment = '#'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("the convenience view does not parse as CSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("read %d row(s), want a header and one record", len(rows))
	}

	cells := map[string]string{}
	for i, name := range rows[0] {
		cells[name] = rows[1][i]
	}
	for name, want := range map[string]string{
		"schema_version":              "1",
		"status":                      "active",
		"measurement.quantity":        "example-quantity",
		"measurement.published.value": "1.5",
		"publication.date":            "1900",
		"group.id":                    "example-group",
	} {
		if cells[name] != want {
			t.Errorf("%s is %q, want %q", name, cells[name], want)
		}
	}
	// A field the record does not carry is an empty cell rather than a word a
	// reader could take for a value.
	if cells["superseded_by"] != "" {
		t.Errorf("a field the record does not carry came out as %q", cells["superseded_by"])
	}
}

// The lossless artifact is the authority, so its bytes have to be a function of
// what was built and of nothing else. Two writes of one artifact are the
// cheapest form of that question, and #28 is where the whole build is measured
// against it.
func TestTheLosslessArtifactIsTheSameBytesTwice(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
		"record/another-quantity/1901-example-01.toml": wellFormed,
	})
	set := load(t, root)
	artifact, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var first, second bytes.Buffer
	if err := WriteJSON(&first, artifact); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := WriteJSON(&second, artifact); err != nil {
		t.Fatalf("write: %v", err)
	}
	if first.String() != second.String() {
		t.Fatal("two writes of one artifact produced different bytes, so nothing downstream can compare a rebuild against anything")
	}
}

func TestTheSidecarCarriesTheSameStampAndTheSameDroppedList(t *testing.T) {
	root := corpus(t, map[string]string{
		"record/example-quantity/1900-example-01.toml": wellFormed,
	})
	set := load(t, root)
	artifact, err := Build(root, set, provenance())
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var out bytes.Buffer
	if err := WriteCSVStamp(&out, artifact, set.Any()); err != nil {
		t.Fatalf("write: %v", err)
	}

	var sidecar CSVStamp
	if err := json.Unmarshal(out.Bytes(), &sidecar); err != nil {
		t.Fatalf("the sidecar does not parse as JSON, which is the one thing it exists to be: %v", err)
	}
	if !reflect.DeepEqual(sidecar.Stamp, artifact.Stamp) {
		t.Errorf("the sidecar's stamp is not the artifact's:\n%+v\n%+v", sidecar.Stamp, artifact.Stamp)
	}
	if !sidecar.Lossy || sidecar.Of != CSVName {
		t.Errorf("the sidecar does not say which file it is about or that the file is lossy: %+v", sidecar)
	}
	if len(sidecar.Dropped) != len(Dropped(set.Any())) {
		t.Errorf("the sidecar names %d dropped field(s) and the format drops %d", len(sidecar.Dropped), len(Dropped(set.Any())))
	}
}

// The guard inside the convenience view: a value that is not one cell is an
// error rather than something %v turns into text a reader would take for the
// value. It bites here, on the kind of value a block would produce.
func TestACellRefusesAValueItCannotHold(t *testing.T) {
	for _, value := range []any{
		map[string]any{"start": "1900"},
		[]any{"one", "two"},
	} {
		if _, err := cell(value); err == nil {
			t.Errorf("%T came out as a cell, so a block would be printed as though it were a value", value)
		}
	}

	// And it does hold the kinds a record actually writes.
	for _, value := range []any{nil, "text", true, int64(1), 1.5} {
		if _, err := cell(value); err != nil {
			t.Errorf("%T is a value a record writes and the cell refused it: %v", value, err)
		}
	}
}
