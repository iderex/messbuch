// Package schema reads the machine readable schema files under schema/.
//
// schema/record-<n>.toml is the single authority for what a measurement record
// of version <n> may contain. This package turns that file into something a
// program can evaluate and does not restate any of it: no field name, no
// presence rule, no closed value set and no pattern appears in this source.
// Adding a field to the schema file therefore changes what the validator
// refuses without changing a line of Go, which is the property
// schema/README.md argues the file exists for.
//
// It fails closed on its own input. A key in a schema file that this package
// does not read is refused rather than skipped, because a rule written in the
// authority and understood by nothing is worse than an absent one: it reads as
// enforced to everybody who opens the file.
package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Dir is the directory the schema files live in, relative to the repository
// root.
const Dir = "schema"

// fileNamePattern matches a schema file and captures its version, so the set
// of versions comes from the tree rather than from a list in this source.
var fileNamePattern = regexp.MustCompile(`^record-([0-9]+)\.toml$`)

// A Set is every schema version this tree carries a file for.
//
// A record names the version it was written against and is read against that
// version's file. schema/README.md fixes that no version's file is ever
// deleted, so a set that has lost one is a corpus that has lost the ability to
// read part of itself.
type Set struct {
	Versions map[int]*Schema
}

// A Schema is one version of the record field set.
type Schema struct {
	// Version is the number a record's schema_version has to name to be read
	// against this file.
	Version int

	// Path is where the file was read from, so a refusal can name it.
	Path string

	RecordPath              *regexp.Regexp
	ExcludedDirectoryPrefix string
	Identifier              *regexp.Regexp
	IdentifierMaxLength     int
	Coverage                *regexp.Regexp

	// Fields are the top-level fields, each addressed by its TOML path inside
	// a record.
	Fields []*Field

	// Composites are the shapes a field can declare, keyed by the name a type
	// reference uses.
	Composites map[string]*Composite

	// Scalars are the named scalar types, keyed the same way.
	Scalars map[string]*Scalar
}

// A Field is one entry of the schema: a top-level field addressed by Path, or
// a member of a composite addressed by Name.
//
// Every key a schema file writes on an entry appears here, including the ones
// no check reads. They are carried rather than dropped so that the loader can
// refuse a key nobody read, and each unread key says in its comment why it is
// unread.
type Field struct {
	Path     string   `toml:"path"`
	Name     string   `toml:"name"`
	Type     string   `toml:"type"`
	Presence string   `toml:"presence"`
	Values   []string `toml:"values"`
	Literals []string `toml:"literals"`

	MinLength int      `toml:"min_length"`
	Minimum   *float64 `toml:"minimum"`

	RequiredWhen *Condition `toml:"required_when"`
	RefusedWhen  *Condition `toml:"refused_when"`

	// MachineDecidable false marks a condition no reading of a record can
	// evaluate. The structural leg leaves such a field alone rather than
	// guessing, which is why the flag is read here and not merely stored.
	MachineDecidable *bool  `toml:"machine_decidable"`
	ConditionNote    string `toml:"condition_note"`

	// Read by nothing here. ResolvesTo is a corpus-wide question rather than a
	// question about one file, and belongs to the meaning leg.
	ResolvesTo string `toml:"resolves_to"`

	// Read by nothing here. These are statements to a reader about why a field
	// is shaped as it is; none of them narrows what a file may contain.
	Means                string   `toml:"means"`
	DoesNotMean          string   `toml:"does_not_mean"`
	FixedBy              string   `toml:"fixed_by"`
	NeededBy             []string `toml:"needed_by"`
	OptionalBecause      string   `toml:"optional_because"`
	Structural           bool     `toml:"structural"`
	AbsenceIsWritten     bool     `toml:"absence_is_written"`
	FreeText             bool     `toml:"free_text"`
	ReadableByAnalysis   *bool    `toml:"readable_by_analysis"`
	NamedHere            bool     `toml:"named_here"`
	AppendOnly           bool     `toml:"append_only"`
	SyntaxCheckedAgainst string   `toml:"syntax_checked_against"`
}

