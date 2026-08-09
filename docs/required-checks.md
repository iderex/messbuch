# Proposed required checks for the protected branch

This document is a proposal and nothing else. Changing the branch protection is
the maintainer's action on the repository settings, not a change that arrives by
pull request, so this file cannot make any of it true.

## The protection today

The default branch carries one active ruleset. It requires a pull request, it
refuses deletion and non-fast-forward pushes, and it requires five status
checks. It requires no approving review.

    gh api repos/iderex/messbuch/rulesets --jq '.[] | "\(.id) \(.name) \(.enforcement)"'
    20525370 gate active

    gh api repos/iderex/messbuch/rulesets/20525370 --jq '[.rules[].type]'
    ["deletion","non_fast_forward","pull_request","required_status_checks"]

    gh api repos/iderex/messbuch/rulesets/20525370 \
      --jq '[.rules[] | select(.type=="required_status_checks") | .parameters.required_status_checks[].context]'
    ["DCO sign-off","dependency-review","Deterministic PR-hygiene checks","Reject Trojan Source Unicode","Audit workflows (zizmor)"]

So a check outside that list can be red and the merge button still works. That
is where `No network outside the net package`, `Format and lint`, `Build and
test`, `Analyze (go)` and `Validate the corpus` stand: each reports on every
pull request and refuses nothing on the branch until its string is added to the
set above. `Build and test` runs the whole local gate, so the widest of the five
is the one with the least force.

Run those commands rather than trusting the output pasted above. The output is
what they printed on 2026-08-09. The output this section carried before was
from 2026-08-07 and the third command printed an empty list then; when the
setting moved is not recorded anywhere and is not claimed here. The document
did not move with it, which is the failure this paragraph is about. Restating a live
setting in a document is how a document starts lying, and this file is the one
place least able to afford it.

## What the protection matches

The required status check is matched by the literal name of the check run. For
a workflow job that is the job's `name:`, and where a job declares no `name:`
it is the job id. It is not the workflow's `name:` and it is not the file name.

Two failure modes follow from that, and both are worse than having no
requirement.

A required check that never reports blocks every merge forever. The protection
waits for a check run with that exact name and there is nothing to distinguish
never-reported from still-running. This happens when a job is renamed, when a
check is required under the workflow name instead of the job name, and when the
workflow's trigger does not cover pull requests at all. The last one is not
hypothetical here: `Scorecard analysis` in `.github/workflows/scorecard.yml`
has no `pull_request` trigger and its job additionally carries
`if: github.event.repository.default_branch == github.ref_name`. Requiring it
would deadlock the branch, and it is listed below under what must not be
required precisely so that nobody adds it for looking useful.

A required check that reports on only some pull requests is worse than no
check, because the branch looks guarded and is not. A workflow whose trigger
filters by path, by branch or by base is in this class. Every check proposed
below is checked against this and the answer is stated per check rather than
assumed.

## Proposed set, from what exists today

These already report on pull requests in this repository. The count is not
stated, because a count in a sentence drifts against the table under it.

