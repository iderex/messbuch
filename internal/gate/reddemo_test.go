package gate

import "testing"

// TestRedDemoTheTestLegRefusesAFailingTest exists on a branch that is never
// merged. It fails on purpose, so that `Build and test` has been seen red
// under its real name on a real pull request rather than assumed to work.
//
// The gate stops at the first failing leg, so the run this produces also shows
// that nothing after `test` reported a result nobody looked at.
func TestRedDemoTheTestLegRefusesAFailingTest(t *testing.T) {
	t.Error("deliberate failure: this branch exists to be refused")
}
