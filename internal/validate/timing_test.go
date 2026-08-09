package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// The size the bound is measured at, and the bound.
//
// The size is absolute rather than a multiple of the corpus. The corpus holds
// no record today, so ten times its size is zero, and a test asserting a bound
// at ten times the current size is satisfied by validating nothing in a
// measurable fraction of no time at all. The multiplier stays behind the floor
// for the same reason at every size it would produce for years: reading many
// small files and any check quadratic in the number of records both cost
// nothing at a hundred records and are the whole cost at ten thousand.
//
// Ten thousand is a wall the corpus is being built towards rather than one it
// is near. Whether it is the right size for the seed corpus is that
// milestone's question; raising it here costs one constant.
const (
	timedRecords = 10000
	smallerRun   = 1000

	// The budget for validating timedRecords once the files have been read
	// once. See the disclosure in the test: this is the validator's own cost
	// and not the whole cost of the leg.
	//
	// It sits about five times over the measurement rather than just above it,
	// because this runs on whatever machine the gate runs on and a bound that
	// reds on a loaded runner teaches people to skip the suite. What it is
	// sized to catch is a regression of the kind this bound exists for, which
	// arrives as a factor and not as a percentage: a vocabulary file opened
	// per record, or a scan over supersessions that is quadratic in the number
	// of records.
	timedBudget = 10 * time.Second

	// How much worse than linear the walk may scale between smallerRun and
	// timedRecords. Ten times the records is ten times the work when the cost
	// is per record and a hundred times when it is quadratic, so this sits
	// between the two and needs no hardware to be read against.
	scalingAllowance = 25.0
)

// TestValidationStaysInsideItsTimeBound is the regression guard.
//
// It prints the hardware and the size beside the number, because a duration
// without those is not a bound, it is a number, and it gets quoted back on
// different hardware as though it were a property of the code.
//
// THE ASSERTION IS ON THE SECOND WALK AND THE FIRST IS ONLY REPORTED. On the
// host this was written on, the first walk over freshly written files costs
// around twenty times the second, and that cost is the operating system
// reading ten thousand new small files rather than anything this package does.
// A bound over the first walk would measure the host, move with whatever scans
// files on it, and drown the per-record regression it is here to catch. So the
// number below is the validator's own cost and NOT the whole cost of the leg
// on a fresh checkout, where every file is read for the first time. The first
// walk is logged so that half is visible rather than dropped.
func TestValidationStaysInsideItsTimeBound(t *testing.T) {
	root, first := generatedCorpus(t, timedRecords)
	took := walk(t, root, timedRecords)

	t.Logf("%d record(s) on %s/%s with %d logical CPU(s): %v on the first walk, %v on the second, against a budget of %v",
		timedRecords, runtime.GOOS, runtime.GOARCH, runtime.NumCPU(),
		first.Round(time.Millisecond), took.Round(time.Millisecond), timedBudget)

	if took > timedBudget {
		t.Errorf("validating %d record(s) took %v, over the budget of %v. That is the shape of regression this bound exists for: a second file opened per record, or a scan quadratic in the number of records",
			timedRecords, took.Round(time.Millisecond), timedBudget)
	}
}

// TestValidationScalesWithTheNumberOfRecords is the half of the bound that
// needs no hardware to be read.
//
// A budget in seconds says nothing about which machine it was met on. A ratio
// between two sizes on the same machine in the same run does: ten times the
// records is ten times the work while the cost is per record, and a hundred
// times when something is quadratic. A quadratic scan for superseded
// references is the case #29 names, and it would pass the budget above on a
// fast machine while making the gate unusable at the size the corpus is aiming
// for.
func TestValidationScalesWithTheNumberOfRecords(t *testing.T) {
	small, _ := generatedCorpus(t, smallerRun)
	smallTook := walk(t, small, smallerRun)

	large, _ := generatedCorpus(t, timedRecords)
	largeTook := walk(t, large, timedRecords)

	if smallTook <= 0 {
		t.Skipf("%d record(s) validated below this clock's resolution, so there is no ratio to take", smallerRun)
	}
	ratio := float64(largeTook) / float64(smallTook)
	t.Logf("%d record(s) in %v and %d in %v, a factor of %.1f for %.0f times the records, against an allowance of %.0f",
		smallerRun, smallTook.Round(time.Millisecond), timedRecords, largeTook.Round(time.Millisecond),
		ratio, float64(timedRecords)/float64(smallerRun), scalingAllowance)

	if ratio > scalingAllowance {
		t.Errorf("%.0f times the records cost %.1f times the time, over the allowance of %.0f. Something in the walk is worse than per-record",
			float64(timedRecords)/float64(smallerRun), ratio, scalingAllowance)
	}
}