| Matching string | Workflow file | What it refuses | Reports on every pull request |
| --- | --- | --- | --- |
| `DCO sign-off` | `.github/workflows/dco.yml` | A pull request any of whose non-merge commits lacks a `Signed-off-by` trailer matching that commit's author. Fails closed if the commit range cannot be walked. | Yes. Triggered on `pull_request` with the default types and no branch or path filter. |
| `dependency-review` | `.github/workflows/dependency-review.yml` | A newly introduced or upgraded dependency carrying a known advisory at severity low or above. | Yes. Triggered on `pull_request` with no filter. The job declares no `name:`, so the check run takes the job id. |
| `Reject Trojan Source Unicode` | `.github/workflows/unicode-guard.yml` | Bidirectional and invisible Unicode control characters in tracked text, the Trojan Source class. Fails closed on a scanner error rather than reading it as a clean tree. | Yes. Triggered on `pull_request` against `"**"`, no path filter. |
| `Audit workflows (zizmor)` | `.github/workflows/zizmor.yml` | Any actionable zizmor finding at low severity or above in the workflow files, and a workflow that fails to parse. | Yes. Triggered on `pull_request` against `"**"`. The job's only conditional is on the SARIF upload step, not on the gating step and not on the job. |
| `Deterministic PR-hygiene checks` | `.github/workflows/pr-hygiene.yml` | A pull request body naming no issue, and a commit message carrying a byte outside printable ASCII plus line feed and horizontal tab. Fails closed if the commit range cannot be walked, and refuses before reading the pull request at all if its own fixtures do not behave as it claims. | Yes. Triggered on `pull_request` with types `opened`, `synchronize`, `reopened` and `edited`, and no branch or path filter. |
| `No network outside the net package` | `.github/workflows/no-network-imports.yml` | A package other than `internal/net` and its own dependencies transitively reaching a network-capable import path, test files included. Fails closed: a tree whose import graph cannot be computed, and a module whose package set comes back empty, are both refusals. | Yes. Triggered on `pull_request` with the default types and no branch, path or base filter. |
| `Format and lint` | `.github/workflows/format-and-lint.yml` | Go source that is not what gofmt writes, a tracked Markdown, YAML or TOML file carrying trailing whitespace or no final newline, and anything `go vet` objects to. Fails closed: a tree whose files cannot be walked, and a tree carrying no source and no prose at all, are both refusals. | Yes. Triggered on `pull_request` with the default types and no branch, path or base filter. |
| `Analyze (go)` | `.github/workflows/codeql.yml` | Any result the CodeQL `security-extended` suite reports over the Go source, under the local threat model, so command-line arguments, environment variables and file contents count as untrusted input. Fails closed: no SARIF, or one that will not parse, is a refusal rather than a clean tree. The refusal does not depend on the upload to the code-scanning tab, which is a separate step allowed to fail. | Yes. Triggered on `pull_request` with the default types and no branch, path or base filter. The gate step has no condition, so a fork or Dependabot pull request is refused on a finding even though its upload is skipped. |
| `Validate the corpus` | `.github/workflows/validate-corpus.yml` | A file under `record/` that is not a well formed record against the schema version it names: a parse failure, a path that is not a record's path, an unknown or misspelled field, a missing required field, a field another field's value requires or forbids, a wrong type, a value outside a closed set, a list below its minimum, a number below its minimum, the same entry twice in one list, and a block carrying none of the members it must carry one of. The list is printed by `go run . refusals` rather than kept current here. Fails closed: a schema set it cannot read and a record directory it cannot walk are both refusals rather than a clean corpus. | Yes. Triggered on `pull_request` with the default types and no branch, path or base filter. |
| `Build and test` | `.github/workflows/build-and-test.yml` | Whatever any leg of `go run . ci` refuses, because it invokes the whole command rather than one leg. Today that is the toolchain pin against the running release, the module set against `go.sum` in locked mode, the build, the tests, the headless rule, the corpus decode, the formatting and lint, and the offline boundary. The run prints its own covered set and the leg it stopped at. | Yes. Triggered on `pull_request` with the default types and no branch, path or base filter. |

`Deterministic PR-hygiene checks` has a property anyone requiring it needs to
know, and it is deliberate rather than a defect. It refuses only where the head
branch lives in this repository. A pull request from a fork gets every finding
as an annotation and the check still reports green, so requiring it does not
block an outside contribution and does not block on one. A run that did not
apply the refusing tier prints that it did not, so a green line on a fork pull
request is not the same statement as a green line on a branch from here. The
argument for the split is in the issue that built it, #21.

The diff-size finding in that check is a warning and can never red it. A bulk
transcription of a measurement series legitimately runs long.

`dependency-review` had a property worth writing down and no longer has it.
While there was no dependency manifest here it reported without having anything
to review, which is a check green on an empty subject rather than one that
examined a dependency set and found it clean. #14 landed `go.mod` and `go.sum`,
so it now has a subject:

    git ls-files go.mod go.sum
    go.mod
    go.sum

