// Package compat refuses a schema change that makes an existing corpus
// unreadable, or readable and differently interpreted.
//
// The parity map names this as the stand-in for the compatibility build on the
// board this gate was copied from. There the risk is a plugin that no longer
// loads into its host. Here it is worse than that, because the failure has a
// quiet form: a corpus that still loads and now means something else.
//
// # The two directions
//
// A new tool reading an old corpus has to keep working. Released artifacts stay
// in circulation and somebody will point last year's data at this year's
// binary.
//
// An old tool reading a new corpus has to fail loudly rather than silently
// misread it. That half is already refused by the validator: an unknown field
// is a refusal rather than something dropped, and a schema version this tree
// carries no file for is a refusal rather than a guess. What this package adds
// is the frozen evidence that both go on being refused.
//
// # What is frozen, and why it is a reading rather than a file
//
// Under testdata/schema-<n>/ sits a small corpus written against schema version
// n, and beside it readings.json, which is what this tree made of those bytes
// on the day the version was frozen. A fixture that is merely re-validated
// proves that it still passes. A fixture whose READING is compared proves that
// it still means the same thing, and the difference is the whole issue: a unit
// convention that changes, a field that starts being read from somewhere else
// and a value that starts being decoded as a different type all leave a corpus
// that still validates.
//
// The expected readings were produced by this code and that is deliberate
// rather than an oversight. A test whose expected value came from the code
// under test proves only that the code has not changed, and here that is
// exactly the property wanted: this check exists to refuse an unintended change
// in interpretation. What it cannot do is say the reading was right on the day
// it was frozen, and nothing here claims it.
//
// # A deliberate break is allowed and is not what this refuses
//
// Breaking compatibility on purpose means bumping the schema version, adding a
// directory here at the old version, and writing what the break is in the
// changelog. This check is what makes those steps unavoidable rather than
// optional, and it refuses the accidental break rather than the argued one.
package compat

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/iderex/messbuch/internal/schema"
	"github.com/iderex/messbuch/internal/validate"
)

// Dir is where the frozen corpora live, relative to the repository root.
const Dir = "internal/compat/testdata"

// ReadingsName is the frozen reading beside each frozen corpus.
const ReadingsName = "readings.json"

// versionDir is the name of a frozen corpus directory, which carries the
// schema version in its own name so that neither the file list nor a second
// list has to say which version a fixture was written against.
const versionPrefix = "schema-"

// A Reading is what this tree makes of one frozen file.
//
// Exactly one of Fields and Refusals is set. A file that earns a refusal has no
// reading, and a file that earns none has no refusal, which is the validator's
// own shape: acceptance is the outcome with no second chance.
type Reading struct {
	// Fields is the file decoded, for a file the validator accepts.
	Fields map[string]any `json:"fields,omitempty"`

	// Refusals is the site and field of each refusal, for a file it does not.
	// The found and expected texts are deliberately not frozen: they are
	// sentences written for a person, and freezing them would turn a reworded
	// message into a compatibility break.
	Refusals []string `json:"refusals,omitempty"`
}

// A Version is one frozen corpus and what this tree makes of it now.
type Version struct {
	Schema   int
	Files    []string
	Readings map[string]Reading
}

// Report is what one pass over every frozen corpus found.
type Report struct {
	Versions []Version

	// Differences is one line per file whose reading moved, empty where none
	// did.
	Differences []string
}

// Check reads every frozen corpus against the schema this tree carries now and
// compares each reading with the one frozen beside it.
//
// It fails closed in three directions. A frozen directory that cannot be read,
// a readings file that cannot be parsed, and a set of frozen corpora that is
// empty are all errors rather than a clean pass: a check with no fixture
// examines nothing and would report the same green line forever.
func Check(root string) (*Report, error) {
	set, err := schema.Load(root)
	if err != nil {
		return nil, err
	}

	dirs, err := versionDirs(root)
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no frozen corpus under %s, so this examined nothing; a compatibility check with no fixture is a green line about nothing", Dir)
	}

	report := &Report{}
	for _, version := range dirs {
		now, files, err := read(root, version, set)
		if err != nil {
			return nil, err
		}
		frozen, err := frozenReadings(root, version)
		if err != nil {
			return nil, err
		}
		report.Versions = append(report.Versions, Version{Schema: version, Files: files, Readings: now})
		report.Differences = append(report.Differences, compare(version, frozen, now)...)
	}
	return report, nil
}

// Freeze returns the readings file's bytes for one version, which is how a
// frozen corpus is created and how this package's own suite proves that a
// moved reading is caught rather than described.
func Freeze(root string, version int) ([]byte, error) {
	set, err := schema.Load(root)
	if err != nil {
		return nil, err
	}
	readings, _, err := read(root, version, set)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(readings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cannot render the readings of schema version %d: %w", version, err)
	}
	return append(out, '\n'), nil
}

// versionDirs is every frozen schema version, in ascending order.
func versionDirs(root string) ([]int, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(Dir)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot read %s, so which schema versions are frozen is unknown rather than none: %w", Dir, err)
	}

	var versions []int
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), versionPrefix) {
			continue
		}
		version, convErr := strconv.Atoi(strings.TrimPrefix(entry.Name(), versionPrefix))
		if convErr != nil {
			return nil, fmt.Errorf("%s/%s does not name a schema version, so nothing could say which schema its files were written against", Dir, entry.Name())
		}
		versions = append(versions, version)
	}
	sort.Ints(versions)
	return versions, nil
}