// A Condition is a presence rule written as data.
//
// Field names a field of the record: a path containing a dot is read from the
// record's root, and a bare name is read from the block the entry sits in.
// VocabularyField and SourceStated name something no record states on its own,
// and each is declared in the file rather than left for a checker to discover.
type Condition struct {
	Field           string `toml:"field"`
	VocabularyField string `toml:"vocabulary_field"`
	SourceStated    *bool  `toml:"source_stated"`

	Equals    any   `toml:"equals"`
	NotEquals any   `toml:"not_equals"`
	In        []any `toml:"in"`
	Present   *bool `toml:"present"`
	Absent    *bool `toml:"absent"`
	Never     *bool `toml:"never"`

	// Read by nothing here; it is prose beside the condition.
	AndNote string `toml:"and_note"`
}

// A Composite is a shape one or more fields declare as their type.
type Composite struct {
	Members []*Field `toml:"member"`

	// AtLeastOneOf names members of which at least one has to be present. It is
	// the locator's rule and it is data rather than a branch in this package.
	AtLeastOneOf []string `toml:"at_least_one_of"`

	// Read by nothing here.
	FixedBy string `toml:"fixed_by"`
	Note    string `toml:"note"`
}

// A Scalar is a named scalar type: a pattern, a set of date formats, or a pair
// of members.
type Scalar struct {
	PatternKey   string   `toml:"pattern_key"`
	MaxLengthKey string   `toml:"max_length_key"`
	Formats      []string `toml:"formats"`
	Members      []string `toml:"members"`
	MemberType   string   `toml:"member_type"`

	// Read by nothing here.
	Note string `toml:"note"`
}

// schemaFile is the whole of one schema file, in the shape TOML decodes into.
type schemaFile struct {
	SchemaVersion           int                   `toml:"schema_version"`
	FixedOn                 string                `toml:"fixed_on"`
	RecordPathPattern       string                `toml:"record_path_pattern"`
	ExcludedDirectoryPrefix string                `toml:"excluded_directory_prefix"`
	IdentifierPattern       string                `toml:"identifier_pattern"`
	IdentifierMaxLength     int                   `toml:"identifier_max_length"`
	CoveragePattern         string                `toml:"coverage_pattern"`
	Field                   []*Field              `toml:"field"`
	Type                    map[string]*Composite `toml:"type"`
	Scalar                  map[string]*Scalar    `toml:"scalar"`
}

// Load reads every schema file under root/schema.
//
// It fails closed in three directions. A directory it cannot walk is a refusal
// rather than an empty set read as a complete one. A file whose name says one
// version and whose schema_version says another is a refusal, because the two
// are the same claim and a record naming a version has to reach one file. And
// a set holding no version at all is a refusal, since a validator with no
// authority to read against accepts everything.
func Load(root string) (*Set, error) {
	dir := filepath.Join(root, Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s, so the schema set is unknown: %w", filepath.ToSlash(dir), err)
	}

	set := &Set{Versions: map[int]*Schema{}}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		m := fileNamePattern.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		named, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("%s: the version in the file name is not a number: %w", entry.Name(), err)
		}
		rel := filepath.ToSlash(filepath.Join(Dir, entry.Name()))
		s, err := loadOne(filepath.Join(dir, entry.Name()), rel)
		if err != nil {
			return nil, err
		}
		if s.Version != named {
			return nil, fmt.Errorf("%s: the file name says version %d and schema_version says %d; a record naming a version has to reach exactly one file", rel, named, s.Version)
		}
		set.Versions[s.Version] = s
	}
	if len(set.Versions) == 0 {
		return nil, fmt.Errorf("no schema/record-<n>.toml under %s, so there is nothing for a record to be read against", filepath.ToSlash(dir))
	}
	return set, nil
}

