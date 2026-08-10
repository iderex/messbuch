package validate

import (
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/iderex/messbuch/internal/schema"
)

// Coverage guided fuzzing of the record parser (#52).
//
// The surface is what somebody else writes. A record arrives as a file in a
// pull request from a person nobody has met, and the analysis will read a
// corpus a user assembled anywhere, so the bytes reaching Record are untrusted
// by any reasonable definition. What a crafted file could do here is exhaust
// memory, recurse without bound, take pathological time, or walk out of the
// tree through a path it names.
//
// The schema is loaded from this repository rather than from the fuzzed input.
// A target feeding arbitrary bytes to the schema loader would be fuzzing
// internal/schema, which reads a file this project controls and no contributor
// supplies. That is a different threat and it is not this one.
//
// # What the targets assert beyond not crashing
//
// A fuzz target that only waits for a panic finds the crash class and nothing
// else, and the defects this package can actually ship are quieter than that.
// Each target below checks properties instead, and each property is one a real
// change could break:
//
// Two runs over the same bytes produce the same refusals, in the same order.
// The built artifact is required to be reproducible byte for byte, which is
// #28, and a walk whose order comes from map iteration breaks that without
// failing anything.
//
// Every refusal names a site in Catalogue, names the file it was handed, and
// says both what was found and what was expected. A refusal that says a file
// is invalid without saying where is a punishment rather than a message, which
// is this package's own sentence about itself.
//
// A file that earns no refusal decodes as TOML and names a schema version this
// tree carries. Acceptance is the outcome with no second chance: a record that
// passes here is data everything downstream trusts.
//
// # From a crash to a fixture
//
// This is the path the issue asks to be written down, and it has been walked
// once against a defect introduced on purpose. The steps:
//
//  1. The failing run writes the input under
//     internal/validate/testdata/fuzz/<target>/<hash>. That file is the crash.
//  2. Minimise it. The fuzzer shrinks before it writes, so the file is usually
//     already small; cut it further by hand until one more cut makes the
//     failure go away.
//  3. Add it as a seed in this file, with a comment naming the defect it
//     exposed. A seed runs on every ordinary `go test`, so the input is
//     refused forever after without the fuzzer having to rediscover it.
//  4. Delete the file from testdata. It has become a fixture; leaving both
//     means the same input runs twice under two names.
//
// A crash that reaches a refusal site with no fixture of its own discharges
// that site's obligation under the gate's refusal-sites leg rather than
// creating a parallel set of proofs beside it.

// seeds are the inputs every ordinary run starts from, and the corpus the
// scheduled run mutates.
//
// The near-misses matter more than the well-formed record. A mutator handed
// only legal bytes spends its first minutes rediscovering that random bytes
// are not TOML, and the interesting region is one edit away from legal rather
// than far from it.
func seeds() []string {
	return []string{
		wellFormed,

		// Not TOML at all: the unterminated string is the shortest route to
		// the decoder's own error path.
		"schema_version = 1\nstatus = \"active\n",

		// No version, and a version that is not a number. Both stop the walk
		// before any field is read, so they are the two shapes that exercise
		// the early return rather than the checker.
		"status = \"active\"\n",
		"schema_version = \"1\"\n",

		// A version with no file in this tree.
		"schema_version = 99\n",

		// One unknown field, which is the misspelling case, and a second one
		// beside it, because two unknown siblings at one level is what makes
		// the order of the walk visible at all.
		strings.Replace(wellFormed, "status = \"active\"", "statuss = \"active\"", 1),
		strings.Replace(wellFormed, "status = \"active\"", "statuss = \"active\"\nblinded = false", 1),

		// A value of the right kind whose text does not satisfy its type: the
		// identifier pattern is the first parser underneath the decoder.
		strings.Replace(wellFormed, "quantity = \"example-quantity\"", "quantity = \"Example Quantity\"", 1),

		// A number where the schema wants a float. A number written without a
		// decimal point is a TOML integer, and the two are different types to
		// everything that reads the file.
		strings.Replace(wellFormed, "value = 1.5", "value = 1", 1),

		// The empty file, which is legal TOML and carries nothing.
		"",
	}
}

// check runs one input and asserts everything this package promises about
// whatever comes back.
func check(t *testing.T, rel string, content []byte, set *schema.Set) {
	t.Helper()

	first := Record(rel, content, set)
	second := Record(rel, content, set)

	if len(first) != len(second) {
		t.Fatalf("two runs over the same bytes produced %d and %d refusal(s)", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("two runs over the same bytes disagree at %d:\n\t%s\n\t%s", i, first[i], second[i])
		}
	}

	ordered := make([]Refusal, len(first))
	copy(ordered, first)
	Sort(ordered)
	for i := range first {
		if first[i] != ordered[i] {
			t.Fatalf("the refusals came back in an order Sort does not produce, first at %d:\n\t%s\n\t%s", i, first[i], ordered[i])
		}
	}

	for _, r := range first {
		if !siteExists[r.Site] {
			t.Fatalf("refusal names the site %q, which is not in Catalogue, so nothing would print it: %s", r.Site, r)
		}
		if r.File != rel {
			t.Fatalf("refusal names the file %q and this input was handed %q: %s", r.File, rel, r)
		}
		if r.Found == "" || r.Expected == "" {
			t.Fatalf("refusal says neither what was found nor what was expected, which tells a contributor nothing: %+v", r)
		}
	}

	if len(first) == 0 {
		var root map[string]any
		if _, err := toml.Decode(string(content), &root); err != nil {
			t.Fatalf("a file that does not decode as TOML earned no refusal: %v", err)
		}
		version, ok := root["schema_version"].(int64)
		if !ok {
			t.Fatalf("a file naming no integer schema version earned no refusal")
		}
		if _, ok := set.Versions[int(version)]; !ok {
			t.Fatalf("a file naming schema version %d, which this tree carries no file for, earned no refusal", version)
		}
	}
}

