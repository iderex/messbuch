package validate

import (
	"strings"
	"testing"

	"github.com/iderex/messbuch/internal/schema"
)

// wellFormed is the fixture every refusal below is one edit away from.
//
// INVENTED. Nothing in it is transcribed from any publication, no quantity,
// group, technique or identifier named here exists, and none of it may be
// cited or copied into a record. It is a shape and nothing else. Its job is to
// be legal, so that a fixture derived from it differs from a legal record in
// exactly one place and the refusal it earns can be attributed to that place.
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

// fixturePath is where the fixture pretends to live. The path is legal under
// the schema's own pattern, so a refusal about a path is a fixture of its own
// rather than something every other fixture also earns.
const fixturePath = "record/example-quantity/1900-example-01.toml"

func load(t *testing.T) *schema.Set {
	t.Helper()
	set, err := schema.Load("../..")
	if err != nil {
		t.Fatalf("this repository's own schema does not load: %v", err)
	}
	return set
}

// refuse runs one fixture and returns what it earned.
func refuse(t *testing.T, body string) []Refusal {
	t.Helper()
	return Record(fixturePath, []byte(body), load(t))
}

// exactly asserts that a fixture is refused, that every refusal it earned
// names the same site, and that the site is the one named.
//
// The set comparison is the leg that matters. A fixture tripping two checks
// proves neither, because the refusal under test could have been produced by
// the other defect and nobody would see it.
func exactly(t *testing.T, body, site string) []Refusal {
	t.Helper()
	rs := refuse(t, body)
	if len(rs) == 0 {
		t.Fatalf("expected %s and the fixture was accepted", site)
	}
	for _, r := range rs {
		if r.Site != site {
			t.Fatalf("expected %s alone and this fixture also earned %s: %s", site, r.Site, r)
		}
	}
	return rs
}

// replace edits the fixture in exactly one place and fails if the edit did not
// apply, so a renamed field in the schema turns into a failing test rather
// than a fixture that quietly stopped testing anything.
func replace(t *testing.T, old, new string) string {
	t.Helper()
	if !strings.Contains(wellFormed, old) {
		t.Fatalf("the fixture no longer contains %q, so this edit tests nothing", old)
	}
	return strings.Replace(wellFormed, old, new, 1)
}

func TestStructureAcceptsAWellFormedRecord(t *testing.T) {
	if rs := refuse(t, wellFormed); len(rs) != 0 {
		for _, r := range rs {
			t.Errorf("%s", r)
		}
		t.Fatalf("the fixture every other test edits is itself refused, so none of them proves anything")
	}
}

func TestStructureRefusesAFileThatDoesNotParse(t *testing.T) {
	rs := exactly(t, "schema_version = 1\nstatus = \"active\n", "record-does-not-parse")
	if !strings.Contains(rs[0].Found, "line") {
		t.Fatalf("a parse failure has to say where it is: %s", rs[0])
	}
}

func TestStructureRefusesARecordNamingNoSchemaVersion(t *testing.T) {
	exactly(t, replace(t, "schema_version = 1\n", ""), "schema-version")
}

func TestStructureRefusesASchemaVersionThatIsNotANumber(t *testing.T) {
	exactly(t, replace(t, "schema_version = 1", `schema_version = "1"`), "schema-version")
}

func TestStructureRefusesAVersionThisTreeHasNoFileFor(t *testing.T) {
	rs := exactly(t, replace(t, "schema_version = 1", "schema_version = 9"), "schema-version")
	if !strings.Contains(rs[0].Expected, "1") {
		t.Fatalf("the refusal has to say which versions exist: %s", rs[0])
	}
}

func TestStructureRefusesAnUnknownField(t *testing.T) {
	exactly(t, replace(t, "status = \"active\"", "status = \"active\"\nconfidence = \"high\""), "unknown-field")
}

