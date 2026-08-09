package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimal is a schema file small enough to read in one go and complete enough
// to load. Every test below is one edit away from it, so the refusal a test
// earns can be attributed to that edit.
//
// INVENTED. The field it declares is not a field of a record and the patterns
// are not this project's patterns.
const minimal = `
schema_version = 1
fixed_on = "1900-01-01"
record_path_pattern = '^record/x/y\.toml$'
excluded_directory_prefix = "_"
identifier_pattern = '^[a-z]+$'
identifier_max_length = 4
coverage_pattern = '^k=1$'

[[field]]
path = "example"
type = "identifier"
presence = "required"
means = "Invented for a fixture."

[scalar.identifier]
pattern_key = "identifier_pattern"
max_length_key = "identifier_max_length"
`

// at writes one schema directory and returns the root above it.
func at(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		path := filepath.Join(root, Dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func edit(t *testing.T, old, new string) string {
	t.Helper()
	if !strings.Contains(minimal, old) {
		t.Fatalf("the fixture no longer contains %q, so this edit tests nothing", old)
	}
	return strings.Replace(minimal, old, new, 1)
}

func TestSchemaLoadsTheFixture(t *testing.T) {
	set, err := Load(at(t, map[string]string{"record-1.toml": minimal}))
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Versions) != 1 || set.Versions[1] == nil {
		t.Fatalf("expected version 1 alone, got %v", set.SortedVersions())
	}
}

func TestSchemaLoadsThisRepositorysOwnSchema(t *testing.T) {
	set, err := Load("../..")
	if err != nil {
		t.Fatalf("the schema this repository ships does not load: %v", err)
	}
	if len(set.Versions) == 0 {
		t.Fatalf("no version loaded")
	}
}

// The property this package is here for. A key the validator would not have
// applied is refused rather than skipped, because a rule written in the
// authority and applied by nothing reads as enforced to everybody who opens
// the file.
func TestSchemaRefusesAKeyItDoesNotRead(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-1.toml": edit(t, `presence = "required"`, "presence = \"required\"\nunits_are_metric = true")}))
	if err == nil {
		t.Fatalf("a key nothing reads was accepted")
	}
	if !strings.Contains(err.Error(), "units_are_metric") {
		t.Fatalf("the refusal has to name the key: %v", err)
	}
}

// The near miss for that rule: every key the real schema carries is read, so
// the refusal above is about a key nobody declared rather than about the
// loader being narrow.
func TestSchemaReadsEveryKeyTheRealSchemaWrites(t *testing.T) {
	if _, err := Load("../.."); err != nil {
		t.Fatalf("a key of this repository's own schema is unread: %v", err)
	}
}

func TestSchemaRefusesAFileWhoseNameAndVersionDisagree(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-2.toml": minimal}))
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("a record naming a version has to reach exactly one file: %v", err)
	}
}

func TestSchemaRefusesATypeDefinedNowhere(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-1.toml": edit(t, `type = "identifier"`, `type = "quantity-name"`)}))
	if err == nil || !strings.Contains(err.Error(), "quantity-name") {
		t.Fatalf("a field whose type is defined nowhere is a field nothing checks: %v", err)
	}
}

func TestSchemaRefusesAPresenceItCannotApply(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-1.toml": edit(t, `presence = "required"`, `presence = "recommended"`)}))
	if err == nil || !strings.Contains(err.Error(), "recommended") {
		t.Fatalf("a presence rule nothing implements was accepted: %v", err)
	}
}

func TestSchemaRefusesAConditionalFieldCarryingNoCondition(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-1.toml": edit(t, `presence = "required"`, `presence = "conditional"`)}))
	if err == nil || !strings.Contains(err.Error(), "no condition") {
		t.Fatalf("a conditional field with no condition was accepted: %v", err)
	}
}

func TestSchemaRefusesTheSamePathDeclaredTwice(t *testing.T) {
	twice := minimal + "\n[[field]]\npath = \"example\"\ntype = \"identifier\"\npresence = \"optional\"\nmeans = \"Invented.\"\n"
	_, err := Load(at(t, map[string]string{"record-1.toml": twice}))
	if err == nil || !strings.Contains(err.Error(), "twice") {
		t.Fatalf("a path declared twice leaves the last entry read deciding it: %v", err)
	}
}

func TestSchemaRefusesAPatternThatDoesNotCompile(t *testing.T) {
	_, err := Load(at(t, map[string]string{"record-1.toml": edit(t, `identifier_pattern = '^[a-z]+$'`, `identifier_pattern = '^[a-z'`)}))
	if err == nil || !strings.Contains(err.Error(), "identifier_pattern") {
		t.Fatalf("a pattern that does not compile was accepted: %v", err)
	}
}

func TestSchemaRefusesADirectoryWithNoSchemaInIt(t *testing.T) {
	if _, err := Load(at(t, map[string]string{"README.md": "no schema here\n"})); err == nil {
		t.Fatalf("a validator with no authority to read against accepts everything")
	}
}

func TestSchemaRefusesAMissingDirectory(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatalf("a missing schema directory was read as an empty one")
	}
}

func TestSchemaRefusesAnAtLeastOneOfNamingNoMember(t *testing.T) {
	body := minimal + "\n[type.example-block]\nat_least_one_of = [\"nowhere\"]\n\n[[type.example-block.member]]\nname = \"somewhere\"\ntype = \"identifier\"\npresence = \"optional\"\nmeans = \"Invented.\"\n"
	_, err := Load(at(t, map[string]string{"record-1.toml": body}))
	if err == nil || !strings.Contains(err.Error(), "nowhere") {
		t.Fatalf("a rule naming a member that does not exist can never bite: %v", err)
	}
}