`Format and lint` checks and never applies. What repairs a refusal is one
command, `go run . fmt`, and it writes exactly what the check demands because
both go through the same function rather than through two ideas of what
formatted means. The refusal names that command. The check is not a prose
style, a line length or a spelling, its lint half is `go vet` and nothing else,
and `LICENSE` and `DCO` are read by neither half because their bytes are copies
of texts this project may not normalise.

`Build and test` is the only one of these that reports on the whole command
rather than on one leg of it, and that has a consequence worth knowing before
requiring it. A formatting defect reds it and reds `Format and lint`, and an
import defect reds it and reds `No network outside the net package`. The narrow
strings are the ones that say which property failed. This one says whether the
command a contributor runs is green, which is the property that stops the
remote half of the gate from checking a different set from the local half.

Its column above says what it refuses today rather than for all time, and the
run is the authority rather than the row. A leg added to the command widens
this check with no change to this file and no change to the workflow, which is
the point of one command and is also how this row goes stale:

    go run . ci

`No network outside the net package` carries its own limits and they are
printed beside its result on every run rather than only here. It refuses a
package that can reach a socket through the import graph. It does not refuse a
package that shells out to a program that opens one, it does not refuse a
dependency that opens a socket from inside code the graph shows as reached for
another reason, and it does not refuse a network-capable path that is absent
from the table it reads. `docs/decisions/0009-offline-by-default.md` is where
that boundary was argued and why an assurance with an unstated edge is worse in
an audit than a narrower one that says where it stops.

`Validate the corpus` answers the structural question and no other, and the
distinction is worth knowing before requiring it. It reads one file at a time
against `schema/record-<n>.toml` and never opens a second file to decide the
first, so a quantity naming no vocabulary entry, a group identifier resolving to
no registry file, a supersession pointing at a record that does not exist and a
conversion factor that is the wrong number all pass it. Those belong to the
meaning leg, which is #25 and is not built. Its own limits print beside its
result on every run, and what it can refuse is printed rather than restated
anywhere:

    go run . refusals

It is also the check whose subject can be empty. There is no record in this
repository yet, so the leg reports the count it read, and a count of zero is
what it prints rather than a green line that reads like a corpus somebody
checked.

## Must not be required

| Matching string | Workflow file | Why not |
| --- | --- | --- |
| `Scorecard analysis` | `.github/workflows/scorecard.yml` | No `pull_request` trigger, and the job is guarded to the default branch. It can never report on a pull request, so requiring it blocks every merge permanently. |
| `zizmor` | `.github/workflows/zizmor.yml` | This is the code-scanning check run from the SARIF upload, not the job that fails on findings. It is absent on fork and Dependabot pull requests. Require `Audit workflows (zizmor)` instead. See the measurement below. |
| `CodeQL` | `.github/workflows/codeql.yml` | The same shape one file over. This is the code-scanning check run from the SARIF upload, not the job that fails on findings, and the upload step is conditioned to same-repository non-Dependabot pull requests, so no such check run exists on a fork or Dependabot pull request. Require `Analyze (go)` instead. |

## Proposed additions, as each lands

These do not exist yet. Each row's string is fixed by the issue that builds the
check, and a check whose string is not fixed yet is listed with its string
blank rather than guessed at, because a guessed string is the rename failure
above with extra steps.

| Matching string | Built by issue | What it will refuse |
| --- | --- | --- |

The table is empty. `Validate the corpus` was its only row and it landed under
#24; its row is in the table above with the rest. An empty table here is not a
statement that nothing further is coming, only that no check with a fixed
string is waiting to be built today.

When one of these lands, the pull request that lands it fills in its row here.
A check that exists and is not in this table is a check the maintainer will not
know to require.

A row in this table naming a closed issue is the same defect as a check that is
missing from it, and it is harder to see. This one carried #10 until #65 was
opened. #10 asked the decision record to name the check, the record named it,
and #10 closed as completed, so the column read as an owner where there was
none and the check was owed by nobody. Read the column as a claim to be checked
rather than as a fact:

    gh issue list --repo iderex/messbuch --state open --limit 200 \
      --json number --jq '[.[].number] | sort | join(" ")'

