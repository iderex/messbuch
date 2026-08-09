package schema

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// rendered is a schema carrying one of every shape the printer has a sentence
// for: a closed set, a list with a minimum, a conditional field whose
// condition no record decides, a composite with an at-least-one-of rule, and
// each of the four kinds of scalar type.
//
// INVENTED. None of these is a field of a record and none of the patterns is
// this project's pattern. A fixture that judged against the real schema would
// prove the state of the tree on the day it ran rather than the printer.
const rendered = `
schema_version = 1
fixed_on = "1900-01-01"
record_path_pattern = '^record/x/y\.toml$'
excluded_directory_prefix = "_"
identifier_pattern = '^[a-z]+$'
identifier_max_length = 4
coverage_pattern = '^k=1$'

[[field]]
path = "state"
type = "string"
presence = "required"
values = ["open", "shut"]
means = "Invented for a fixture."
does_not_mean = "Anything about a record."

[[field]]
path = "note"
type = "string"
presence = "optional"
optional_because = "Invented for a fixture."

[[field]]
path = "amount"
type = "float"
presence = "conditional"
required_when = { field = "state", equals = "open" }
refused_when = { field = "state", in = ["shut"] }
minimum = 0.0

[[field]]
path = "signature"
type = "string"
presence = "conditional"
required_when = { source_stated = true }
machine_decidable = false
condition_note = "Invented for a fixture."

[[field]]
path = "readings"
type = "list<reading>"
presence = "required"
min_length = 2

[[field]]
path = "label"
type = "identifier-or-literal"
presence = "optional"
literals = ["none", "unknown"]

[[field]]
path = "seen"
type = "date-range"
presence = "optional"

[[field]]
path = "quantity"
type = "identifier"
presence = "required"
resolves_to = "vocabulary/<value>.toml"

[[field]]
path = "spread"
type = "any"
presence = "optional"

[type.reading]
note = "Invented for a fixture."
at_least_one_of = ["low", "high"]
fixed_by = "docs/decisions/0004-record-schema.md"

[[type.reading.member]]
name = "low"
type = "float"
presence = "optional"

[[type.reading.member]]
name = "high"
type = "float"
presence = "conditional"
refused_when = { field = "low", absent = true }

[scalar.identifier]
pattern_key = "identifier_pattern"
max_length_key = "identifier_max_length"
note = "Invented for a fixture."

[scalar.date-range]
members = ["start", "end"]
member_type = "partial-date"

[scalar.partial-date]
formats = ["YYYY", "YYYY-MM"]

[scalar.float]

[scalar.any]
note = "Invented for a fixture."
`

// render prints one fixture set and returns what a reader would see.
func render(t *testing.T, files map[string]string) string {
	t.Helper()
	set, err := Load(at(t, files))
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := Print(&out, set); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// TestPrintSaysEverythingTheFixtureDeclares is the printer's whole obligation:
// a reader who never opens the file learns what the file says. A rendering
// that silently dropped a field would leave a contributor writing a record
// against a rule they were never shown.
func TestPrintSaysEverythingTheFixtureDeclares(t *testing.T) {
	out := render(t, map[string]string{"record-1.toml": rendered})

	for _, want := range []string{
		"schema version 1, fixed on 1900-01-01, read from schema/record-1.toml",
		`^record/x/y\.toml$`,
		`begins with "_"`,
		"state",
		"a string, one of open or shut",
		"Required.",
		"Does not mean: Anything about a record.",
		"Optional.",
		"Optional because: Invented for a fixture.",
		"Required when state is open.",
		"Refused when state is one of shut.",
		"not below 0",
		"Required when the source stated one.",
		"Not decidable from the record alone: Invented for a fixture.",
		"a list of reading, with at least 2 entries",
		"an identifier, or one of none or unknown",
		"Resolves to vocabulary/<value>.toml",
		"BLOCKS",
		"reading",
		"At least one of low or high has to be present.",
		"Fixed by docs/decisions/0004-record-schema.md",
		"Refused when low is absent.",
		"VALUE TYPES",
		"a string matching ^[a-z]+$, of at most 4 characters",
		"a block with start and end, each a partial-date",
		"a date written as YYYY or YYYY-MM",
		"no syntax beyond what the note below says",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the rendering does not carry %q\n\n%s", want, out)
		}
	}
}

