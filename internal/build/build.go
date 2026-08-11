// Package build turns the tracked records into the artifact a consumer loads.
//
// Contributors want files they can read and edit, one record per file, in a
// format a text editor and a diff are useful on. Anybody consuming the corpus
// wants one thing they can open. This package is the step between the two, and
// it is a step rather than a second copy of the data: the records under
// record/ are the authority and the artifact is derived from them every time.
//
// # The artifact is not tracked
//
// #27 left the choice open between a tracked artifact whose bytes a check
// rebuilds and compares, and an untracked one produced at release time. The
// stamp decides it. docs/decisions/0011-corpus-versioning.md requires
// corpus_revision to be the full commit identifier the artifact was built
// from, and a tracked file is part of the commit that adds it, so its bytes
// would have to contain the identifier of a commit that does not exist until
// after those bytes are fixed. Every later commit changes the revision and
// therefore the artifact, so a check comparing the tracked file against a
// rebuild would refuse it on the next commit whatever that commit changed.
//
// The same record has already chosen the same way for a separate reason: a
// release is a tag, the artifacts are built from that tag, and publishing an
// artifact from anything else is not a release.
//
// So the guard changes shape with the choice. There is nothing to compare a
// rebuild against, and what is refused instead is the artifact being tracked
// at all, which is the same failure caught one step earlier: a build output
// committed by somebody in a hurry and then edited by hand. The gate's
// artifact-untracked leg is where that refusal lives.
//
// # Which format is the authority
//
// The JSON artifact is lossless and is the authority. The CSV is a convenience
// view for somebody with a spreadsheet, it cannot carry the uncertainty
// representation, and it says which fields it dropped rather than dropping
// them quietly. Both carry the same stamp, so a file that has been copied
// somewhere else still says what it is.
package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/iderex/messbuch/internal/schema"
	"github.com/iderex/messbuch/internal/validate"
)

// Dir is where the build writes, relative to the repository root.
//
// One directory rather than the root, so that .gitignore names one path and
// the artifact-untracked leg has one place to look.
const Dir = "build"

// The names of what the build writes. The sidecar exists because the `#` lines
// carrying the stamp in a CSV are a convention rather than part of any CSV
// specification, and a reader whose parser chokes on them would otherwise have
// to strip the provenance to reach the data.
const (
	JSONName     = "corpus.json"
	CSVName      = "corpus.csv"
	CSVStampName = "corpus.csv.stamp.json"
)

// StampVersion is the version of the stamp format itself, which is the first
// field of every stamp so that a reader knows what the rest of it means.
const StampVersion = 1

// Outputs is every path the build writes, relative to the repository root, in
// the order it writes them.
//
// The artifact-untracked leg reads this rather than a list of its own, so a
// format added here cannot be added without the guard reaching it.
func Outputs() []string {
	out := make([]string, 0, 3)
	for _, name := range []string{JSONName, CSVName, CSVStampName} {
		out = append(out, Dir+"/"+name)
	}
	return out
}

// A Stamp says what the artifact is, so that a copy of it somewhere else still
// answers the question.
//
// The field names and their meanings are fixed by
// docs/decisions/0011-corpus-versioning.md and are not decided here. There is
// no wall-clock time in it, deliberately: its presence would make byte for
// byte reproduction impossible, and that reproduction is what turns a
// reproduction claim into a check.
type Stamp struct {
	StampVersion   int    `json:"stamp_version"`
	CorpusVersion  string `json:"corpus_version"`
	CorpusRevision string `json:"corpus_revision"`
	CorpusState    string `json:"corpus_state"`
	// Present only when CorpusState is dirty, which is what the record asks
	// for: a build off an edited tree is still identifiable, and a clean one
	// carries no field a reader has to interpret.
	CorpusDirtyDigest string `json:"corpus_dirty_digest,omitempty"`

	ToolVersion  string `json:"tool_version"`
	ToolRevision string `json:"tool_revision"`
	ToolState    string `json:"tool_state"`

	Command string `json:"command"`

	// Options and Selection are empty here and are written rather than left
	// out. The record requires every option in force to be printed, including
	// the ones left at their default, because a default that is not printed
	// can change between versions and silently change a published number. A
	// build takes no analysis option and no selection expression, and an empty
	// object says that, where an absent field would leave a reader guessing
	// whether the build had options nobody printed.
	Options   map[string]string `json:"options"`
	Selection string            `json:"selection"`

	SelectedCount int            `json:"selected_count"`
	Excluded      map[string]int `json:"excluded"`
}

// A Record is one file of the corpus as the artifact carries it.
//
// Fields is the file decoded and nothing else. No field is renamed, dropped or
// interpreted on the way through, which is what makes the JSON artifact the
// authority rather than a view.
type Record struct {
	Path   string         `json:"path"`
	Fields map[string]any `json:"fields"`
}