// TestTheGeneratedRecordsAreOutsideTheCorpus is the other half of this issue,
// and it is the half that matters more.
//
// A fabricated measurement inside a corpus of measurements would end this
// project's credibility, and these fixtures are the only place synthetic
// records are written at all. The separation is not a directory name: the
// generator refuses a root inside this repository, so a later test that passed
// the repository root would fail rather than write.
func TestTheGeneratedRecordsAreOutsideTheCorpus(t *testing.T) {
	root, _ := generatedCorpus(t, 4)

	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	written, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if rel, err := filepath.Rel(repository, written); err == nil && !strings.HasPrefix(rel, "..") {
		t.Fatalf("the generator wrote into %s, which is inside this repository at %s", written, repository)
	}

	// The guard, proved by handing it the one root it has to refuse.
	if err := outsideTheRepository(repository); err == nil {
		t.Error("the generator accepted this repository as its root, so nothing stops a synthetic record being written where the corpus lives")
	}
	if err := outsideTheRepository(filepath.Join(repository, RecordDir)); err == nil {
		t.Error("the generator accepted the corpus directory itself as its root")
	}
	if err := outsideTheRepository(root); err != nil {
		t.Errorf("the generator refused a temporary directory outside this repository: %v", err)
	}
}

// walk times one validation of a generated corpus and fails on anything that
// would make the number mean something other than it says.
func walk(t *testing.T, root string, want int) time.Duration {
	t.Helper()
	set := load(t)
	start := time.Now()
	report, err := Corpus(root, set)
	took := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if report.Records != want {
		t.Fatalf("the walk read %d record(s) and the generator wrote %d, so this is not a measurement of the size it claims",
			report.Records, want)
	}
	if len(report.Refusals) > 0 {
		t.Fatalf("%d generated record(s) were refused, so this measured the refusal path rather than the accepting one:\n  %v",
			len(report.Refusals), report.Refusals[0])
	}
	return took
}

// generatedCorpus returns a root holding n synthetic records, and how long the
// first walk over them took.
//
// The corpora are built once per size and shared, because writing and then
// first-reading ten thousand small files is the expensive part of this file by
// a wide margin, and paying it twice would put minutes on the gate for a
// measurement that does not change. The first walk is timed here because there
// is exactly one of them per corpus and it is the number the disclosure above
// is about.
//
// Each record gets its own quantity every hundred files, so the walk crosses
// many directories rather than reading one directory with n files in it.
// Reading many small files is one of the cost centres this bound is measured
// against and a flat directory would not exercise it.
func generatedCorpus(t *testing.T, n int) (string, time.Duration) {
	t.Helper()
	corpora.Lock()
	defer corpora.Unlock()
	if made, ok := corpora.bySize[n]; ok {
		return made.root, made.firstWalk
	}

	root, err := os.MkdirTemp("", "messbuch-timing-")
	if err != nil {
		t.Fatal(err)
	}
	if err := outsideTheRepository(root); err != nil {
		t.Fatal(err)
	}

	const perDirectory = 100
	for i := range n {
		quantity := fmt.Sprintf("generated-quantity-%03d", i/perDirectory)
		year := 1900 + i%perDirectory
		dir := filepath.Join(root, RecordDir, quantity)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := strings.Replace(wellFormed, `quantity = "example-quantity"`, fmt.Sprintf("quantity = %q", quantity), 1)
		body = strings.Replace(body, `date = "1900"`, fmt.Sprintf("date = %q", fmt.Sprint(year)), 1)
		name := fmt.Sprintf("%04d-generated-%02d.toml", year, i%perDirectory)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	first := walk(t, root, n)
	corpora.bySize[n] = corpus{root: root, firstWalk: first}
	return root, first
}

// corpora holds the generated corpora for the run, so each size is written and
// first-read once.
var corpora = struct {
	sync.Mutex
	bySize map[int]corpus
}{bySize: map[int]corpus{}}

type corpus struct {
	root      string
	firstWalk time.Duration
}

// TestMain removes what the generator wrote. Nothing here is inside the
// repository, and a test that leaves ten thousand invented records on the disk
// after it has passed is a test somebody eventually finds and wonders about.
func TestMain(m *testing.M) {
	code := m.Run()
	for _, made := range corpora.bySize {
		os.RemoveAll(made.root)
	}
	os.Exit(code)
}

// outsideTheRepository refuses a root that a generated record could reach the
// corpus from.
func outsideTheRepository(root string) error {
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return err
	}
	target, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(repository, target)
	if err != nil {
		// Different volumes have no relative path between them, which is as
		// far outside as a path gets.
		return nil
	}
	if rel == "." || !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("%s is inside this repository at %s, and no generated record may be written there: a fabricated measurement in a corpus of measurements is the one defect that would end this project's credibility",
			target, repository)
	}
	return nil
}
