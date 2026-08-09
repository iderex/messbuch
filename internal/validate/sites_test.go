package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iderex/messbuch/internal/schema"
)

// The fixtures in this file exist because the gate's refusal-site accounting
// asked for them by name. Each reaches one place in the validator that no
// other fixture executes, and the accounting is what found them: fourteen
// branches were refusing nothing that anybody had watched, and a branch nobody
// has watched fire is a branch nobody knows is wired up.
//
// They are grouped here rather than spread through the behaviour tests so that
// what they are for stays legible. Every one of them still asserts what came
// out, because a fixture that only executes a line proves the line exists.

// earns runs a fixture and reports whether it produced a refusal at that site.
func earns(t *testing.T, body, site string) []Refusal {
	t.Helper()
	rs := refuse(t, body)
	for _, r := range rs {
		if r.Site == site {
			return rs
		}
	}
	t.Fatalf("expected %s and got %v", site, rs)
	return nil
}

// A path the schema declares fields underneath, given something that is not a
// table. Nothing below it can be looked for, so the refusal is about the table
// rather than about the fields.
func TestStructureRefusesABlockPathThatIsNotABlock(t *testing.T) {
	body := replace(t, "[method]\ntechnique = \"example-technique\"\n\n", "")
	body = strings.Replace(body, "blinding = \"not-stated\"", "blinding = \"not-stated\"\nmethod = \"beam\"", 1)
	rs := earns(t, body, "wrong-type")
	for _, r := range rs {
		if r.Site == "wrong-type" && r.Field != "method" {
			t.Fatalf("the refusal has to name the path that is not a table: %s", r)
		}
	}
}

// A top-level field another top-level field's value requires. The member-level
// arm of the same rule is reached by the conversion block, and the two are
// different places in the source.
func TestStructureRefusesATopLevelFieldTheRecordOwes(t *testing.T) {
	body := replace(t, "[measurement.normalized]\nvalue = 1.5\nunit = \"s\"\n", "")
	body = strings.Replace(body, "[[measurement.normalized.uncertainty]]\ncomponent = \"statistical\"\nplus = 0.1\nminus = 0.1\ncoverage = \"k=1\"\ninterval_kind = \"frequentist\"\n", "", 1)
	rs := earns(t, body, "conditional-field-missing")
	found := false
	for _, r := range rs {
		if r.Site == "conditional-field-missing" && r.Field == "measurement.normalized" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the normalized block is owed when the status says normalized: %v", rs)
	}
}

func TestStructureRefusesAListFieldThatIsNotAList(t *testing.T) {
	body := replace(t, "blinding = \"not-stated\"", "blinding = \"not-stated\"\ncorrection = \"a typo was fixed\"")
	earns(t, body, "wrong-type")
}

func TestStructureRefusesABlockFieldThatIsNotABlock(t *testing.T) {
	body := replace(t, "[source.locator]\npage = \"1\"\n", "")
	body = strings.Replace(body, "[source]\n", "[source]\nlocator = \"page 1\"\n", 1)
	earns(t, body, "wrong-type")
}

// An identifier of legal shape and illegal length. The pattern alone accepts
// it, which is why the length is checked beside the pattern rather than
// written into it.
func TestStructureRefusesAnIdentifierLongerThanTheSchemaAllows(t *testing.T) {
	long := "a" + strings.Repeat("bc", 30)
	if len(long) != 61 {
		t.Fatalf("this fixture is meant to be one character over the limit and is %d", len(long))
	}
	rs := earns(t, replace(t, `quantity = "example-quantity"`, `quantity = "`+long+`"`), "malformed-value")
	if !strings.Contains(rs[0].Found, "61 characters") {
		t.Fatalf("the refusal has to say how long it was: %s", rs[0])
	}
}

func TestStructureRefusesADateThatIsNotAString(t *testing.T) {
	earns(t, replace(t, `date = "1900"`, "date = 1900"), "wrong-type")
}

