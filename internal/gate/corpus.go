package gate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// corpusDecodesLeg refuses a tracked TOML file that does not decode.
//
// This is the syntax question and only the syntax question: does the file
// parse at all. Whether a record carries the right fields, whether a value
// sits inside a closed set and whether a file is in the right place for what
// it claims to be are the validator's refusals, and they are owed by the
// record-format milestone rather than by this leg. Two refusals for one defect
// would be one refusal too many, so this leg says what it is in its own
// output.
//
// It fails closed. A tree whose file list cannot be walked is a refusal, not
// an empty set read as a clean one.
func corpusDecodesLeg(root string) (string, error) {
	files, err := tomlFiles(root)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no TOML file under %s, so this leg examined nothing", root)
	}
	for _, rel := range files {
		if err := decodesAsTOML(filepath.Join(root, rel)); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%d TOML file(s) decode; field and placement rules are not checked here", len(files)), nil
}

// tomlFiles lists the TOML files under root, relative to it, in a fixed order.
//
// The .git directory is skipped because it is not the corpus, and a path
// component starting with a dot is skipped for the same reason a checkout does
// not carry one as data.
func tomlFiles(root string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(name), ".toml") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s, so the set of TOML files is unknown: %w", root, err)
	}
	sort.Strings(found)
	return found, nil
}

// decodesAsTOML reports whether one file parses, naming the file and the line
// when it does not.
//
// A contributor transcribing a measurement is doing careful, unglamorous work,
// and a refusal that says the file is broken without saying where is a
// punishment rather than a message.
func decodesAsTOML(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	var whatever map[string]any
	if _, err := toml.Decode(string(b), &whatever); err != nil {
		var parseErr toml.ParseError
		if errors.As(err, &parseErr) {
			return fmt.Errorf("%s line %d: %s", filepath.ToSlash(path), parseErr.Position.Line, parseErr.Message)
		}
		return fmt.Errorf("%s: %w", filepath.ToSlash(path), err)
	}
	return nil
}
