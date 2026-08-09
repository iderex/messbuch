package gate

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures are base64 in source, which is this repository's rule for a
// fixture whose bytes are the point. A raw literal would be two things at
// once: the bytes the fixture carries, and a line of this file that the leg
// under test reads while scanning the repository's own test files. The second
// would refuse this file for carrying the very strings it exists to prove are
// refused.
var headlessFixtures = map[string]string{
	"network": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkibmV0L2h0dHAiCgkidGVzdGluZyIKKQoKZnVu" +
		"YyBUZXN0UmVhY2hlc1RoZU5ldHdvcmsodCAqdGVzdGluZy5UKSB7CglpZiBodHRwLk1ldGhv" +
		"ZEdldCA9PSAiIiB7CgkJdC5FcnJvcigidW51c2VkIikKCX0KfQo=",
	"display": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkib3MiCgkidGVzdGluZyIKKQoKZnVuYyBUZXN0" +
		"UmVhY2hlc0FEaXNwbGF5KHQgKnRlc3RpbmcuVCkgewoJaWYgb3MuR2V0ZW52KCJESVNQTEFZ" +
		"IikgPT0gIiIgewoJCXQuU2tpcCgibm8gZGlzcGxheSIpCgl9Cn0K",
	"elevation": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkib3MvZXhlYyIKCSJ0ZXN0aW5nIgopCgpmdW5j" +
		"IFRlc3RSZWFjaGVzRm9yUHJpdmlsZWdlcyh0ICp0ZXN0aW5nLlQpIHsKCWlmIGVyciA6PSBl" +
		"eGVjLkNvbW1hbmQoInN1ZG8iLCAiLW4iLCAidHJ1ZSIpLlJ1bigpOyBlcnIgIT0gbmlsIHsK" +
		"CQl0LkVycm9yKGVycikKCX0KfQo=",
	"outside": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkib3MiCgkidGVzdGluZyIKKQoKZnVuYyBUZXN0" +
		"UmVhY2hlc091dHNpZGVUaGVSZXBvc2l0b3J5KHQgKnRlc3RpbmcuVCkgewoJaWYgXywgZXJy" +
		"IDo9IG9zLlJlYWRGaWxlKCIvZXRjL2hvc3RzIik7IGVyciAhPSBuaWwgewoJCXQuRXJyb3Io" +
		"ZXJyKQoJfQp9Cg==",
	"home": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkib3MiCgkidGVzdGluZyIKKQoKZnVuYyBUZXN0" +
		"UmVhY2hlc1RoZUhvbWVEaXJlY3RvcnkodCAqdGVzdGluZy5UKSB7CglkaXIsIGVyciA6PSBv" +
		"cy5Vc2VySG9tZURpcigpCglpZiBlcnIgIT0gbmlsIHx8IGRpciA9PSAiIiB7CgkJdC5FcnJv" +
		"cigibm8gaG9tZSIpCgl9Cn0K",
	"clean": "cGFja2FnZSBzYW1wbGUKCmltcG9ydCAoCgkib3MiCgkicGF0aC9maWxlcGF0aCIKCSJ0ZXN0" +
		"aW5nIgopCgovLyBUaGlzIHRlc3QgbWVudGlvbnMgYSBkaXNwbGF5IGFuZCBzdWRvIGluIGEg" +
		"Y29tbWVudCwgYW5kIGNhcnJpZXMgYSBidWlsZAovLyBjb25zdHJhaW50IGFzIGEgc3RyaW5n" +
		"LCBub25lIG9mIHdoaWNoIGlzIHJlYWNoaW5nIGZvciBhbnl0aGluZy4KZnVuYyBUZXN0UmVh" +
		"Y2hlc05vdGhpbmcodCAqdGVzdGluZy5UKSB7CglkaXIgOj0gdC5UZW1wRGlyKCkKCXBhdGgg" +
		"Oj0gZmlsZXBhdGguSm9pbihkaXIsICJub3RlLm1kIikKCWlmIGVyciA6PSBvcy5Xcml0ZUZp" +
		"bGUocGF0aCwgW11ieXRlKCIvL2dvOmJ1aWxkIG9ubGluZQoiKSwgMG82NDQpOyBlcnIgIT0g" +
		"bmlsIHsKCQl0LkZhdGFsKGVycikKCX0KCWlmIF8sIGVyciA6PSBvcy5SZWFkRmlsZShmaWxl" +
		"cGF0aC5Kb2luKCJ0ZXN0ZGF0YSIsICJub3RlLm1kIikpOyBlcnIgPT0gbmlsIHsKCQl0Lkxv" +
		"ZygicmVsYXRpdmUgcGF0aHMgYXJlIG9yZGluYXJ5IikKCX0KfQo=",
}