func TestStructureRefusesADateRangeThatIsNotABlock(t *testing.T) {
	body := replace(t, `data_taken_status = "none-in-source"`, "data_taken_status = \"reported\"\ndata_taken = \"1900\"")
	earns(t, body, "wrong-type")
}

// One end of a range alone is a different fact from a range, so a date range
// carries both of its members.
func TestStructureRefusesADateRangeMissingAnEnd(t *testing.T) {
	body := replace(t, `data_taken_status = "none-in-source"`, `data_taken_status = "reported"`)
	body = strings.Replace(body, "[method]\n", "[publication.data_taken]\nstart = \"1900\"\n\n[method]\n", 1)
	rs := earns(t, body, "missing-required-field")
	found := false
	for _, r := range rs {
		if r.Field == "publication.data_taken.end" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the missing end is not named: %v", rs)
	}
}

func TestStructureRefusesAMemberADateRangeDoesNotHave(t *testing.T) {
	body := replace(t, `data_taken_status = "none-in-source"`, `data_taken_status = "reported"`)
	body = strings.Replace(body, "[method]\n", "[publication.data_taken]\nstart = \"1900\"\nend = \"1901\"\nmiddle = \"1900\"\n\n[method]\n", 1)
	earns(t, body, "unknown-field")
}

func TestStructureRefusesAPlainStringFieldThatIsNotAString(t *testing.T) {
	earns(t, replace(t, `unit = "example unit"`, "unit = 4"), "wrong-type")
}

func TestStructureRefusesAnIntegerFieldThatIsNotAnInteger(t *testing.T) {
	body := replace(t, "[source.locator]\n", "[source.print]\npublisher = \"Invented Journal\"\npages = \"1-2\"\nyear = \"1900\"\nauthors = [\"Example\"]\n\n[source.locator]\n")
	earns(t, body, "wrong-type")
}

func TestStructureRefusesABooleanFieldThatIsNotABoolean(t *testing.T) {
	earns(t, replace(t, "exact = true", `exact = "true"`), "wrong-type")
}

func TestStructureRefusesARequiredMemberOfABlock(t *testing.T) {
	body := replace(t, "coverage = \"k=1\"\ninterval_kind = \"frequentist\"\n", "coverage = \"k=1\"\n")
	rs := earns(t, body, "missing-required-field")
	if rs[0].Field != "measurement.published.uncertainty[0].interval_kind" {
		t.Fatalf("the missing member is not named where it sits: %v", rs)
	}
}

// The arm that catches a type the schema declares and this package would not
// have applied. It needs a schema of its own, because this repository's schema
// declares no such type and the loader is what keeps it that way.
func TestStructureRefusesAValueOfATypeNothingKnowsHowToCheck(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `
schema_version = 1
fixed_on = "1900-01-01"
record_path_pattern = '^record/x/y\.toml$'
excluded_directory_prefix = "_"
identifier_pattern = '^[a-z]+$'
identifier_max_length = 4
coverage_pattern = '^k=1$'

[[field]]
path = "schema_version"
type = "integer"
presence = "required"
means = "Invented for a fixture."

[[field]]
path = "oddity"
type = "oddity"
presence = "required"
means = "Invented for a fixture."

[scalar.oddity]
note = "A named scalar carrying no pattern, no formats and no members, so nothing decides a value written under it."
`
	if err := os.WriteFile(filepath.Join(root, "schema", "record-1.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := schema.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	rs := Record("record/x/y.toml", []byte("schema_version = 1\noddity = \"anything\"\n"), set)
	if len(rs) != 1 || rs[0].Site != "wrong-type" {
		t.Fatalf("a type nothing knows how to check has to be refused rather than passed: %v", rs)
	}
	if !strings.Contains(rs[0].Expected, "oddity") {
		t.Fatalf("the refusal has to name the type: %s", rs[0])
	}
}
