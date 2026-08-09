// The readable rendering of the schema.
//
// schema/record-<n>.toml is the authority and it is written for a program to
// read. A contributor transcribing a measurement should not have to read TOML
// entries with a presence vocabulary in them, and should not have to open any
// Go source to find out what a field may contain. `go run . schema` prints the
// same file as sentences.
//
// This renders rather than restates. No field name, no value set, no pattern
// and no presence rule appears in this source: everything printed below comes
// out of the loaded schema, so a field added to the file appears here without
// a line of Go moving, and a rendering that has drifted from the authority is
// not a state this file can be in.
package schema

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Print writes every schema version in the set, in version order.
func Print(w io.Writer, set *Set) error {
	for i, version := range set.SortedVersions() {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := set.Versions[version].print(w); err != nil {
			return err
		}
	}
	return nil
}

// print writes one version.
func (s *Schema) print(w io.Writer) error {
	p := &printer{w: w}

	p.line("schema version %d, fixed on %s, read from %s", s.Version, s.FixedOn, s.Path)
	p.line("")
	p.line("A record is a file whose path matches")
	p.line("    %s", s.RecordPath)
	p.line("and a directory under record/ whose name begins with %q holds no records.", s.ExcludedDirectoryPrefix)
	p.line("")
	p.line("FIELDS, in the order the schema declares them.")
	for _, f := range s.Fields {
		p.line("")
		p.entry(f.Path, f)
	}

	if len(s.Composites) > 0 {
		p.line("")
		p.line("BLOCKS, the shapes a field above declares as its type.")
		for _, name := range sortedNames(s.Composites) {
			c := s.Composites[name]
			p.line("")
			p.line("  %s", name)
			if c.Note != "" {
				p.wrapped("    ", c.Note)
			}
			if len(c.AtLeastOneOf) > 0 {
				p.wrapped("    ", "At least one of "+choices(c.AtLeastOneOf)+" has to be present.")
			}
			if c.FixedBy != "" {
				p.line("    Fixed by %s", c.FixedBy)
			}
			for _, m := range c.Members {
				p.line("")
				p.entry("  "+m.Name, m)
			}
		}
	}

	if len(s.Scalars) > 0 {
		p.line("")
		p.line("VALUE TYPES, what a value of that type has to look like.")
		for _, name := range sortedNames(s.Scalars) {
			sc := s.Scalars[name]
			p.line("")
			p.line("  %s", name)
			p.line("    %s", scalarShape(s, sc))
			if sc.Note != "" {
				p.wrapped("    ", sc.Note)
			}
		}
	}

	return p.err
}

// entry writes one field or one composite member: what it is called, what it
// may hold, when it has to be there, and what it means.
func (p *printer) entry(label string, f *Field) {
	p.line("  %s", label)
	p.line("    %s", strings.Join(shape(f), ", "))
	for _, sentence := range presence(f) {
		p.wrapped("    ", sentence)
	}
	if f.Means != "" {
		p.wrapped("    ", f.Means)
	}
	if f.DoesNotMean != "" {
		p.wrapped("    ", "Does not mean: "+f.DoesNotMean)
	}
	if f.OptionalBecause != "" {
		p.wrapped("    ", "Optional because: "+f.OptionalBecause)
	}
	if f.ResolvesTo != "" {
		p.wrapped("    ", "Resolves to "+f.ResolvesTo+", which the meaning leg reads and this one does not.")
	}
	if f.FixedBy != "" {
		p.line("    Fixed by %s", f.FixedBy)
	}
}

// shape says what a value of this field may be: its type, its closed set, its
// literals and any bound on it.
func shape(f *Field) []string {
	out := []string{typeName(f.Type)}
	if f.MinLength > 0 {
		out = append(out, fmt.Sprintf("with at least %d entr%s", f.MinLength, plural(f.MinLength)))
	}
	if len(f.Values) > 0 {
		out = append(out, "one of "+choices(f.Values))
	}
	if len(f.Literals) > 0 {
		out = append(out, "or one of "+choices(f.Literals))
	}
	if f.Minimum != nil {
		out = append(out, fmt.Sprintf("not below %v", *f.Minimum))
	}
	return out
}

// typeName says a type the way a reader meets it rather than the way the file
// writes it, so that a list and an alternative do not have to be decoded.
//
// A name that is not one of the plain kinds is printed as it stands and is
// described in its own right further down, under BLOCKS or under VALUE TYPES.
// The alternative form drops its suffix here because the literals it allows
// are printed beside it from the field's own list.
func typeName(typ string) string {
	if elem, ok := ElementType(typ); ok {
		return "a list of " + elem
	}
	if base, ok := LiteralAlternative(typ); ok {
		return typeName(base)
	}
	return article(typ) + " " + typ
}