// An Artifact is one build.
type Artifact struct {
	Stamp   Stamp    `json:"stamp"`
	Records []Record `json:"records"`
}

// A Provenance is what the tree says about itself, read once and passed in.
//
// It is a parameter rather than something Build discovers, so that every test
// in this package states the provenance it is building against instead of
// depending on the state of whatever checkout it happens to run in. A suite
// that reads the real repository proves the state of the tree on the day it
// ran rather than the guard.
type Provenance struct {
	CorpusVersion  string
	CorpusRevision string
	CorpusState    string
	ToolVersion    string
	ToolRevision   string
	ToolState      string
}

// Build reads every record under record/ and returns the artifact.
//
// It refuses to build from a corpus that does not validate. An artifact
// assembled out of records nothing accepted is worse than no artifact: it is
// data with a stamp on it, and the stamp is what a consumer trusts.
//
// It fails closed on the walk for the same reason validate.Corpus does. A tree
// whose files cannot be listed is an error rather than an empty corpus read as
// a clean one.
func Build(root string, set *schema.Set, prov Provenance) (*Artifact, error) {
	rule := set.Any()
	if rule == nil {
		return nil, fmt.Errorf("the schema set carries no version, so no file could be read as a record")
	}

	report, err := validate.Corpus(root, set)
	if err != nil {
		return nil, err
	}
	if len(report.Refusals) > 0 {
		lines := make([]string, 0, len(report.Refusals))
		for _, r := range report.Refusals {
			lines = append(lines, r.String())
		}
		return nil, fmt.Errorf("%d refusal(s) over %d record(s), so there is nothing to build; an artifact assembled out of records nothing accepted is data with a stamp on it:\n  %s",
			len(report.Refusals), report.Records, strings.Join(lines, "\n  "))
	}

	records, excluded, err := read(root, rule)
	if err != nil {
		return nil, err
	}

	stamp := Stamp{
		StampVersion:   StampVersion,
		CorpusVersion:  prov.CorpusVersion,
		CorpusRevision: prov.CorpusRevision,
		CorpusState:    prov.CorpusState,
		ToolVersion:    prov.ToolVersion,
		ToolRevision:   prov.ToolRevision,
		ToolState:      prov.ToolState,
		Command:        "build",
		Options:        map[string]string{},
		Selection:      "",
		SelectedCount:  len(records),
		Excluded:       excluded,
	}
	if stamp.CorpusState == StateDirty {
		stamp.CorpusDirtyDigest = digest(records)
	}
	return &Artifact{Stamp: stamp, Records: records}, nil
}