// loadOne reads one schema file and refuses anything in it this package would
// not have applied.
func loadOne(path, rel string) (*Schema, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", rel, err)
	}
	var f schemaFile
	md, err := toml.Decode(string(b), &f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rel, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		var keys []string
		for _, k := range undecoded {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("%s: %d key(s) this validator does not read: %s. A rule written in the authority and applied by nothing reads as enforced to everybody who opens the file, so it is refused here rather than skipped",
			rel, len(keys), strings.Join(keys, ", "))
	}

	s := &Schema{
		Version:                 f.SchemaVersion,
		Path:                    rel,
		ExcludedDirectoryPrefix: f.ExcludedDirectoryPrefix,
		IdentifierMaxLength:     f.IdentifierMaxLength,
		Fields:                  f.Field,
		Composites:              f.Type,
		Scalars:                 f.Scalar,
	}
	if s.Composites == nil {
		s.Composites = map[string]*Composite{}
	}
	if s.Scalars == nil {
		s.Scalars = map[string]*Scalar{}
	}

	for _, p := range []struct {
		key   string
		text  string
		store **regexp.Regexp
	}{
		{"record_path_pattern", f.RecordPathPattern, &s.RecordPath},
		{"identifier_pattern", f.IdentifierPattern, &s.Identifier},
		{"coverage_pattern", f.CoveragePattern, &s.Coverage},
	} {
		if p.text == "" {
			return nil, fmt.Errorf("%s: %s is empty, and a pattern nothing can fail is a rule that is not there", rel, p.key)
		}
		compiled, err := regexp.Compile(p.text)
		if err != nil {
			return nil, fmt.Errorf("%s: %s does not compile: %w", rel, p.key, err)
		}
		*p.store = compiled
	}

	if err := s.check(); err != nil {
		return nil, err
	}
	return s, nil
}

// check refuses a schema whose own entries do not hold together.
//
// The failure it prevents is a validator that reads a broken authority and
// reports a clean corpus: a field naming a type that is defined nowhere would
// otherwise be a field nothing checks, and it would look exactly like a field
// that passed.
func (s *Schema) check() error {
	if s.Version <= 0 {
		return fmt.Errorf("%s: schema_version is %d, and a record cannot name a version that is not a positive number", s.Path, s.Version)
	}
	if s.IdentifierMaxLength <= 0 {
		return fmt.Errorf("%s: identifier_max_length is %d", s.Path, s.IdentifierMaxLength)
	}
	if len(s.Fields) == 0 {
		return fmt.Errorf("%s: no field, so every record would be empty and legal", s.Path)
	}

	seen := map[string]bool{}
	for _, f := range s.Fields {
		if f.Path == "" {
			return fmt.Errorf("%s: a field entry carries no path", s.Path)
		}
		if seen[f.Path] {
			return fmt.Errorf("%s: %s is declared twice, so which entry decides it is whichever the loader read last", s.Path, f.Path)
		}
		seen[f.Path] = true
		if err := s.checkEntry(f, f.Path); err != nil {
			return err
		}
	}
	for name, c := range s.Composites {
		if len(c.Members) == 0 {
			return fmt.Errorf("%s: composite type %s declares no member", s.Path, name)
		}
		members := map[string]bool{}
		for _, m := range c.Members {
			if m.Name == "" {
				return fmt.Errorf("%s: a member of composite type %s carries no name", s.Path, name)
			}
			if members[m.Name] {
				return fmt.Errorf("%s: composite type %s declares member %s twice", s.Path, name, m.Name)
			}
			members[m.Name] = true
			if err := s.checkEntry(m, name+"."+m.Name); err != nil {
				return err
			}
		}
		for _, one := range c.AtLeastOneOf {
			if !members[one] {
				return fmt.Errorf("%s: composite type %s requires one of %s and declares no member of that name", s.Path, name, one)
			}
		}
	}
	return nil
}