// FuzzRecord is the whole-file target: arbitrary bytes at the entry point a
// record arrives through.
func FuzzRecord(f *testing.F) {
	set := load(f)
	for _, seed := range seeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		check(t, fixturePath, []byte(body), set)
	})
}

// FuzzRecordValues is the structure aware target, and it is the one that
// reaches the number and unit parsing underneath the decoder.
//
// A whole-file mutator spends almost all of its budget on bytes that do not
// decode, so the pattern matcher, the identifier length bound, the date
// formats and the float reading are all behind a wall of parse failures it
// rarely gets through. This one splices fuzzed values into a legal record
// instead, so those parsers are reached on the first execution rather than by
// luck.
//
// The three text slots are quoted on the way in, so what is fuzzed is the
// value rather than the TOML around it. The number slot is spliced raw,
// because a number's text IS what the decoder parses and quoting it would test
// the string path instead.
//
// The half-width arrives as a float rather than as text, and the difference is
// what makes the schema's own bounds reachable at all. A raw text slot spends
// almost every mutation on bytes that stop the file decoding, so the record
// never gets far enough for a minimum to be applied to anything: measured on
// this target, a text slot over eight million executions produced no value
// below the one bound in reach. A float slot is formatted into legal TOML on
// every execution, so what is fuzzed there is the number the schema has to
// judge rather than the decoder's ability to read it. Both surfaces are wanted
// and they are different surfaces.
func FuzzRecordValues(f *testing.F) {
	set := load(f)
	f.Add("example-quantity", "example unit", "1900", "1.5", 0.1)
	f.Add("", "", "", "", 0.0)
	f.Add("Example Quantity", "m s^-2", "1900-01-32", "1e309", 1e-309)
	f.Add(strings.Repeat("q", 61), "\n", "1900-02-30", "0x1p-1", 0.1)
	f.Add("example--quantity", "\u00b5m", "01-01-1900", "nan", 0.5)

	// A half-width below the minimum the schema fixes for it. This one is a
	// crash the fuzzer found rather than a case somebody thought of: the
	// defect it exposed was the site id in the below-minimum refusal
	// misspelled by one character, which reaches newRefusal with an id no
	// catalogue entry matches and panics there. It came back minimised to the
	// negative number and four empty-ish neighbours, and it is kept because
	// the branch it reaches is the only bound in this target's reach.
	f.Add("0", "0", "0", "0", -82.5)

	f.Fuzz(func(t *testing.T, quantity, unit, date, number string, halfWidth float64) {
		body := wellFormed
		for _, edit := range [][2]string{
			{"quantity = \"example-quantity\"", "quantity = " + strconv.Quote(quantity)},
			{"unit = \"example unit\"", "unit = " + strconv.Quote(unit)},
			{"date = \"1900\"", "date = " + strconv.Quote(date)},
			{"value = 1.5", "value = " + number},
			{"plus = 0.1", "plus = " + tomlFloat(halfWidth)},
		} {
			if !strings.Contains(body, edit[0]) {
				t.Fatalf("the fixture no longer contains %q, so this target splices nothing", edit[0])
			}
			body = strings.Replace(body, edit[0], edit[1], 1)
		}
		check(t, fixturePath, []byte(body), set)
	})
}

// tomlFloat writes a float the way TOML spells it.
//
// The three special values are spelled out rather than left to the standard
// library, whose NaN and Inf are not TOML and would turn every one of them
// into a parse failure. They are the values worth keeping: a minimum is a
// comparison, and a comparison against a value that is not ordered is where a
// bound quietly stops holding.
func tomlFloat(v float64) string {
	switch {
	case math.IsNaN(v):
		return "nan"
	case math.IsInf(v, 1):
		return "inf"
	case math.IsInf(v, -1):
		return "-inf"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// FuzzRecordPath is the path target.
//
// A record's identity is its path, and the path is written by whoever files
// the record. Corpus turns an accepted path into a file read, so a pattern
// that accepted a parent traversal or an absolute path would be a reader
// walking out of the tree on input somebody else chose. The property is that
// acceptance implies confinement, and it is asserted against the standard
// library's own answer rather than against a second pattern written here.
func FuzzRecordPath(f *testing.F) {
	set := load(f)
	rule := set.Any()
	if rule == nil {
		f.Fatal("the schema set carries no version, so no path rule could be fuzzed")
	}

	f.Add(fixturePath)
	f.Add("record/example-quantity/1900-example-01.TOML")
	f.Add("record/../record/example-quantity/1900-example-01.toml")
	f.Add("record/example-quantity/../../1900-example-01.toml")
	f.Add("record/_example/1900-example-01.toml")
	f.Add("record//example-quantity/1900-example-01.toml")

	f.Fuzz(func(t *testing.T, path string) {
		if !rule.RecordPath.MatchString(path) {
			return
		}
		if !strings.HasPrefix(path, RecordDir+"/") {
			t.Fatalf("the path rule accepted %q, which is not under %s/", path, RecordDir)
		}
		if !filepath.IsLocal(filepath.FromSlash(path)) {
			t.Fatalf("the path rule accepted %q, which does not stay inside the tree it is joined onto", path)
		}
	})
}