// article is here because the type names come out of the schema file and one
// of them begins with a vowel. Reading it off the name keeps the rendering
// correct for a type this file has never seen.
func article(word string) string {
	if word != "" && strings.ContainsRune("aeiou", rune(word[0])) {
		return "an"
	}
	return "a"
}

// presence turns the presence rule into sentences.
//
// A conditional field carries up to three of them: when it is required, when
// it is refused, and the note the schema writes where no reading of a record
// can decide the condition. That last one is printed rather than dropped,
// because a contributor who is told a field is required and then meets a
// validator that never asks for it learns not to trust either.
func presence(f *Field) []string {
	switch f.Presence {
	case "conditional":
		var out []string
		if f.RequiredWhen != nil {
			out = append(out, "Required when "+f.RequiredWhen.Describe()+".")
		}
		if f.RefusedWhen != nil {
			out = append(out, "Refused when "+f.RefusedWhen.Describe()+".")
		}
		if f.MachineDecidable != nil && !*f.MachineDecidable && f.ConditionNote != "" {
			out = append(out, "Not decidable from the record alone: "+f.ConditionNote)
		}
		return out
	case "":
		return nil
	default:
		return []string{upperFirst(f.Presence) + "."}
	}
}

// Describe puts a condition into the words a refusal and this rendering both
// use, so that the sentence a contributor reads in the printed schema is the
// sentence the validator refuses them with.
func (c *Condition) Describe() string {
	if c == nil {
		return "the schema says so"
	}
	switch {
	case c.Never != nil && *c.Never:
		return "never"
	case c.VocabularyField != "":
		return "the quantity's vocabulary entry says so, in " + c.VocabularyField
	case c.SourceStated != nil && *c.SourceStated:
		return "the source stated one"
	case c.SourceStated != nil:
		return "the source stated none"
	case c.Present != nil && *c.Present:
		return c.Field + " is present"
	case c.Present != nil:
		return c.Field + " is absent"
	case c.Absent != nil && *c.Absent:
		return c.Field + " is absent"
	case c.Absent != nil:
		return c.Field + " is present"
	case c.Equals != nil:
		return fmt.Sprintf("%s is %v", c.Field, c.Equals)
	case c.NotEquals != nil:
		return fmt.Sprintf("%s is not %v", c.Field, c.NotEquals)
	case len(c.In) > 0:
		parts := make([]string, 0, len(c.In))
		for _, one := range c.In {
			parts = append(parts, fmt.Sprint(one))
		}
		return fmt.Sprintf("%s is one of %s", c.Field, strings.Join(parts, ", "))
	}
	return "the schema says so"
}

// scalarShape says what a value of a named scalar type has to look like.
func scalarShape(s *Schema, sc *Scalar) string {
	if pattern := s.Pattern(sc); pattern != nil {
		text := "a string matching " + pattern.String()
		if max, ok := maxLength(s, sc); ok {
			text += fmt.Sprintf(", of at most %d characters", max)
		}
		return text
	}
	if len(sc.Formats) > 0 {
		return "a date written as " + choices(sc.Formats)
	}
	if len(sc.Members) > 0 {
		return "a block with " + list(sc.Members) + ", each a " + sc.MemberType
	}
	return "no syntax beyond what the note below says"
}

// maxLength resolves the bound a scalar type names, where it names one.
func maxLength(s *Schema, sc *Scalar) (int, bool) {
	if sc.MaxLengthKey == "identifier_max_length" {
		return s.IdentifierMaxLength, true
	}
	return 0, false
}

// A printer holds the first write error, so that every call site below can
// stay a statement rather than a branch, and a closed pipe is still reported.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) line(format string, args ...any) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, format+"\n", args...)
}

// wrapped prints a sentence at a width a terminal shows without folding it
// somewhere the reader did not choose.
func (p *printer) wrapped(indent, text string) {
	const width = 76
	line := indent
	for _, word := range strings.Fields(text) {
		switch {
		case line == indent:
			line += word
		case len(line)+1+len(word) > width:
			p.line("%s", line)
			line = indent + word
		default:
			line += " " + word
		}
	}
	if line != indent {
		p.line("%s", line)
	}
}

// choices writes a set of alternatives, of which a value is one.
func choices(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " or " + items[len(items)-1]
}

// list writes a set of things that hold together rather than alternatives.
func list(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

func upperFirst(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

// sortedNames keeps a map's entries in one order, so two runs print the same
// bytes.
func sortedNames[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for name := range m {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
