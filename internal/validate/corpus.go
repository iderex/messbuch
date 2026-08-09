package validate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iderex/messbuch/internal/schema"
)

// RecordDir is the directory the corpus lives in, relative to the repository
// root. It is the one path this package names, because a walk has to start
// somewhere and the schema's record path pattern is anchored on it.
const RecordDir = "record"

// A Report is what one walk over the corpus found.
type Report struct {
	// Records is how many files were read as records.
	Records int

	// Skipped is how many files were passed over because they sit under a
	// directory the schema excludes.
	Skipped int

	// Excluded names the excluded directories that actually held something, so
	// that a run says what it did not read rather than leaving it to be
	// inferred from a count.
	Excluded []string

	Refusals []Refusal
}

// Corpus walks root/record and validates every record it finds.
//
// It fails closed on the walk. A tree whose file list cannot be read is an
// error rather than an empty corpus read as a clean one, which is the failure
// that makes a green line mean nothing.
//
// An empty corpus is not an error. There is no record in this repository yet,
// and refusing on that would red every run until the first series lands. What
// the caller gets instead is the count, and a count of zero is a sentence a
// reader can act on.
func Corpus(root string, set *schema.Set) (*Report, error) {
	rule := set.Any()
	if rule == nil {
		return nil, fmt.Errorf("the schema set carries no version, so no file could be read as a record")
	}
	dir := filepath.Join(root, RecordDir)
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("cannot read %s, so the corpus is unknown rather than empty: %w", RecordDir, err)
	}

	report := &Report{}
	excluded := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		slashed := filepath.ToSlash(rel)
		if d.IsDir() {
			return nil
		}
		if excludedBy(slashed, rule.ExcludedDirectoryPrefix) {
			report.Skipped++
			excluded[filepath.ToSlash(filepath.Dir(slashed))] = true
			return nil
		}
		return report.file(root, slashed, rule, set)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s, so the set of records is unknown: %w", RecordDir, err)
	}
	for name := range excluded {
		report.Excluded = append(report.Excluded, name)
	}
	sort.Strings(report.Excluded)
	Sort(report.Refusals)
	return report, nil
}

// file reads one file under record/ and either refuses its path or validates
// its contents.
func (r *Report) file(root, rel string, rule *schema.Schema, set *schema.Set) error {
	if !rule.RecordPath.MatchString(rel) {
		r.Refusals = append(r.Refusals, newRefusal("record-in-the-wrong-place", rel, "", "a file at this path",
			"a path matching "+rule.RecordPath.String()+
				"; a record's identity is its path, so a file in the wrong place claims an identity nothing can read"))
		return nil
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", rel, err)
	}
	r.Records++
	r.Refusals = append(r.Refusals, Record(rel, content, set)...)
	return nil
}

// excludedBy reports whether any directory component of a path under record/
// begins with the prefix the schema excludes.
//
// The walk descends into such a directory rather than skipping it, so that the
// files inside can be counted and named. A directory that is stepped over
// leaves a run unable to say whether it held one file or a hundred, and the
// whole point of printing what was examined is that what was not examined is
// visible beside it.
func excludedBy(rel, prefix string) bool {
	if prefix == "" {
		return false
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts[:len(parts)-1] {
		if strings.HasPrefix(part, prefix) {
			return true
		}
	}
	return false
}