// read walks record/ and decodes every file whose path is a record path.
//
// A file that is not at a record path is counted as excluded rather than
// refused here, and that is sound only because Build validated first: a file
// under record/ that is neither at a record path nor under an excluded
// directory earns record-in-the-wrong-place, so a clean validation leaves the
// excluded directories as the only way to reach this branch.
func read(root string, rule *schema.Schema) ([]Record, map[string]int, error) {
	dir := filepath.Join(root, validate.RecordDir)
	if _, err := os.Stat(dir); err != nil {
		return nil, nil, fmt.Errorf("cannot read %s, so the corpus is unknown rather than empty: %w", validate.RecordDir, err)
	}

	// An empty slice rather than a nil one, so an empty corpus serialises as an
	// empty list rather than as a null a consumer has to special-case.
	records := []Record{}
	excluded := map[string]int{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		slashed := filepath.ToSlash(rel)
		if !rule.RecordPath.MatchString(slashed) {
			excluded[ExcludedDirectory]++
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		fields := map[string]any{}
		if _, decodeErr := toml.Decode(string(content), &fields); decodeErr != nil {
			return fmt.Errorf("%s decoded during validation and not here, which is a defect in this package rather than in the record: %w", slashed, decodeErr)
		}
		records = append(records, Record{Path: slashed, Fields: fields})
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot walk %s, so the set of records is unknown: %w", validate.RecordDir, err)
	}

	// Path order rather than walk order, so two builds of one tree produce one
	// artifact. The walk is already ordered on every filesystem this runs on
	// and sorting is the cheap way not to depend on that.
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	return records, excluded, nil
}

// digest is a fingerprint of what was built, for a build off an edited tree.
//
// It is over the paths and the raw field content in path order, so two dirty
// builds of two different edits are distinguishable, which is the whole use of
// the field. A clean build carries no digest and the revision answers instead.
func digest(records []Record) string {
	h := sha256.New()
	for _, r := range records {
		fmt.Fprintf(h, "%s\n%v\n", r.Path, r.Fields)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ExcludedDirectory is the one reason a file under record/ is counted as
// excluded rather than built.
//
// One reason rather than a set, and that is a statement about what can reach
// the count rather than a simplification: Build validates first, and a file
// under record/ that is neither at a record path nor under a directory the
// schema excludes earns record-in-the-wrong-place, so a clean validation
// leaves the excluded directories as the only way there.
const ExcludedDirectory = "under-an-excluded-directory"

// The two states a tree is in, in the spelling the stamp uses.
const (
	StateClean = "clean"
	StateDirty = "dirty"
)

// Unreleased is what the stamp carries where no tag names the revision.
const Unreleased = "unreleased"

// corpusPaths and toolPaths are what each half of the stamp is a statement
// about.
//
// Two lists rather than one working-tree check, because the corpus and the
// tool are versioned separately and a stamp saying the corpus is dirty when
// somebody edited a Go file would be a warning nobody could act on. Splitting
// them is also what lets a build write into build/ without calling itself
// dirty.
var (
	corpusPaths = []string{"record", "vocabulary", "group", "schema"}
	toolPaths   = []string{"main.go", "internal", "go.mod", "go.sum"}
)

// GitProvenance reads the tree's own provenance out of git.
//
// It is separate from Build and returns a value rather than writing one, so
// that the one place this package reaches outside itself is a function a test
// does not have to call.
func GitProvenance(root string) (Provenance, error) {
	revision, err := git(root, "rev-parse", "HEAD")
	if err != nil {
		return Provenance{}, fmt.Errorf("cannot read the revision, so no artifact could say what it was built from: %w", err)
	}

	corpusState, err := state(root, corpusPaths)
	if err != nil {
		return Provenance{}, err
	}
	toolState, err := state(root, toolPaths)
	if err != nil {
		return Provenance{}, err
	}

	return Provenance{
		CorpusVersion:  version(root, "corpus-v*", "corpus-v"),
		CorpusRevision: revision,
		CorpusState:    corpusState,
		ToolVersion:    version(root, "v*", "v"),
		ToolRevision:   revision,
		ToolState:      toolState,
	}, nil
}

// version returns the version a tag gives the revision, or unreleased.
//
// docs/decisions/0011-corpus-versioning.md says the version is unreleased when
// the revision is not at or descended from a tag, which is exactly what
// describing the nearest ancestor tag answers. A tree with no such tag is not
// an error: this repository has never been released, so unreleased is the
// ordinary answer rather than a failure.
func version(root, match, prefix string) string {
	tag, err := git(root, "describe", "--tags", "--match", match, "--abbrev=0")
	if err != nil || tag == "" {
		return Unreleased
	}
	return strings.TrimPrefix(tag, prefix)
}

// state reports whether the tracked files under the given paths differ from
// what the revision holds.
func state(root string, paths []string) (string, error) {
	args := append([]string{"status", "--porcelain", "--"}, paths...)
	out, err := git(root, args...)
	if err != nil {
		return "", fmt.Errorf("cannot read the working tree state of %s, so the stamp cannot say whether it is clean: %w", strings.Join(paths, ", "), err)
	}
	if out == "" {
		return StateClean, nil
	}
	return StateDirty, nil
}

// WriteAll writes every format into root/build and returns the paths written.
//
// The directory is created here rather than tracked, which is the choice this
// package's own comment argues, and nothing else in the tree writes into it.
func WriteAll(root string, a *Artifact, s *schema.Schema) ([]string, error) {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create %s, so there is nowhere to write the artifact: %w", Dir, err)
	}

	writers := []struct {
		name  string
		write func(*os.File) error
	}{
		{JSONName, func(f *os.File) error { return WriteJSON(f, a) }},
		{CSVName, func(f *os.File) error { return WriteCSV(f, a, s) }},
		{CSVStampName, func(f *os.File) error { return WriteCSVStamp(f, a, s) }},
	}

	var written []string
	for _, w := range writers {
		path := filepath.Join(dir, w.name)
		file, err := os.Create(path)
		if err != nil {
			return written, fmt.Errorf("cannot write %s: %w", w.name, err)
		}
		if err := w.write(file); err != nil {
			file.Close()
			return written, err
		}
		if err := file.Close(); err != nil {
			return written, fmt.Errorf("cannot close %s: %w", w.name, err)
		}
		written = append(written, Dir+"/"+w.name)
	}
	return written, nil
}

// git runs one git command in the tree and returns its trimmed output.
//
// A subprocess rather than a library, and it is the second program this
// repository shells out to after the Go toolchain. What is wanted here is what
// git itself says about the checkout, and a reimplementation of that reading
// would be a second answer that can disagree with the first.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