// TestPrintCarriesEveryNameTheRealSchemaDeclares reads the authority this
// repository actually ships rather than a fixture, because the failure worth
// refusing is a field that lands in schema/record-<n>.toml and never reaches
// the reader. It asserts presence and nothing about wording, so it does not
// become a second place the schema is written down.
func TestPrintCarriesEveryNameTheRealSchemaDeclares(t *testing.T) {
	set, err := Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := Print(&out, set); err != nil {
		t.Fatal(err)
	}
	text := out.String()

	names := 0
	for _, version := range set.SortedVersions() {
		s := set.Versions[version]
		for _, f := range s.Fields {
			names++
			if !strings.Contains(text, f.Path) {
				t.Errorf("version %d declares %s and the rendering does not name it", version, f.Path)
			}
			for _, value := range f.Values {
				if !strings.Contains(text, value) {
					t.Errorf("version %d allows %s at %s and the rendering does not name it", version, value, f.Path)
				}
			}
		}
		for name, c := range s.Composites {
			names++
			if !strings.Contains(text, name) {
				t.Errorf("version %d declares the block %s and the rendering does not name it", version, name)
			}
			for _, m := range c.Members {
				if !strings.Contains(text, m.Name) {
					t.Errorf("version %d declares %s.%s and the rendering does not name it", version, name, m.Name)
				}
			}
		}
		for name := range s.Scalars {
			names++
			if !strings.Contains(text, name) {
				t.Errorf("version %d declares the type %s and the rendering does not name it", version, name)
			}
		}
	}
	if names == 0 {
		t.Fatal("no name was checked, so this test asserted nothing about the schema this repository ships")
	}
	t.Logf("%d declared name(s) checked against the rendering", names)
}

// TestPrintIsTheSameBytesTwice holds the property a reader diffs two runs on.
// Fields come out of a slice and blocks and value types out of maps, and a map
// walked in its own order would print two orders on two runs.
func TestPrintIsTheSameBytesTwice(t *testing.T) {
	files := map[string]string{"record-1.toml": rendered}
	first := render(t, files)
	for range 8 {
		if second := render(t, files); second != first {
			t.Fatalf("two runs printed different bytes\n\nfirst:\n%s\n\nsecond:\n%s", first, second)
		}
	}
}

// TestPrintWritesTheVersionsInOrder covers the set rather than the schema. A
// corpus keeps every version's file, so the rendering has to be readable when
// there is more than one, and in an order that does not move.
func TestPrintWritesTheVersionsInOrder(t *testing.T) {
	second := strings.Replace(rendered, "schema_version = 1", "schema_version = 2", 1)
	out := render(t, map[string]string{"record-1.toml": rendered, "record-2.toml": second})

	one := strings.Index(out, "schema version 1,")
	two := strings.Index(out, "schema version 2,")
	if one < 0 || two < 0 {
		t.Fatalf("both versions should be printed\n\n%s", out)
	}
	if one > two {
		t.Errorf("version 2 was printed before version 1")
	}
}

// TestPrintReportsAWriteFailure covers the one thing that can go wrong on the
// way out. A rendering half written to a closed pipe and reported as complete
// is a reader who thinks they have seen the whole schema.
func TestPrintReportsAWriteFailure(t *testing.T) {
	set, err := Load(at(t, map[string]string{"record-1.toml": rendered}))
	if err != nil {
		t.Fatal(err)
	}
	broken := errors.New("the pipe is shut")
	for _, after := range []int{0, 1, 5} {
		w := &failing{after: after, err: broken}
		if err := Print(w, set); !errors.Is(err, broken) {
			t.Errorf("a writer failing after %d line(s) gave %v, wanted the write error", after, err)
		}
	}
}

// failing accepts a number of writes and then refuses.
type failing struct {
	after int
	n     int
	err   error
}

func (f *failing) Write(p []byte) (int, error) {
	if f.n >= f.after {
		return 0, f.err
	}
	f.n++
	return len(p), nil
}

