// Package validate holds the structural leg of the corpus validator: is this
// file a well formed record at all.
//
// It answers that question and no other. Whether the value matches the source,
// whether the quantity a record names has a vocabulary entry, whether a group
// identifier resolves to a registry file and whether a conversion factor is
// the right one are questions about meaning, and they are owed by #25 rather
// than refused here. The line between the two is worth stating rather than
// leaving to a reader: this package reads one file against schema/record-<n>.
// toml and never reads a second file to decide the first.
//
// Two properties matter more than the length of the refusal list.
//
// Every refusal names the file, the field, what was found and what was
// expected. A contributor transcribing a measurement is doing careful,
// unglamorous work, and a refusal that says the file is invalid without saying
// where is a punishment rather than a message.
//
// An unknown field is refused rather than ignored. A typo in a field name that
// is silently dropped produces a record that validates and is missing data,
// which is the worst outcome available, because nothing downstream can detect
// it.
//
// Nothing in this file names a field of a record, a value of a closed set or a
// presence rule. All of it is read from the schema, so a field added there is
// checked here without a line of Go moving.
package validate

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/iderex/messbuch/internal/schema"
)

// A Site is one place this package can refuse, with the one line it refuses
// for.
//
// The catalogue is what `go run . refusals` prints. It is here rather than in
// a document because a list of refusals in a document drifts against the code
// that produces them, and the drift is invisible: a reader trusts the document
// precisely where it has stopped being true.
type Site struct {
	ID      string
	Refuses string
}

// Catalogue is every refusal this leg can produce.
//
// A refusal whose id is not here, and an id here that no input reaches, are
// both refused by this package's own suite. That is the accounting at the id
// level. The accounting per refusal SITE, which catches a second branch added
// under an id that already has a fixture, is the gate's refusal-sites leg and
// reads this package's source rather than this list.
var Catalogue = []Site{
	{"record-does-not-parse", "a record file that is not TOML at all"},
	{"record-in-the-wrong-place", "a file under record/ whose path is not the path a record is written at"},
	{"schema-version", "a record that names no schema version, names one that is not a number, or names one this tree carries no file for"},
	{"unknown-field", "a field the schema does not declare, including a misspelling of one it does"},
	{"missing-required-field", "a field the schema requires and the record does not carry"},
	{"conditional-field-missing", "a field another field's value requires, on a record that does not carry it"},
	{"conditional-field-refused", "a field another field's value refuses, on a record that carries it"},
	{"wrong-type", "a value of a different kind from the one the field is declared as"},
	{"malformed-value", "a value of the right kind whose text does not satisfy the type's own syntax"},
	{"value-outside-set", "a value outside the closed set the schema fixes for the field"},
	{"list-too-short", "a list shorter than the schema's minimum for it"},
	{"below-minimum", "a number below the minimum the schema fixes for the field"},
	{"duplicate-in-list", "the same entry written twice in one list"},
	{"no-member-present", "a block that has to carry at least one of a set of members and carries none"},
}

// siteExists is the catalogue as a set, so that a refusal naming an id nobody
// declared is a failure of this package rather than a finding about a record.
var siteExists = func() map[string]bool {
	m := make(map[string]bool, len(Catalogue))
	for _, s := range Catalogue {
		m[s.ID] = true
	}
	return m
}()

// A Refusal is one thing wrong with one record.
type Refusal struct {
	// Site is the catalogue id, so the accounting can be done over refusals
	// rather than over sentences.
	Site string

	// File is the record's path relative to the repository root.
	File string

	// Field is the TOML path inside the record. It is empty where the refusal
	// is about the file rather than about a field in it.
	Field string

	Found    string
	Expected string
}

// newRefusal is the one place a Refusal is made.
//
// Every refusal in this package goes through it, which is what lets the
// accounting over refusal SITES find them by reading the source: a site is a
// call to this function, and the site list is derived rather than remembered.
func newRefusal(site, file, field, found, expected string) Refusal {
	if !siteExists[site] {
		panic("validate: refusal site " + site + " is not in the catalogue, so nothing would print it")
	}
	return Refusal{Site: site, File: file, Field: field, Found: found, Expected: expected}
}

func (r Refusal) String() string {
	where := r.File
	if r.Field != "" {
		where += ": " + r.Field
	}
	return fmt.Sprintf("%s: found %s, expected %s [%s]", where, r.Found, r.Expected, r.Site)
}