## Reading the strings off a real run rather than off the YAML

A string derived from a workflow file is a claim about what the file says, not a
measurement of what the platform named the check run. The measurement is one
command against a pull request that has actually run them:

    gh api repos/iderex/messbuch/commits/<head-sha>/check-runs \
      --jq '[.check_runs[] | {name, app: .app.slug, conclusion}] | sort_by(.name)'

Where that output disagrees with the table above, the output is right and the
table is a defect.

Run on the head of the pull request that added this document,
`606ee3130247726b296556c6967de0777ea33f82`:

    [{"app":"github-actions","conclusion":"success","name":"Audit workflows (zizmor)"},
     {"app":"github-actions","conclusion":"success","name":"DCO sign-off"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"dependency-review"},
     {"app":"github-advanced-security","conclusion":"success","name":"zizmor"}]

That run predates `Deterministic PR-hygiene checks`, so it confirms the strings
that existed on that commit and says nothing about the one added since.
`Deterministic PR-hygiene checks` is confirmed by the same command run against
`c36a5e5c5ef3e280e9ef5b55cbb1436c94868ef1`, a commit on the pull request that
added it:

    [{"app":"github-actions","conclusion":"success","name":"Audit workflows (zizmor)"},
     {"app":"github-actions","conclusion":"success","name":"DCO sign-off"},
     {"app":"github-actions","conclusion":"success","name":"Deterministic PR-hygiene checks"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"dependency-review"},
     {"app":"github-advanced-security","conclusion":"success","name":"zizmor"}]

The commit named there is not the head that merged, because the output has to
exist before the sentence quoting it can be written. Naming the commit the
command was actually run against is the point; running it against an older
commit and reading the result as a statement about the table as it stands today
would be the defect this section exists to prevent.

`Build and test` is confirmed the same way, against
`bc5615d280e16fc766b377505363b186adbed176`, the head of the pull request that
added it:

    [{"app":"github-actions","conclusion":"success","name":"Audit workflows (zizmor)"},
     {"app":"github-actions","conclusion":"success","name":"Build and test"},
     {"app":"github-actions","conclusion":"success","name":"DCO sign-off"},
     {"app":"github-actions","conclusion":"success","name":"Deterministic PR-hygiene checks"},
     {"app":"github-actions","conclusion":"success","name":"Format and lint"},
     {"app":"github-actions","conclusion":"success","name":"No network outside the net package"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"dependency-review"},
     {"app":"github-advanced-security","conclusion":"success","name":"zizmor"}]

That output also confirms `Format and lint`, which landed between the two runs
above and had not been read off a run until here.

A green line is half of what a required string needs. The other half is that
the name goes red when the property fails, and for `Build and test` that was
watched on `d06e44c0d6a84488fb281d40968f83140680980e`, a branch carrying a test
that fails on purpose and which is never merged:

    gh api repos/iderex/messbuch/commits/d06e44c0d6a84488fb281d40968f83140680980e/check-runs \
      --jq '[.check_runs[] | select(.name=="Build and test") | {name, app: .app.slug, conclusion}]'
    [{"name":"Build and test","app":"github-actions","conclusion":"failure"}]

`Analyze (go)` is confirmed against `b54cd1bae81ba3c0f091384fb2c3727c170fce73`,
the head of the pull request that added it:

    [{"app":"github-actions","conclusion":"success","name":"Analyze (go)"},
     {"app":"github-actions","conclusion":"success","name":"Audit workflows (zizmor)"},
     {"app":"github-actions","conclusion":"success","name":"Build and test"},
     {"app":"github-advanced-security","conclusion":"success","name":"CodeQL"},
     {"app":"github-actions","conclusion":"success","name":"DCO sign-off"},
     {"app":"github-actions","conclusion":"success","name":"Deterministic PR-hygiene checks"},
     {"app":"github-actions","conclusion":"success","name":"Format and lint"},
     {"app":"github-actions","conclusion":"success","name":"No network outside the net package"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"dependency-review"},
     {"app":"github-advanced-security","conclusion":"success","name":"zizmor"}]