// The near miss for the unknown-field refusal. A misspelling of a real field
// is the mistake somebody actually makes, and silently dropping it produces a
// record that validates and is missing data, which nothing downstream can see.
func TestStructureRefusesAMisspeltFieldRatherThanDroppingIt(t *testing.T) {
	// This fixture earns two refusals on purpose and it is the pair that
	// carries the proof. The misspelling is refused as a field the schema does
	// not declare, and the field it was meant to be is reported missing. A
	// validator that dropped the unknown key would produce only the second,
	// and the contributor would be told a field is missing while looking at a
	// line that appears to set it.
	rs := refuse(t, replace(t, "blinding = ", "blindng = "))
	sites := map[string]string{}
	for _, r := range rs {
		sites[r.Site] = r.Field
	}
	if sites["unknown-field"] != "blindng" {
		t.Fatalf("the misspelling is not refused by name: %v", rs)
	}
	if sites["missing-required-field"] != "blinding" {
		t.Fatalf("the field the misspelling was meant to be is not reported missing: %v", rs)
	}
	if len(rs) != 2 {
		t.Fatalf("expected exactly those two, got %d: %v", len(rs), rs)
	}
}

func TestStructureRefusesAMissingRequiredField(t *testing.T) {
	exactly(t, replace(t, "technique = \"example-technique\"\n", ""), "missing-required-field")
}

func TestStructureRefusesAFieldAnotherFieldRequires(t *testing.T) {
	body := replace(t, "uncertainty_status = \"reported\"", "uncertainty_status = \"none-in-source\"")
	rs := refuse(t, body)
	// Changing the status refuses the uncertainty blocks that the old status
	// required, so this fixture earns the refused arm as well; what it must
	// carry is the missing arm for the block that is now required to be
	// absent nowhere else.
	if len(rs) == 0 {
		t.Fatalf("expected the uncertainty blocks to be refused under none-in-source")
	}
	for _, r := range rs {
		if r.Site != "conditional-field-refused" {
			t.Fatalf("expected the refused arm alone, got %s", r)
		}
	}
}

func TestStructureRefusesAFieldTheRecordOwesUnderItsOwnValues(t *testing.T) {
	body := replace(t, "exact = true", "exact = false")
	rs := refuse(t, body)
	if len(rs) != 2 {
		t.Fatalf("expected the two members an inexact factor owes, got %d: %v", len(rs), rs)
	}
	for _, r := range rs {
		if r.Site != "conditional-field-missing" {
			t.Fatalf("expected conditional-field-missing alone, got %s", r)
		}
	}
}

func TestStructureRefusesAFieldAnotherFieldForbids(t *testing.T) {
	exactly(t, replace(t, "status = \"active\"", "status = \"active\"\nsuperseded_by = \"record/example-quantity/1901-example-01.toml\""), "conditional-field-refused")
}

func TestStructureRefusesAWrongType(t *testing.T) {
	exactly(t, replace(t, `quantity = "example-quantity"`, "quantity = 4"), "wrong-type")
}

// The near miss for the float type. A half-width written without a decimal
// point is a TOML integer, it looks exactly like a number, and everything
// downstream reading it as a float gets a type it did not ask for.
func TestStructureRefusesAWholeNumberWhereAFloatIsDeclared(t *testing.T) {
	rs := exactly(t, replace(t, "plus = 0.1", "plus = 0"), "wrong-type")
	if !strings.Contains(rs[0].Expected, "decimal point") {
		t.Fatalf("the refusal has to say what the difference is: %s", rs[0])
	}
}

func TestStructureRefusesAMalformedIdentifier(t *testing.T) {
	exactly(t, replace(t, `quantity = "example-quantity"`, `quantity = "Example Quantity"`), "malformed-value")
}

// The near miss for an identifier: a hyphen that is not the hyphen it looks
// like. U+2011 renders as a hyphen, the repository's Unicode guard permits it
// because it is neither bidirectional nor invisible, and nothing but this
// pattern refuses it.
func TestStructureRefusesAHyphenThatIsNotAHyphen(t *testing.T) {
	exactly(t, replace(t, `quantity = "example-quantity"`, "quantity = \"example‑quantity\""), "malformed-value")
}

// The near miss for a date: a year written with two digits. It is the mistake
// a transcriber working through a series of papers from one decade makes, and
// a corpus sorted by year puts it in the first century.
func TestStructureRefusesATwoDigitYear(t *testing.T) {
	exactly(t, replace(t, `date = "1900"`, `date = "00"`), "malformed-value")
}