func TestDescribeSaysWhichFieldDecided(t *testing.T) {
	yes, no := true, false
	for _, c := range []struct {
		name string
		cond *Condition
		want string
	}{
		{"nothing at all", nil, "the schema says so"},
		{"an empty condition", &Condition{}, "the schema says so"},
		{"never", &Condition{Never: &yes}, "never"},
		{"a vocabulary field", &Condition{VocabularyField: "epochs"}, "the quantity's vocabulary entry says so, in epochs"},
		{"the source stated one", &Condition{SourceStated: &yes}, "the source stated one"},
		{"the source stated none", &Condition{SourceStated: &no}, "the source stated none"},
		{"present", &Condition{Field: "a", Present: &yes}, "a is present"},
		{"not present", &Condition{Field: "a", Present: &no}, "a is absent"},
		{"absent", &Condition{Field: "a", Absent: &yes}, "a is absent"},
		{"not absent", &Condition{Field: "a", Absent: &no}, "a is present"},
		{"equals", &Condition{Field: "a", Equals: "x"}, "a is x"},
		{"not equals", &Condition{Field: "a", NotEquals: "x"}, "a is not x"},
		{"one of", &Condition{Field: "a", In: []any{"x", "y"}}, "a is one of x, y"},
	} {
		if got := c.cond.Describe(); got != c.want {
			t.Errorf("%s: got %q, wanted %q", c.name, got, c.want)
		}
	}
}

func TestTypeNameReadsTheSchemaSpellings(t *testing.T) {
	for _, c := range []struct{ typ, want string }{
		{"string", "a string"},
		{"integer", "an integer"},
		{"identifier", "an identifier"},
		{"list<reading>", "a list of reading"},
		{"identifier-or-literal", "an identifier"},
		{"list<identifier-or-literal>", "a list of identifier-or-literal"},
		{"", "a "},
	} {
		if got := typeName(c.typ); got != c.want {
			t.Errorf("%q: got %q, wanted %q", c.typ, got, c.want)
		}
	}
}

func TestPresenceOfAnEntryTheSchemaLeftEmpty(t *testing.T) {
	if got := presence(&Field{}); got != nil {
		t.Errorf("an entry with no presence rule gave %q, and there is no sentence to write about it", got)
	}
	if got := presence(&Field{Presence: "conditional"}); got != nil {
		t.Errorf("a conditional entry with no condition gave %q; the loader refuses one, so nothing is invented here", got)
	}
}

func TestJoinersReadAsSentences(t *testing.T) {
	for _, c := range []struct {
		items          []string
		choice, joined string
	}{
		{nil, "", ""},
		{[]string{"a"}, "a", "a"},
		{[]string{"a", "b"}, "a or b", "a and b"},
		{[]string{"a", "b", "c"}, "a, b or c", "a, b and c"},
	} {
		if got := choices(c.items); got != c.choice {
			t.Errorf("choices(%v) gave %q, wanted %q", c.items, got, c.choice)
		}
		if got := list(c.items); got != c.joined {
			t.Errorf("list(%v) gave %q, wanted %q", c.items, got, c.joined)
		}
	}
	if got := plural(1); got != "y" {
		t.Errorf("one entry is written %q", got)
	}
	if got := upperFirst(""); got != "" {
		t.Errorf("an empty sentence gave %q", got)
	}
}

// TestWrappedKeepsWordsWhole is the one property a reader notices immediately.
// A fold inside a field name or a pattern would make the rendering unreadable
// exactly where it is being read most carefully.
func TestWrappedKeepsWordsWhole(t *testing.T) {
	const width = 76
	out := render(t, map[string]string{"record-1.toml": rendered})
	long := strings.Repeat("word ", 60) + "supercalifragilisticexpialidociousandthensome"

	sb := &strings.Builder{}
	p := &printer{w: sb}
	p.wrapped("    ", long)
	for _, line := range strings.Split(strings.TrimRight(sb.String(), "\n"), "\n") {
		if !strings.HasPrefix(line, "    ") {
			t.Errorf("a folded line lost its indent: %q", line)
		}
		if len(line) > width && !strings.Contains(line, "supercali") {
			t.Errorf("a line is %d characters wide: %q", len(line), line)
		}
	}
	if strings.Contains(sb.String(), "supercali\n") {
		t.Error("a word longer than the width was split rather than left whole")
	}
	if out == "" {
		t.Error("the fixture rendered nothing")
	}
}