That output is also where the `CodeQL` row in the previous table comes from. It
is a second check run under a second name from a second app, produced by the
upload rather than by the verdict, which is `zizmor` and `Audit workflows
(zizmor)` again in a different file.

Red, on `653264ed4277c900f3b96d0c135c02b40186c19a`, a branch carrying a path
read from the command line and an allocation sized from it, which is never
merged:

    gh api repos/iderex/messbuch/commits/653264ed4277c900f3b96d0c135c02b40186c19a/check-runs       --jq '[.check_runs[] | select(.name|test("Analyze|CodeQL")) | {name, app: .app.slug, conclusion}] | sort_by(.name)'
    [{"name":"Analyze (go)","app":"github-actions","conclusion":"failure"},
     {"name":"CodeQL","app":"github-advanced-security","conclusion":"failure"}]

`Validate the corpus` is confirmed against
`c1635d9176c7d6a2f46b7c96791daff873e3f09f`, the head of the pull request that
added it:

    [{"app":"github-actions","conclusion":"success","name":"Analyze (go)"},
     {"app":"github-actions","conclusion":"success","name":"Audit workflows (zizmor)"},
     {"app":"github-actions","conclusion":"success","name":"Build and test"},
     {"app":"github-advanced-security","conclusion":"success","name":"CodeQL"},
     {"app":"github-actions","conclusion":"success","name":"DCO sign-off"},
     {"app":"github-actions","conclusion":"success","name":"Deterministic PR-hygiene checks"},
     {"app":"github-actions","conclusion":"success","name":"Format and lint"},
     {"app":"github-actions","conclusion":"success","name":"No network outside the net package"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Reject Trojan Source Unicode"},
     {"app":"github-actions","conclusion":"success","name":"Validate the corpus"},
     {"app":"github-actions","conclusion":"success","name":"dependency-review"},
     {"app":"github-advanced-security","conclusion":"success","name":"zizmor"}]

Red, on `f759a56fe7456d644311381a0f88af24621deae3`, a branch carrying a file
under `record/` that misspells a required field, which is never merged:

    gh api repos/iderex/messbuch/commits/f759a56fe7456d644311381a0f88af24621deae3/check-runs \
      --jq '[.check_runs[] | select(.name=="Validate the corpus") | {name, app: .app.slug, conclusion}]'
    [{"app":"github-actions","conclusion":"failure","name":"Validate the corpus"}]

The older output also shows two things the table did not predict, and both are
the sort of thing that turns a required check into a deadlock.

`Reject Trojan Source Unicode` appears twice, because
`.github/workflows/unicode-guard.yml` triggers on both `push` to `"**"` and
`pull_request` to `"**"`, and a pull request from a branch in this repository
produces one run of each. Two check runs share one name. Anyone requiring that
string should know that the protection is resolving a name that is not unique on
the commit, and should not read a single green line in the pull request page as
proof that both ran.

`zizmor` is a second, different check run, produced by the
`github-advanced-security` app rather than by Actions. It is the code-scanning
result of the SARIF upload, not the job that gates on findings. Requiring
`zizmor` instead of `Audit workflows (zizmor)` would look correct and would be
wrong twice over: it gates on an upload rather than on a verdict, and the upload
step in `.github/workflows/zizmor.yml` carries both a `continue-on-error: true`
and an `if:` restricting it to pushes to `main` and to same-repository
non-Dependabot pull requests. On a fork pull request or a Dependabot pull
request that step does not run, so no such check run is created, so a required
`zizmor` would block those merges permanently. The gating step, `Fail on
actionable findings`, has no condition and runs everywhere.

The name to require is `Audit workflows (zizmor)`. The name to leave alone is
`zizmor`.

## What this document does not do

Nothing here is enforcement. No check in this repository reads this file, no
route compares it against the live ruleset, and nothing refuses a pull request
that adds a workflow without adding a row. It drifts the moment a job is renamed
and the only thing that catches the drift is a person running the commands
above.
