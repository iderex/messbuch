# Proposed required checks for the protected branch

This document is a proposal and nothing else. Changing the branch protection is
the maintainer's action on the repository settings, not a change that arrives by
pull request, so this file cannot make any of it true.

## The protection today

The default branch carries one active ruleset. It requires a pull request and it
refuses deletion and non-fast-forward pushes. It requires no status check at
all, and it requires no approving review.

    gh api repos/iderex/messbuch/rulesets --jq '.[] | "\(.id) \(.name) \(.enforcement)"'
    20525370 gate active

    gh api repos/iderex/messbuch/rulesets/20525370 --jq '[.rules[].type]'
    ["deletion","non_fast_forward","pull_request"]

    gh api repos/iderex/messbuch/rulesets/20525370 \
      --jq '[.rules[] | select(.type=="required_status_checks") | .parameters.required_status_checks[].context]'
    []

So every check in this repository can be red and the merge button still works.
Until that changes, the gate is advice.

Run those commands rather than trusting the output pasted above. The output is
what they printed on 2026-08-07 and it is a fact about that day. Restating a
live setting in a document is how a document starts lying, and this file is the
one place least able to afford it.

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

These four already report on pull requests in this repository.

| Matching string | Workflow file | What it refuses | Reports on every pull request |
| --- | --- | --- | --- |
| `DCO sign-off` | `.github/workflows/dco.yml` | A pull request any of whose non-merge commits lacks a `Signed-off-by` trailer matching that commit's author. Fails closed if the commit range cannot be walked. | Yes. Triggered on `pull_request` with the default types and no branch or path filter. |
| `dependency-review` | `.github/workflows/dependency-review.yml` | A newly introduced or upgraded dependency carrying a known advisory at severity low or above. | Yes. Triggered on `pull_request` with no filter. The job declares no `name:`, so the check run takes the job id. |
| `Reject Trojan Source Unicode` | `.github/workflows/unicode-guard.yml` | Bidirectional and invisible Unicode control characters in tracked text, the Trojan Source class. Fails closed on a scanner error rather than reading it as a clean tree. | Yes. Triggered on `pull_request` against `"**"`, no path filter. |
| `Audit workflows (zizmor)` | `.github/workflows/zizmor.yml` | Any actionable zizmor finding at low severity or above in the workflow files, and a workflow that fails to parse. | Yes. Triggered on `pull_request` against `"**"`. The job's only conditional is on the SARIF upload step, not on the gating step and not on the job. |

`dependency-review` has a property worth writing down before it is discovered.
There is no dependency manifest in this repository today, so it reports without
having anything to review. That is a check reporting green on an empty subject,
not a check that examined a dependency set and found it clean, and it stays that
way until the toolchain issue lands a lockfile.

## Must not be required

| Matching string | Workflow file | Why not |
| --- | --- | --- |
| `Scorecard analysis` | `.github/workflows/scorecard.yml` | No `pull_request` trigger, and the job is guarded to the default branch. It can never report on a pull request, so requiring it blocks every merge permanently. |
| `zizmor` | `.github/workflows/zizmor.yml` | This is the code-scanning check run from the SARIF upload, not the job that fails on findings. It is absent on fork and Dependabot pull requests. Require `Audit workflows (zizmor)` instead. See the measurement below. |

## Proposed additions, as each lands

These do not exist yet. Each row's string is fixed by the issue that builds the
check, and a check whose string is not fixed yet is listed with its string
blank rather than guessed at, because a guessed string is the rename failure
above with extra steps.

| Matching string | Built by issue | What it will refuse |
| --- | --- | --- |
| `Build and test` | #16 | A pull request that does not build, or whose tests fail, or that produces a warning, run through the same single command a contributor runs locally. |
| `Validate the corpus` | #24 | A record file that is not a well formed record: a parse failure, an unknown field, a missing required field, a wrong type, a value outside a closed set, a duplicate identifier, or a file in the wrong place for what it claims to be. |
| not yet fixed | #17 | Formatting and lint. |
| not yet fixed | #18 | Static analysis of the source, reporting into the code scanning tab. The issue fixes the job name as `Analyze` extended with the language where the analysis distinguishes languages, so the exact string is not knowable until the language decision lands. |
| not yet fixed | #21 | Pull request hygiene for this board. |
| `No network outside the net package` | #10 | A package other than `net` reaching a network-capable API through the import graph. The string is fixed by `docs/decisions/0009-offline-by-default.md`. |

When one of these lands, the pull request that lands it fills in its row here.
A check that exists and is not in this table is a check the maintainer will not
know to require.

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

The four strings in the table above are confirmed by that output. It also shows
two things the table did not predict, and both are the sort of thing that turns
a required check into a deadlock.

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