// b64 decodes one fixture. A fixture that does not decode is a mistake in this
// file rather than a finding about anything, so it ends the test.
func b64(t *testing.T, name string) string {
	t.Helper()
	encoded, ok := headlessFixtures[name]
	if !ok {
		t.Fatalf("no fixture named %q", name)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return string(decoded)
}

// headlessTree writes one fixture into a fresh tree as a test file.
func headlessTree(t *testing.T, name string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "sample")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sample_test.go"), []byte(b64(t, name)), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// refuses asserts that the fixture is refused and that the refusal names the
// test and what it reached for. Naming the file alone would send somebody to a
// file with a dozen tests in it.
func refuses(t *testing.T, fixture string, wants ...string) {
	t.Helper()
	_, err := headlessLeg(headlessTree(t, fixture))
	if err == nil {
		t.Fatalf("the %s fixture passed", fixture)
	}
	for _, want := range wants {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q: %v", want, err)
		}
	}
}

// The four conditions, one fixture each. This is the part of #15 that makes
// the rule a rule rather than an explanation of one: a guard nobody has seen
// fail is a guard nobody knows works.
func TestHeadlessRefusesATestThatReachesTheNetwork(t *testing.T) {
	refuses(t, "network", "sample_test.go", "net/http", "the network")
}

func TestHeadlessRefusesATestThatReachesADisplay(t *testing.T) {
	// The name is taken from the list the leg reads rather than written out,
	// because a literal here is a line of a test file this leg scans.
	refuses(t, "display", "TestReachesADisplay", displayNames[0], "a display")
}

func TestHeadlessRefusesATestThatReachesForPrivileges(t *testing.T) {
	refuses(t, "elevation", "TestReachesForPrivileges", privilegedTools[0], "privileges")
}

func TestHeadlessRefusesATestThatReachesOutsideTheRepository(t *testing.T) {
	refuses(t, "outside", "TestReachesOutsideTheRepository", "outside this repository")
}

// The home directory is the same condition by a different route, and it is the
// one a path rule alone misses because no path is written down.
func TestHeadlessRefusesATestThatReadsTheHomeDirectory(t *testing.T) {
	refuses(t, "home", "TestReachesTheHomeDirectory", "os.UserHomeDir", "outside this repository")
}

// The near miss. A test that uses the directory the framework gives it, reads
// a relative path, and mentions a display and sudo in a comment, is ordinary
// work. A leg that refused this would be routed around within a week.
func TestHeadlessAcceptsATestThatReachesNothing(t *testing.T) {
	examined, err := headlessLeg(headlessTree(t, "clean"))
	if err != nil {
		t.Fatalf("an ordinary test was refused: %v", err)
	}
	if !strings.Contains(examined, OnlineHarness) {
		t.Errorf("the result does not name the harness it excluded: %q", examined)
	}
}

// The harness is excluded rather than merely unmentioned. Scanning it would
// refuse the one place in this project that is allowed to reach the outside
// world.
func TestHeadlessDoesNotScanTheOnlineHarness(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(OnlineHarness))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "online_test.go"), []byte(b64(t, "network")), 0o644); err != nil {
		t.Fatal(err)
	}
	ordinary := filepath.Join(root, "internal", "sample")
	if err := os.MkdirAll(ordinary, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ordinary, "sample_test.go"), []byte(b64(t, "clean")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := headlessLeg(root); err != nil {
		t.Fatalf("a test inside the online harness was refused: %v", err)
	}
}

// Fails closed. A tree with no test file is a leg that examined nothing, which
// is not the same statement as a suite that obeys the rule.
func TestHeadlessRefusesATreeWithNoTestFile(t *testing.T) {
	if _, err := headlessLeg(t.TempDir()); err == nil {
		t.Fatal("a tree with no test file passed")
	}
}

// The lists are data, and data with no entries refuses nothing.
func TestHeadlessListsAreNotEmpty(t *testing.T) {
	if len(displayNames) == 0 || len(privilegedTools) == 0 || len(homeReaders) == 0 {
		t.Fatal("one of the lists this leg reads is empty")
	}
}

// The tree this repository actually is.
func TestTheTreeObeysTheHeadlessRule(t *testing.T) {
	examined, err := headlessLeg("../..")
	if err != nil {
		t.Fatalf("this repository has a test that reaches for something the gate does not have: %v", err)
	}
	if !strings.Contains(examined, OnlineHarness) {
		t.Errorf("the result does not name the harness it excluded: %q", examined)
	}
}

// The gate says on every run which suite it did not cover, naming the harness,
// its constraint and the command that runs it. Without this, a green run is a
// statement about an unstated subset.
func TestTheGateNamesTheHarnessItDidNotRun(t *testing.T) {
	var out chunk
	leg := Leg{
		ID:      "test",
		Subject: "every package's tests pass",
		Limits:  limitsOfTheTestLeg,
		Run:     func(string) (string, error) { return "some packages tested", nil },
	}
	if err := Run(&out, []Leg{leg}, t.TempDir()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	for _, want := range []string{OnlineHarness, OnlineTag, OnlineInvocation, "was not run"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the run does not carry %q:\n%s", want, out.String())
		}
	}
}

// And the declared gate is what carries it, which the test above cannot reach
// because it builds its own leg.
func TestTheDeclaredGateDisclosesTheHarness(t *testing.T) {
	for _, leg := range Legs() {
		if leg.ID == "test" {
			if !strings.Contains(leg.Limits, OnlineInvocation) {
				t.Errorf("the test leg does not name the harness invocation: %q", leg.Limits)
			}
			return
		}
	}
	t.Fatal("the gate declares no test leg")
}