// A date of the right shape naming a day that does not exist. The pattern
// alone would accept it, which is why the formats are parsed rather than
// matched.
func TestStructureRefusesADayThatDoesNotExist(t *testing.T) {
	exactly(t, replace(t, `date = "1900"`, `date = "1900-02-30"`), "malformed-value")
}

func TestStructureAcceptsEveryPrecisionADateMayBeWrittenAt(t *testing.T) {
	for _, written := range []string{"1900", "1900-02", "1900-02-28"} {
		if rs := refuse(t, replace(t, `date = "1900"`, `date = "`+written+`"`)); len(rs) != 0 {
			t.Fatalf("%s is refused and the schema names it a format: %v", written, rs)
		}
	}
}

func TestStructureRefusesAValueOutsideAClosedSet(t *testing.T) {
	exactly(t, replace(t, `statement_kind = "primary-result"`, `statement_kind = "primary result"`), "value-outside-set")
}

// The near miss for a closed set: every value the schema does declare is
// accepted, so the refusal above is about membership rather than about the
// field being present at all.
func TestStructureAcceptsEveryValueTheClosedSetDeclares(t *testing.T) {
	set := load(t)
	var values []string
	for _, f := range set.Versions[1].Fields {
		if f.Path == "source.statement_kind" {
			values = f.Values
		}
	}
	if len(values) < 2 {
		t.Fatalf("the schema no longer fixes a closed set for source.statement_kind, so this test proves nothing")
	}
	for _, value := range values {
		body := replace(t, `statement_kind = "primary-result"`, `statement_kind = "`+value+`"`)
		if rs := refuse(t, body); len(rs) != 0 {
			t.Fatalf("%s is in the schema's own set and was refused: %v", value, rs)
		}
	}
}

func TestStructureRefusesAListShorterThanItsMinimum(t *testing.T) {
	body := replace(t, "[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n", "identifier = []\n")
	exactly(t, body, "list-too-short")
}

func TestStructureRefusesANumberBelowItsMinimum(t *testing.T) {
	exactly(t, replace(t, "plus = 0.1", "plus = -0.1"), "below-minimum")
}

func TestStructureRefusesTheSameIdentifierWrittenTwice(t *testing.T) {
	body := replace(t,
		"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n",
		"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n\n[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n")
	exactly(t, body, "duplicate-in-list")
}

// The near miss for the duplicate rule: two identifiers that differ are two
// identifiers, and the rule must not refuse a source named in two schemes.
func TestStructureAcceptsTwoIdentifiersThatDiffer(t *testing.T) {
	body := replace(t,
		"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n",
		"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n\n[[source.identifier]]\nscheme = \"arxiv\"\nvalue = \"0000.00000\"\n")
	if rs := refuse(t, body); len(rs) != 0 {
		t.Fatalf("two identifiers in two schemes are two identifiers: %v", rs)
	}
}

// The exemption the duplicate rule needs. Two authors of one paper can carry
// the same name, and that is a fact about the literature rather than a
// transcription written twice.
func TestStructureAcceptsARepeatedNameInAListOfStrings(t *testing.T) {
	body := replace(t,
		"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n",
		"[source.print]\npublisher = \"Invented Journal\"\npages = \"1-2\"\nyear = 1900\nauthors = [\"Example\", \"Example\"]\n")
	// Removing the identifier makes source.resolvable owed, which is a
	// different refusal and would mask this one, so it is written in too.
	body = strings.Replace(body, "[source]\n", "[source]\nresolvable = false\n", 1)
	if rs := refuse(t, body); len(rs) != 0 {
		t.Fatalf("a repeated author name is a fact about a paper: %v", rs)
	}
}

func TestStructureRefusesABlockCarryingNoneOfItsMembers(t *testing.T) {
	exactly(t, replace(t, "[source.locator]\npage = \"1\"", "[source.locator]"), "no-member-present")
}

func TestStructureRefusesAMemberTheCompositeDoesNotDeclare(t *testing.T) {
	exactly(t, replace(t, "[source.locator]\npage = \"1\"", "[source.locator]\npage = \"1\"\npdf_page = \"7\""), "unknown-field")
}

