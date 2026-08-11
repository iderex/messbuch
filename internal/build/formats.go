package build

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/iderex/messbuch/internal/schema"
)

// WriteJSON writes the lossless artifact.
//
// This format is the authority. Every field of every record is carried under
// the name the record wrote it under, so a consumer reading this file and a
// contributor reading the TOML are reading the same thing.
//
// The bytes are a function of the artifact and of nothing else. Map keys are
// sorted by the encoder, records are in path order, and the stamp holds no
// clock, so two builds of one revision produce one file. That property is what
// #28 is going to measure and it is cheaper to have than to add later.
func WriteJSON(w io.Writer, a *Artifact) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a); err != nil {
		return fmt.Errorf("cannot write the lossless artifact: %w", err)
	}
	return nil
}

// Columns are the CSV's columns, in the order the schema declares them.
//
// Derived from the schema rather than listed here, for the reason the rest of
// this tree gives: a list of field names in Go drifts against the file that
// decides them, and the drift is invisible because a column that stopped
// existing looks exactly like one that was never wanted. A field added to the
// schema arrives in this format without a line of Go moving, and a field that
// cannot fit a cell arrives in Dropped instead.
func Columns(s *schema.Schema) []string {
	var out []string
	for _, f := range s.Fields {
		if flat(s, f.Type) {
			out = append(out, f.Path)
		}
	}
	return out
}

// Dropped is every top-level field the CSV cannot carry, in the same order.
//
// It is written into both the CSV's own header lines and the sidecar, because
// the requirement is that a lossy format says what it dropped rather than
// dropping it silently. A reader who does not know a column is missing cannot
// know to go and find it.
func Dropped(s *schema.Schema) []string {
	var out []string
	for _, f := range s.Fields {
		if !flat(s, f.Type) {
			out = append(out, f.Path)
		}
	}
	return out
}

// flat reports whether a declared type fits in one cell.
//
// A list does not, a composite does not, and a scalar declaring members is a
// block in the file however it is named in the schema, so it does not either.
// Anything left is a string, a number, a boolean or a pattern-checked string,
// and all of those are one cell. The default is false: a type this function
// does not recognise is dropped and named, rather than flattened into
// something a reader would take for the value.
func flat(s *schema.Schema, typ string) bool {
	if _, isList := schema.ElementType(typ); isList {
		return false
	}
	base := typ
	if alt, ok := schema.LiteralAlternative(typ); ok {
		base = alt
	}
	if s.CompositeType(base) != nil {
		return false
	}
	if sc := s.ScalarType(base); sc != nil {
		return len(sc.Members) == 0
	}
	switch base {
	case "string", "float", "integer", "boolean":
		return true
	}
	return false
}

// WriteCSV writes the convenience view.
//
// The stamp comes first as `#` lines, one field per line, before the header
// row, and the last of them names what this format dropped. The lossiness is
// stated in the file itself rather than only in a document, because the file
// is what gets copied.
func WriteCSV(w io.Writer, a *Artifact, s *schema.Schema) error {
	columns := Columns(s)
	dropped := Dropped(s)

	var head strings.Builder
	head.WriteString("# This file is a convenience view and is not the authority.\n")
	fmt.Fprintf(&head, "# The lossless artifact is %s, and %s carries these lines as JSON.\n", JSONName, CSVStampName)
	for _, line := range stampLines(a.Stamp) {
		fmt.Fprintf(&head, "# %s\n", line)
	}
	fmt.Fprintf(&head, "# dropped: %s\n", strings.Join(dropped, ", "))
	if _, err := io.WriteString(w, head.String()); err != nil {
		return fmt.Errorf("cannot write the convenience view's stamp: %w", err)
	}

	out := csv.NewWriter(w)
	if err := out.Write(columns); err != nil {
		return fmt.Errorf("cannot write the convenience view's header: %w", err)
	}
	for _, record := range a.Records {
		row := make([]string, 0, len(columns))
		for _, column := range columns {
			cell, err := cell(lookup(record.Fields, column))
			if err != nil {
				return fmt.Errorf("%s: %s: %w", record.Path, column, err)
			}
			row = append(row, cell)
		}
		if err := out.Write(row); err != nil {
			return fmt.Errorf("cannot write %s: %w", record.Path, err)
		}
	}
	out.Flush()
	return out.Error()
}

// A CSVStamp is the sidecar beside the convenience view.
type CSVStamp struct {
	Stamp   Stamp    `json:"stamp"`
	Lossy   bool     `json:"lossy"`
	Of      string   `json:"of"`
	Dropped []string `json:"dropped"`
}

// WriteCSVStamp writes the sidecar.
//
// It exists because the `#` lines are a convention rather than part of any CSV
// specification, so a reader whose parser refuses them would otherwise have to
// strip the provenance to reach the data, which is the opposite of what a
// stamp is for.
func WriteCSVStamp(w io.Writer, a *Artifact, s *schema.Schema) error {
	sidecar := CSVStamp{Stamp: a.Stamp, Lossy: true, Of: CSVName, Dropped: Dropped(s)}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(sidecar); err != nil {
		return fmt.Errorf("cannot write the convenience view's sidecar: %w", err)
	}
	return nil
}

// stampLines renders the stamp one field per line, in the order the stamp
// declares them.
func stampLines(s Stamp) []string {
	lines := []string{
		fmt.Sprintf("stamp_version: %d", s.StampVersion),
		"corpus_version: " + s.CorpusVersion,
		"corpus_revision: " + s.CorpusRevision,
		"corpus_state: " + s.CorpusState,
	}
	if s.CorpusDirtyDigest != "" {
		lines = append(lines, "corpus_dirty_digest: "+s.CorpusDirtyDigest)
	}
	return append(lines,
		"tool_version: "+s.ToolVersion,
		"tool_revision: "+s.ToolRevision,
		"tool_state: "+s.ToolState,
		"command: "+s.Command,
		"options: "+options(s.Options),
		"selection: "+s.Selection,
		fmt.Sprintf("selected_count: %d", s.SelectedCount),
		"excluded: "+counts(s.Excluded),
	)
}

// options renders the options map sorted by name, which is what the versioning
// record asks for, and renders an empty one as a word rather than as nothing.
func options(m map[string]string) string {
	if len(m) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(m))
	for name, value := range m {
		parts = append(parts, name+"="+value)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// counts renders the excluded counts the same way.
func counts(m map[string]int) string {
	if len(m) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(m))
	for reason, n := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", reason, n))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// lookup reads a value out of a decoded record by the dotted path the schema
// addresses it with, and returns nil where the record does not carry it.
func lookup(fields map[string]any, path string) any {
	current := any(fields)
	for _, part := range strings.Split(path, ".") {
		block, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current, ok = block[part]
		if !ok {
			return nil
		}
	}
	return current
}

// cell renders one value.
//
// An absent value is an empty cell, which is the only thing a CSV can say. A
// value of a kind this function does not know is an error rather than a cell
// produced by %v: it would mean flat above called something one cell that is
// not, and a wrong number that looks like a number is the defect this whole
// repository is built against.
func cell(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	default:
		return "", fmt.Errorf("a %T is not a value one cell can hold, so this column is not the flat field the schema was read as declaring", value)
	}
}
