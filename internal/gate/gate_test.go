package gate

import (
	"errors"
	"strings"
	"testing"
)

func okLeg(id string) Leg {
	return Leg{
		ID:      id,
		Subject: "a subject",
		Run:     func(string) (string, error) { return id + " examined something", nil },
	}
}

// A failure ends the run. Without this, a later leg reports a verdict on a
// tree an earlier leg already refused, and the last line of the output is the
// one a reader remembers.
func TestRunStopsAtTheFirstFailure(t *testing.T) {
	reached := false
	legs := []Leg{
		okLeg("first"),
		{
			ID:      "second",
			Subject: "the one that fails",
			Run:     func(string) (string, error) { return "", errors.New("refused for the stated reason") },
		},
		{
			ID:      "third",
			Subject: "must not run",
			Run:     func(string) (string, error) { reached = true; return "", nil },
		},
	}

	var out chunk
	err := Run(&out, legs, t.TempDir())
	if err == nil {
		t.Fatal("a failing leg returned no error")
	}
	if reached {
		t.Error("the leg after the failure ran")
	}

	var failed *ErrLegFailed
	if !errors.As(err, &failed) {
		t.Fatalf("error is %T, want *ErrLegFailed", err)
	}
	if failed.ID != "second" {
		t.Errorf("the error names leg %q, want %q", failed.ID, "second")
	}
	if !strings.Contains(err.Error(), "refused for the stated reason") {
		t.Errorf("the leg's own reason did not reach the caller: %v", err)
	}
	if !strings.Contains(out.String(), "stopped at second") {
		t.Errorf("the run did not say where it stopped:\n%s", out.String())
	}
}

// A leg that is not built says so. This is the property that keeps a partial
// run from reading as a complete one, and it is the reason Legs carries
// entries with no implementation at all.
func TestRunSaysWhichLegsWereNotRun(t *testing.T) {
	legs := []Leg{
		okLeg("built"),
		{
			ID:      "unbuilt",
			Subject: "something nobody has written yet",
			Owed:    "not built. Issue #999 owes it.",
		},
	}

	var out chunk
	if err := Run(&out, legs, t.TempDir()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	got := out.String()
	for _, want := range []string{"not run", "unbuilt", "Issue #999 owes it.", "1 leg(s) ran and 1 were not run"} {
		if !strings.Contains(got, want) {
			t.Errorf("the output does not carry %q:\n%s", want, got)
		}
	}
}

// What a leg examined reaches the reader. A gate that prints "ok" and nothing
// else cannot be told from one that examined an empty set.
func TestRunPrintsWhatEachLegExamined(t *testing.T) {
	var out chunk
	if err := Run(&out, []Leg{okLeg("first")}, t.TempDir()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	if !strings.Contains(out.String(), "first examined something") {
		t.Errorf("the output does not say what the leg examined:\n%s", out.String())
	}
}

// An empty leg list is a refusal rather than a clean run, because "examined
// nothing" and "examined everything and found nothing" print the same way
// otherwise.
func TestRunRefusesAnEmptyGate(t *testing.T) {
	var out chunk
	if err := Run(&out, nil, t.TempDir()); err == nil {
		t.Fatal("a gate with no legs passed")
	}
}

// A leg that is neither built nor owed is refused before anything runs, so the
// refusal cannot be read as a verdict on the repository.
func TestRunRefusesALegThatIsNeitherBuiltNorOwed(t *testing.T) {
	ran := false
	legs := []Leg{
		{ID: "first", Subject: "runs", Run: func(string) (string, error) { ran = true; return "", nil }},
		{ID: "hollow", Subject: "declared and forgotten"},
	}
	var out chunk
	if err := Run(&out, legs, t.TempDir()); err == nil {
		t.Fatal("a leg with neither Run nor Owed passed")
	}
	if ran {
		t.Error("a leg ran before the declaration was refused")
	}
}

// The declared gate itself has to satisfy that rule, which is the case the
// test above cannot reach because it builds its own legs.
func TestTheDeclaredGateIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, leg := range Legs() {
		if leg.ID == "" {
			t.Error("a leg carries no id")
		}
		if seen[leg.ID] {
			t.Errorf("two legs share the id %q, so a failure names an ambiguous leg", leg.ID)
		}
		seen[leg.ID] = true
		if leg.Subject == "" {
			t.Errorf("leg %s says nothing about what it looks at", leg.ID)
		}
		if (leg.Run == nil) == (leg.Owed == "") {
			t.Errorf("leg %s is neither exactly built nor exactly owed", leg.ID)
		}
	}
}

// A leg's limits reach the reader beside its result. A boundary printed only
// in a decision record is a boundary the person reading the green line does
// not have.
func TestRunPrintsALegsLimitsBesideItsResult(t *testing.T) {
	leg := okLeg("bounded")
	leg.Limits = "it reads the import graph.\nIt does not read what a subprocess does."

	var out chunk
	if err := Run(&out, []Leg{leg}, t.TempDir()); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	for _, want := range []string{"it reads the import graph.", "It does not read what a subprocess does."} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the output does not carry %q:\n%s", want, out.String())
		}
	}
}

// The same, when the leg fails. A refusal is the moment its limits matter
// most, because that is when somebody argues about what it covers.
func TestRunPrintsALegsLimitsWhenItFails(t *testing.T) {
	legs := []Leg{{
		ID:      "bounded",
		Subject: "the one that fails",
		Limits:  "what this leg cannot see",
		Run:     func(string) (string, error) { return "", errors.New("refused") },
	}}

	var out chunk
	if err := Run(&out, legs, t.TempDir()); err == nil {
		t.Fatal("the failing leg returned no error")
	}
	if !strings.Contains(out.String(), "what this leg cannot see") {
		t.Errorf("the limits are absent from a failing run:\n%s", out.String())
	}
}

// Only narrows the gate to one leg, which is how a workflow reporting under a
// single fixed check name says something about that property alone.
func TestOnlyNarrowsTheGateToOneLeg(t *testing.T) {
	legs, err := Only([]Leg{okLeg("first"), okLeg("second")}, "second")
	if err != nil {
		t.Fatalf("Only returned %v", err)
	}
	if len(legs) != 1 || legs[0].ID != "second" {
		t.Fatalf("Only returned %d leg(s), first is %q", len(legs), legs[0].ID)
	}
}

// An id nobody declares is an error rather than an empty run. This is the
// guard on renaming a leg: without it, a workflow pinned to the old id runs
// nothing, prints nothing and reports green, and a required check that has
// stopped checking is worse than none.
func TestOnlyRefusesAnIdTheGateDoesNotDeclare(t *testing.T) {
	_, err := Only([]Leg{okLeg("first")}, "no-such-leg")
	if err == nil {
		t.Fatal("an unknown leg id was accepted")
	}
	if !strings.Contains(err.Error(), "first") {
		t.Errorf("the error does not say what the gate does declare: %v", err)
	}
}

// The id the workflow pins has to be one the gate declares. The test above
// proves Only refuses an unknown id; this one proves the id in
// .github/workflows/no-network-imports.yml is not that case.
func TestTheGateDeclaresTheLegTheNetworkWorkflowRuns(t *testing.T) {
	if _, err := Only(Legs(), "no-network-imports"); err != nil {
		t.Fatalf("the workflow's leg id is not in the gate: %v", err)
	}
}

// chunk is a writer that keeps what was written, so a test can read the run's
// own output rather than a summary of it.
type chunk struct{ b strings.Builder }

func (c *chunk) Write(p []byte) (int, error) { return c.b.Write(p) }
func (c *chunk) String() string              { return c.b.String() }