func TestStructureReportsEveryRefusalRatherThanTheFirst(t *testing.T) {
	body := replace(t, `quantity = "example-quantity"`, `quantity = "Example Quantity"`)
	body = strings.Replace(body, `statement_kind = "primary-result"`, `statement_kind = "primary result"`, 1)
	rs := Record(fixturePath, []byte(body), load(t))
	if len(rs) != 2 {
		t.Fatalf("expected both defects in one round, got %d: %v", len(rs), rs)
	}
}

func TestStructureNamesTheFileTheFieldWhatWasFoundAndWhatWasExpected(t *testing.T) {
	rs := exactly(t, replace(t, `quantity = "example-quantity"`, "quantity = 4"), "wrong-type")
	r := rs[0]
	if r.File != fixturePath || r.Field != "measurement.quantity" || r.Found == "" || r.Expected == "" {
		t.Fatalf("a refusal that does not say where and what is a punishment rather than a message: %#v", r)
	}
	if !strings.Contains(r.String(), fixturePath) || !strings.Contains(r.String(), "measurement.quantity") {
		t.Fatalf("the printed line drops what the fields carry: %s", r)
	}
}

// The schema marks two conditions as ones no reading of a record can decide.
// A validator treating them as ordinary conditions would refuse correct
// records, which is the failure this test holds shut.
func TestStructureLeavesAnUndecidableRequirementAlone(t *testing.T) {
	body := replace(t, `status = "active"`, `status = "withdrawn"`)
	for _, r := range refuse(t, body) {
		if r.Field == "superseded_by" {
			t.Fatalf("the schema says a validator cannot require this one: %s", r)
		}
	}
}

// Every name in the catalogue is reached by a fixture in this file, and every
// name a fixture reaches is in the catalogue. This is the accounting at the id
// level, and it is the cheaper half: the gate's refusal-sites leg does the
// same job per place in the source, which is what catches a second branch
// added under a name that already has a fixture.
func TestStructureEveryCatalogueSiteIsReachedByAFixture(t *testing.T) {
	set := load(t)
	reached := map[string]bool{}
	for _, body := range everyFixture(t) {
		for _, r := range Record(fixturePath, []byte(body), set) {
			reached[r.Site] = true
		}
	}
	// The path refusal is earned by a walk rather than by a file's contents,
	// and its fixture is in corpus_test.go.
	reached["record-in-the-wrong-place"] = corpusFixtureReachesThePathRefusal(t)

	var missing []string
	for _, site := range Catalogue {
		if !reached[site.ID] {
			missing = append(missing, site.ID)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%d catalogue site(s) no fixture reaches: %s", len(missing), strings.Join(missing, ", "))
	}
}

// everyFixture is the list the accounting runs over. It is here rather than
// derived from the test functions because a test that stopped running would
// otherwise take its site out of the accounting with it.
func everyFixture(t *testing.T) []string {
	t.Helper()
	return []string{
		"schema_version = 1\nstatus = \"active\n",
		replace(t, "schema_version = 1\n", ""),
		replace(t, "schema_version = 1", "schema_version = 9"),
		replace(t, "status = \"active\"", "status = \"active\"\nconfidence = \"high\""),
		replace(t, "technique = \"example-technique\"\n", ""),
		replace(t, "exact = true", "exact = false"),
		replace(t, "status = \"active\"", "status = \"active\"\nsuperseded_by = \"record/example-quantity/1901-example-01.toml\""),
		replace(t, `quantity = "example-quantity"`, "quantity = 4"),
		replace(t, `quantity = "example-quantity"`, `quantity = "Example Quantity"`),
		replace(t, `statement_kind = "primary-result"`, `statement_kind = "primary result"`),
		replace(t, "[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n", "identifier = []\n"),
		replace(t, "plus = 0.1", "plus = -0.1"),
		replace(t,
			"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n",
			"[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n\n[[source.identifier]]\nscheme = \"doi\"\nvalue = \"10.0000/invented-for-a-fixture\"\n"),
		replace(t, "[source.locator]\npage = \"1\"", "[source.locator]"),
	}
}