// Sort puts refusals in an order that does not depend on map iteration or on
// the order a walk happened to reach them, so two runs over one tree print the
// same bytes.
func Sort(rs []Refusal) {
	sort.Slice(rs, func(i, j int) bool {
		a, b := rs[i], rs[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Field != b.Field {
			return a.Field < b.Field
		}
		if a.Site != b.Site {
			return a.Site < b.Site
		}
		return a.Found < b.Found
	})
}

// Record returns every structural refusal one record file earns.
//
// It returns all of them rather than the first, so that one round of fixes
// clears one file. The exception is a record whose version cannot be
// established: without a schema there is nothing to check the rest against,
// and a list of refusals derived from a guessed version would be worse than a
// short one.
func Record(rel string, content []byte, set *schema.Set) []Refusal {
	var root map[string]any
	if _, err := toml.Decode(string(content), &root); err != nil {
		var parseErr toml.ParseError
		found := err.Error()
		if ok := asParseError(err, &parseErr); ok {
			found = fmt.Sprintf("a parse failure at line %d: %s", parseErr.Position.Line, parseErr.Message)
		}
		return []Refusal{newRefusal("record-does-not-parse", rel, "", found, "a file that decodes as TOML")}
	}

	versionKey := "schema_version"
	raw, present := root[versionKey]
	if !present {
		return []Refusal{newRefusal("schema-version", rel, versionKey, "nothing",
			"one of the schema versions this tree carries a file for: "+versionList(set))}
	}
	version, ok := raw.(int64)
	if !ok {
		return []Refusal{newRefusal("schema-version", rel, versionKey, describe(raw),
			"an integer naming one of "+versionList(set))}
	}
	s, ok := set.Versions[int(version)]
	if !ok {
		return []Refusal{newRefusal("schema-version", rel, versionKey, describe(raw),
			"one of "+versionList(set)+", each of which is a file in this tree; a record naming a version whose file is gone is a record nothing can read")}
	}

	c := &checker{s: s, file: rel, root: root}
	c.declared(root, "", c.tree())
	for _, f := range s.Fields {
		c.entry(f, f.Path, root)
	}
	Sort(c.out)
	return c.out
}

// versionList is the set of versions in a form a refusal can print.
func versionList(set *schema.Set) string {
	var out []string
	for _, v := range set.SortedVersions() {
		out = append(out, fmt.Sprint(v))
	}
	return strings.Join(out, ", ")
}

// asParseError is here because errors.As needs a concrete target and the
// decoder returns its parse error by value.
func asParseError(err error, target *toml.ParseError) bool {
	pe, ok := err.(toml.ParseError)
	if ok {
		*target = pe
	}
	return ok
}

// A checker validates one record against one schema version.
type checker struct {
	s    *schema.Schema
	file string
	root map[string]any
	out  []Refusal
}

func (c *checker) refuse(site, field, found, expected string) {
	c.out = append(c.out, newRefusal(site, c.file, field, found, expected))
}

// A node is one level of the tree of paths the schema declares. An interior
// node carries no field of its own: measurement is a table because
// measurement.quantity is a field, and nothing declares measurement itself.
type node struct {
	children map[string]*node
	field    *schema.Field
}

// tree builds the declared paths into a tree, so that a key in a record can be
// looked up one level at a time.
func (c *checker) tree() *node {
	root := &node{children: map[string]*node{}}
	for _, f := range c.s.Fields {
		at := root
		parts := strings.Split(f.Path, ".")
		for i, part := range parts {
			next, ok := at.children[part]
			if !ok {
				next = &node{children: map[string]*node{}}
				at.children[part] = next
			}
			if i == len(parts)-1 {
				next.field = f
			}
			at = next
		}
	}
	return root
}

// declared refuses a key the schema does not declare, at every depth.
//
// It walks the record rather than the schema, which is the direction that
// catches a typo: a field the schema knows nothing about has no entry to be
// found missing, and would otherwise pass every other check in this file.
func (c *checker) declared(table map[string]any, prefix string, at *node) {
	for _, key := range sortedKeys(table) {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		child, ok := at.children[key]
		if !ok {
			c.refuse("unknown-field", path, "a field this schema does not declare",
				"one of the fields "+c.s.Path+" declares; a field name it does not carry is a transcription that nothing downstream can see is missing")
			continue
		}
		if child.field != nil {
			// The value is checked where its presence is decided, so that a
			// field is not reported twice.
			continue
		}
		sub, ok := table[key].(map[string]any)
		if !ok {
			c.refuse("wrong-type", path, describe(table[key]), "a table, since the schema declares fields underneath it")
			continue
		}
		c.declared(sub, path, child)
	}
}

// entry decides whether a declared top-level field is present when it has to
// be, absent when it must not be, and well formed when it is there.
//
// local is the block a bare name in one of its conditions is read from, which
// for a top-level field is the record itself. Composite members go through
// member instead, because they are addressed inside their own block and a
// member's condition names a sibling.
func (c *checker) entry(f *schema.Field, path string, local map[string]any) {
	value, present := c.at(path)

	switch f.Presence {
	case "required":
		if !present {
			c.refuse("missing-required-field", path, "nothing", "a value; "+c.s.Path+" declares this field required")
			return
		}
	case "optional":
		if !present {
			return
		}
	case "conditional":
		// machine_decidable = false takes the requirement away and leaves the
		// refusal, which is what the schema's own condition_note says of the
		// one field carrying the flag with both halves: a validator can refuse
		// this field on an active record and cannot require it on a withdrawn
		// one. Reading the flag as covering the whole entry would throw away a
		// refusal the schema says is available.
		decidableRequirement := f.MachineDecidable == nil || *f.MachineDecidable
		if holds, decidable := c.holds(f.RequiredWhen, local); decidableRequirement && decidable && holds && !present {
			c.refuse("conditional-field-missing", path, "nothing", "a value, because "+describeCondition(f.RequiredWhen))
			return
		}
		if holds, decidable := c.holds(f.RefusedWhen, local); decidable && holds && present {
			c.refuse("conditional-field-refused", path, describe(value), "no value at all, because "+describeCondition(f.RefusedWhen))
			return
		}
		if !present {
			return
		}
	}
	c.value(path, value, f, local)
}

// at resolves a dotted path from the record's root.
func (c *checker) at(path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = c.root
	for _, part := range parts {
		table, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = table[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// lookup resolves the field a condition names. A name carrying a dot is read
// from the record's root and a bare name from the block the condition sits in,
// which is the rule schema/README.md states about the two shapes.
func (c *checker) lookup(name string, local map[string]any) (any, bool) {
	if strings.Contains(name, ".") {
		return c.at(name)
	}
	v, ok := local[name]
	return v, ok
}

// holds evaluates a condition and says whether it could be evaluated at all.
//
// Two shapes are undecidable from a record and say so in the schema rather
// than leaving a checker to find out: a condition reading the quantity's
// vocabulary entry, which needs a second file and belongs to the meaning leg,
// and a condition about what a source stated, which no file in this tree
// holds.
func (c *checker) holds(cond *schema.Condition, local map[string]any) (result, decidable bool) {
	if cond == nil {
		return false, false
	}
	if cond.Never != nil && *cond.Never {
		return false, true
	}
	if cond.VocabularyField != "" || cond.SourceStated != nil || cond.Field == "" {
		return false, false
	}
	value, present := c.lookup(cond.Field, local)
	switch {
	case cond.Present != nil:
		return present == *cond.Present, true
	case cond.Absent != nil:
		return !present == *cond.Absent, true
	case cond.Equals != nil:
		return present && sameValue(value, cond.Equals), true
	case cond.NotEquals != nil:
		return present && !sameValue(value, cond.NotEquals), true
	case len(cond.In) > 0:
		if !present {
			return false, true
		}
		for _, one := range cond.In {
			if sameValue(value, one) {
				return true, true
			}
		}
		return false, true
	}
	return false, false
}

// describeCondition puts a condition into the sentence a refusal prints, so
// that a contributor is told which other field decided this one.
func describeCondition(cond *schema.Condition) string {
	if cond == nil {
		return "the schema says so"
	}
	switch {
	case cond.Present != nil && *cond.Present:
		return cond.Field + " is present"
	case cond.Present != nil:
		return cond.Field + " is absent"
	case cond.Absent != nil && *cond.Absent:
		return cond.Field + " is absent"
	case cond.Absent != nil:
		return cond.Field + " is present"
	case cond.Equals != nil:
		return fmt.Sprintf("%s is %v", cond.Field, cond.Equals)
	case cond.NotEquals != nil:
		return fmt.Sprintf("%s is not %v", cond.Field, cond.NotEquals)
	case len(cond.In) > 0:
		var parts []string
		for _, one := range cond.In {
			parts = append(parts, fmt.Sprint(one))
		}
		return fmt.Sprintf("%s is one of %s", cond.Field, strings.Join(parts, ", "))
	}
	return "the schema says so"
}

// value checks one present value against the type its field declares.
func (c *checker) value(path string, v any, f *schema.Field, local map[string]any) {
	if elem, isList := schema.ElementType(f.Type); isList {
		items, ok := asList(v)
		if !ok {
			c.refuse("wrong-type", path, describe(v), "a list of "+elem)
			return
		}
		if f.MinLength > 0 && len(items) < f.MinLength {
			c.refuse("list-too-short", path, fmt.Sprintf("a list of %d", len(items)),
				fmt.Sprintf("a list of at least %d, because a list this schema requires and that carries nothing says the same as an absent field and is not the same fact", f.MinLength))
		}
		c.duplicates(path, elem, items)
		for i, item := range items {
			c.ofType(fmt.Sprintf("%s[%d]", path, i), item, elem, f, local)
		}
		return
	}
	c.ofType(path, v, f.Type, f, local)
}

// duplicates refuses the same entry written twice in one list.
//
// A list of strings is exempt, and the exemption is the reason the rule is
// safe: two authors of one paper can carry the same name, and that is a fact
// about the literature. Everywhere else a repeated entry is one statement
// written twice, and nothing downstream can tell it from two statements.
func (c *checker) duplicates(path, elem string, items []any) {
	if elem == "string" {
		return
	}
	seen := map[string]int{}
	for i, item := range items {
		key := fmt.Sprintf("%v", item)
		if first, ok := seen[key]; ok {
			c.refuse("duplicate-in-list", fmt.Sprintf("%s[%d]", path, i), describe(item),
				fmt.Sprintf("an entry that is not already at %s[%d]; one statement written twice cannot be told from two statements by anything reading this record", path, first))
			continue
		}
		seen[key] = i
	}
}

// ofType checks one value against one type name.
func (c *checker) ofType(path string, v any, typ string, f *schema.Field, local map[string]any) {
	if base, ok := schema.LiteralAlternative(typ); ok {
		if text, isText := v.(string); isText {
			for _, literal := range f.Literals {
				if text == literal {
					return
				}
			}
		}
		c.ofType(path, v, base, f, local)
		return
	}

	if composite := c.s.CompositeType(typ); composite != nil {
		block, ok := v.(map[string]any)
		if !ok {
			c.refuse("wrong-type", path, describe(v), "a "+typ+" block")
			return
		}
		c.block(path, typ, composite, block)
		return
	}

	if sc := c.s.ScalarType(typ); sc != nil {
		if pattern := c.s.Pattern(sc); pattern != nil {
			text, ok := v.(string)
			if !ok {
				c.refuse("wrong-type", path, describe(v), "a "+typ+", which is written as a string")
				return
			}
			if !pattern.MatchString(text) {
				c.refuse("malformed-value", path, describe(v), "a "+typ+" matching "+pattern.String())
				return
			}
			if sc.MaxLengthKey == "identifier_max_length" && len(text) > c.s.IdentifierMaxLength {
				c.refuse("malformed-value", path, fmt.Sprintf("a string of %d characters", len(text)),
					fmt.Sprintf("a %s of at most %d", typ, c.s.IdentifierMaxLength))
			}
			return
		}
		if len(sc.Formats) > 0 {
			text, ok := v.(string)
			if !ok {
				c.refuse("wrong-type", path, describe(v), "a "+typ+", which is written as a string")
				return
			}
			for _, format := range sc.Formats {
				if _, err := time.Parse(goLayout(format), text); err == nil {
					return
				}
			}
			c.refuse("malformed-value", path, describe(v),
				"a "+typ+" written as one of "+strings.Join(sc.Formats, ", ")+", and naming a day that exists")
			return
		}
		if len(sc.Members) > 0 {
			block, ok := v.(map[string]any)
			if !ok {
				c.refuse("wrong-type", path, describe(v), "a "+typ+" with the members "+strings.Join(sc.Members, " and "))
				return
			}
			declared := map[string]bool{}
			for _, member := range sc.Members {
				declared[member] = true
				inner, present := block[member]
				if !present {
					c.refuse("missing-required-field", path+"."+member, "nothing",
						"a "+sc.MemberType+"; a "+typ+" carries both of its members, since one end alone is a different fact from a range")
					continue
				}
				c.ofType(path+"."+member, inner, sc.MemberType, f, block)
			}
			for _, key := range sortedKeys(block) {
				if !declared[key] {
					c.refuse("unknown-field", path+"."+key, "a member this schema does not declare",
						"one of "+strings.Join(sc.Members, " and "))
				}
			}
			return
		}
	}

	switch typ {
	case "string":
		text, ok := v.(string)
		if !ok {
			c.refuse("wrong-type", path, describe(v), "a string")
			return
		}
		c.closedSet(path, text, f)
	case "integer":
		if _, ok := v.(int64); !ok {
			c.refuse("wrong-type", path, describe(v), "an integer")
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			c.refuse("wrong-type", path, describe(v), "a boolean")
		}
	case "float":
		number, ok := v.(float64)
		if !ok {
			c.refuse("wrong-type", path, describe(v),
				"a float; a number written without a decimal point is a TOML integer, and the two are different types to everything that reads this file")
			return
		}
		if f.Minimum != nil && number < *f.Minimum {
			c.refuse("below-minimum", path, describe(v), fmt.Sprintf("a float of at least %v", *f.Minimum))
		}
	case "any":
		// The schema says the type is whatever the field named beside it holds,
		// and that field is named in prose rather than as a path. Nothing here
		// can check it, and inventing a rule would be this package deciding a
		// question the schema was not asked.
	default:
		c.refuse("wrong-type", path, describe(v), "a value of type "+typ)
	}
}

// closedSet refuses a value outside the set the schema fixes for the field.
func (c *checker) closedSet(path, text string, f *schema.Field) {
	if len(f.Values) == 0 {
		return
	}
	for _, allowed := range f.Values {
		if text == allowed {
			return
		}
	}
	c.refuse("value-outside-set", path, describe(text),
		"one of "+strings.Join(f.Values, ", ")+"; free text here would quietly destroy the ability to group")
}

// block checks a composite value: its members' presence, its members' values,
// any member the schema does not declare, and the at-least-one-of rule where
// the type carries one.
func (c *checker) block(path, typ string, composite *schema.Composite, block map[string]any) {
	declared := map[string]bool{}
	for _, member := range composite.Members {
		declared[member.Name] = true
	}
	for _, key := range sortedKeys(block) {
		if !declared[key] {
			c.refuse("unknown-field", path+"."+key, "a member this schema does not declare",
				"a member of "+typ+"; a misspelled member is data that silently did not arrive")
		}
	}
	for _, member := range composite.Members {
		c.member(path+"."+member.Name, member, block)
	}
	if len(composite.AtLeastOneOf) > 0 {
		for _, one := range composite.AtLeastOneOf {
			if _, present := block[one]; present {
				return
			}
		}
		c.refuse("no-member-present", path, "a block carrying none of them",
			"at least one of "+strings.Join(composite.AtLeastOneOf, ", "))
	}
}

// member is entry for a composite member, which is addressed by its name
// inside its own block rather than by a path from the record's root.
func (c *checker) member(path string, m *schema.Field, block map[string]any) {
	value, present := block[m.Name]

	switch m.Presence {
	case "required":
		if !present {
			c.refuse("missing-required-field", path, "nothing", "a value; "+c.s.Path+" declares this member required")
			return
		}
	case "optional":
		if !present {
			return
		}
	case "conditional":
		decidableRequirement := m.MachineDecidable == nil || *m.MachineDecidable
		if holds, decidable := c.holds(m.RequiredWhen, block); decidableRequirement && decidable && holds && !present {
			c.refuse("conditional-field-missing", path, "nothing", "a value, because "+describeCondition(m.RequiredWhen))
			return
		}
		if holds, decidable := c.holds(m.RefusedWhen, block); decidable && holds && present {
			c.refuse("conditional-field-refused", path, describe(value), "no value at all, because "+describeCondition(m.RefusedWhen))
			return
		}
		if !present {
			return
		}
	}
	c.value(path, value, m, block)
}

// goLayout turns a format the schema writes for a reader into the reference
// time the standard library parses against.
func goLayout(format string) string {
	replacer := strings.NewReplacer("YYYY", "2006", "MM", "01", "DD", "02")
	return replacer.Replace(format)
}

// sameValue compares a value from a record against a value from a condition.
func sameValue(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// asList accepts every shape the decoder produces for an array, since an array
// of tables and an array of scalars do not arrive as the same Go type.
func asList(v any) ([]any, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

// describe says what was found, in the words a contributor reading a TOML file
// would use.
func describe(v any) string {
	switch value := v.(type) {
	case nil:
		return "nothing"
	case string:
		return fmt.Sprintf("the string %q", value)
	case int64:
		return fmt.Sprintf("the integer %d", value)
	case float64:
		return fmt.Sprintf("the float %v", value)
	case bool:
		return fmt.Sprintf("the boolean %v", value)
	case map[string]any:
		return fmt.Sprintf("a table of %d entries", len(value))
	}
	if items, ok := asList(v); ok {
		return fmt.Sprintf("a list of %d", len(items))
	}
	return fmt.Sprintf("%v", v)
}

// sortedKeys keeps every walk over a table in one order, so that two runs
// print the same refusals in the same places.
func sortedKeys(table map[string]any) []string {
	out := make([]string, 0, len(table))
	for k := range table {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