// read makes this tree's reading of every file frozen at one version.
func read(root string, version int, set *schema.Set) (map[string]Reading, []string, error) {
	dir := filepath.Join(root, filepath.FromSlash(Dir), fmt.Sprintf("%s%d", versionPrefix, version))
	readings := map[string]Reading{}
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".toml") {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		name := filepath.ToSlash(rel)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		// The path handed to the validator is the record path the file claims
		// rather than the path it is stored at, because a fixture under
		// testdata is not at a record path and the refusal about that is a
		// fixture of its own rather than something every file here earns.
		refusals := validate.Record("record/"+name, content, set)
		if len(refusals) > 0 {
			sites := make([]string, 0, len(refusals))
			for _, r := range refusals {
				sites = append(sites, r.Site+" at "+r.Field)
			}
			readings[name] = Reading{Refusals: sites}
			files = append(files, name)
			return nil
		}

		fields := map[string]any{}
		if _, decodeErr := toml.Decode(string(content), &fields); decodeErr != nil {
			return fmt.Errorf("%s earned no refusal and does not decode, which is a defect in the validator rather than in the fixture: %w", name, decodeErr)
		}
		readings[name] = Reading{Fields: fields}
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read the corpus frozen at schema version %d: %w", version, err)
	}
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("the corpus frozen at schema version %d holds no file, so it proves nothing about that version", version)
	}
	sort.Strings(files)
	return readings, files, nil
}

// frozenReadings loads the readings committed beside one frozen corpus.
func frozenReadings(root string, version int) (map[string]Reading, error) {
	path := filepath.Join(root, filepath.FromSlash(Dir), fmt.Sprintf("%s%d", versionPrefix, version), ReadingsName)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the frozen readings of schema version %d, so there is nothing to compare this tree's reading against: %w", version, err)
	}
	readings := map[string]Reading{}
	if err := json.Unmarshal(content, &readings); err != nil {
		return nil, fmt.Errorf("the frozen readings of schema version %d do not parse: %w", version, err)
	}
	return readings, nil
}

// compare returns one line per file whose reading moved.
//
// It reports in both directions. A file that has gone missing from the frozen
// corpus and one that was added without being frozen are both differences,
// because a compatibility check that only compares the files it finds in both
// places can be silenced by deleting a fixture.
func compare(version int, frozen, now map[string]Reading) []string {
	var differences []string
	for name, want := range frozen {
		got, present := now[name]
		if !present {
			differences = append(differences, fmt.Sprintf("schema version %d: %s is frozen and is no longer in the corpus, so nothing reads it any more", version, name))
			continue
		}
		if line := difference(version, name, want, got); line != "" {
			differences = append(differences, line)
		}
	}
	for name := range now {
		if _, present := frozen[name]; !present {
			differences = append(differences, fmt.Sprintf("schema version %d: %s is in the corpus and has no frozen reading, so nothing would notice its meaning moving", version, name))
		}
	}
	sort.Strings(differences)
	return differences
}

// difference says how one file's reading moved, in the terms a reader can act
// on rather than as two blobs of JSON.
func difference(version int, name string, want, got Reading) string {
	if len(want.Refusals) > 0 || len(got.Refusals) > 0 {
		if strings.Join(want.Refusals, ", ") == strings.Join(got.Refusals, ", ") {
			return ""
		}
		return fmt.Sprintf("schema version %d: %s was %s and is now %s",
			version, name, refusalList(want.Refusals), refusalList(got.Refusals))
	}

	wantJSON, gotJSON := render(want.Fields), render(got.Fields)
	if wantJSON == gotJSON {
		return ""
	}
	return fmt.Sprintf("schema version %d: %s reads differently now.\n    frozen: %s\n    now:    %s", version, name, wantJSON, gotJSON)
}

// refusalList renders a refusal set for a sentence, and says so where it is
// empty rather than printing nothing.
func refusalList(sites []string) string {
	if len(sites) == 0 {
		return "accepted"
	}
	return "refused for " + strings.Join(sites, ", ")
}

// render puts a reading in a form two readings can be compared by. The keys are
// sorted by the encoder, so the comparison is about the values rather than
// about map iteration.
func render(fields map[string]any) string {
	out, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf("a reading that cannot be rendered: %v", err)
	}
	return string(out)
}

// FreezeAll rewrites the readings file of every frozen schema version and
// returns the paths written.
//
// It is what `go run . freeze` calls. Rewriting rather than appending, because
// a readings file is what this tree makes of the fixtures now, and a stale
// entry beside a fresh one would be two answers to one question.
func FreezeAll(root string) ([]string, error) {
	versions, err := versionDirs(root)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no frozen corpus under %s, so there is nothing to freeze", Dir)
	}

	var written []string
	for _, version := range versions {
		content, freezeErr := Freeze(root, version)
		if freezeErr != nil {
			return written, freezeErr
		}
		rel := fmt.Sprintf("%s/%s%d/%s", Dir, versionPrefix, version, ReadingsName)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), content, 0o644); err != nil {
			return written, fmt.Errorf("cannot write %s: %w", rel, err)
		}
		written = append(written, rel)
	}
	return written, nil
}