// checkEntry refuses one entry whose type or presence rule this package could
// not apply.
func (s *Schema) checkEntry(f *Field, where string) error {
	switch f.Presence {
	case "required", "optional":
	case "conditional":
		if f.RequiredWhen == nil && f.RefusedWhen == nil {
			return fmt.Errorf("%s: %s is conditional and carries no condition", s.Path, where)
		}
	default:
		return fmt.Errorf("%s: %s carries presence %q, which is not required, optional or conditional", s.Path, where, f.Presence)
	}
	if _, _, err := s.resolveType(f.Type); err != nil {
		return fmt.Errorf("%s: %s: %w", s.Path, where, err)
	}
	return nil
}

// LiteralAlternative returns the base type of a `<type>-or-literal` name and
// reports whether the name was one.
//
// The suffix is a convention of the schema file rather than a rule this
// package holds. schema/README.md says a type is a scalar named under
// [scalar.*] or a composite named under [type.*], and
// measurement.denominator_quantity declares identifier-or-literal, which is
// neither. Reading the suffix here keeps the alternatives where the schema
// writes them, in the field's own literals list, instead of putting a field
// name in this source. The disagreement between the file and its own README is
// left where it is rather than repaired by a validator.
func LiteralAlternative(typ string) (string, bool) {
	const suffix = "-or-literal"
	if strings.HasSuffix(typ, suffix) && len(typ) > len(suffix) {
		return typ[:len(typ)-len(suffix)], true
	}
	return "", false
}

// ElementType returns the member type of a list type and reports whether the
// type was a list at all.
func ElementType(typ string) (string, bool) {
	if strings.HasPrefix(typ, "list<") && strings.HasSuffix(typ, ">") {
		return typ[len("list<") : len(typ)-1], true
	}
	return "", false
}

// resolveType reports what a type name refers to: a composite, a scalar, or
// one of the types the file uses without naming under [scalar.*].
//
// A name it cannot place is an error rather than a type nothing checks.
func (s *Schema) resolveType(typ string) (*Composite, *Scalar, error) {
	if typ == "" {
		return nil, nil, fmt.Errorf("no type")
	}
	if elem, ok := ElementType(typ); ok {
		return s.resolveType(elem)
	}
	if base, ok := LiteralAlternative(typ); ok {
		return s.resolveType(base)
	}
	if c, ok := s.Composites[typ]; ok {
		return c, nil, nil
	}
	if sc, ok := s.Scalars[typ]; ok {
		return nil, sc, nil
	}
	switch typ {
	case "string", "integer", "boolean":
		return nil, nil, nil
	}
	return nil, nil, fmt.Errorf("type %q is defined nowhere in this file, so nothing would check a value written under it", typ)
}

// CompositeType returns the composite type of that name, or nil.
func (s *Schema) CompositeType(typ string) *Composite {
	if elem, ok := ElementType(typ); ok {
		typ = elem
	}
	return s.Composites[typ]
}

// ScalarType returns the named scalar type, or nil.
func (s *Schema) ScalarType(typ string) *Scalar {
	if elem, ok := ElementType(typ); ok {
		typ = elem
	}
	return s.Scalars[typ]
}

// Pattern returns the compiled pattern a scalar type names, or nil where it
// names none.
func (s *Schema) Pattern(sc *Scalar) *regexp.Regexp {
	switch sc.PatternKey {
	case "record_path_pattern":
		return s.RecordPath
	case "identifier_pattern":
		return s.Identifier
	case "coverage_pattern":
		return s.Coverage
	}
	return nil
}

// SortedVersions returns the versions in the set in order, so a line printed
// about the set does not depend on map iteration.
func (set *Set) SortedVersions() []int {
	out := make([]int, 0, len(set.Versions))
	for v := range set.Versions {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

// Any returns one schema from the set, the lowest version, for the callers that
// need a rule the versions share rather than a version's field list. The
// excluded-directory prefix and the record path pattern are such rules: they
// decide which files are records at all, before any record has said which
// version it was written against.
func (set *Set) Any() *Schema {
	versions := set.SortedVersions()
	if len(versions) == 0 {
		return nil
	}
	return set.Versions[versions[0]]
}
