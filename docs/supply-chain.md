# The supply chain: what is pinned, what a release will list, and what the audit found

Pinning, the release inventory and the audit are three separate states of the
supply chain, and they are easy to confuse for each other. All three are below,
so each can be read on its own without being inferred from the other two.

## Pinning

Every action every workflow uses is pinned to a forty-character commit, with the
tag it corresponded to in a trailing comment rather than in the reference:

    grep -rh 'uses:' .github/workflows/ | grep -v 'security-extended' | wc -l
    30

    grep -rh 'uses:' .github/workflows/ | grep -v 'security-extended' \
      | grep -c '@[0-9a-f]\{40\} #'
    30

Thirty uses, thirty pins. The one line the first filter removes is
`- uses: security-extended`, which names a CodeQL query pack inside the CodeQL
configuration and is not an action reference.

The number here read twenty-three until the fuzz workflow landed, and only five
of the seven it moved by are that workflow's. Running the first command against
the tree without it prints twenty-five, so two uses had already arrived under
other landings while this file went on saying twenty-three. That is what a
pasted count does rather than a fault of any one change, and it is the reason
the commands sit above the number instead of underneath it.

One action appears in two workflows and both carry the same commit.
`actions/upload-artifact` is pinned at `v7.0.1` in `.github/workflows/scorecard.yml`
and at the same commit in `.github/workflows/fuzz.yml`, which is deliberate: two
revisions of one action in one tree is the same supply-chain surface twice and a
state nobody can read off a single line.

Two of the counted uses are `actions/cache/restore` and `actions/cache/save` at
one commit, which is one action used twice rather than two actions. Splitting it
is what lets a run keep the corpus it built when its fuzzer found something, and
the argument is at the steps in `.github/workflows/fuzz.yml`.

The Go side has one dependency and it is pinned by `go.mod` with its hash in
`go.sum`:

    go list -m all
    github.com/iderex/messbuch
    github.com/BurntSushi/toml v1.6.0

The restore is locked. `runGo` in `internal/gate/golegs.go` sets
`GOFLAGS=-mod=readonly` and `GOWORK=off` on every `go` invocation the gate makes,
so a requirement missing from `go.mod`, or a sum missing from `go.sum`, is a
refusal rather than an edit that makes the build work. The modules leg then
recomputes the hash of every module in the build list and compares it against
`go.sum`:

    go run . ci modules
    ok       modules            1 module(s) in the build list verify against go.sum

What that does not cover is worth stating beside it. `-mod=readonly` is the Go
toolchain's default, so the setting makes the behaviour explicit rather than
adding it, and a run outside the gate that sets `GOFLAGS` differently is not
refused by anything here. The dependency-review check reads a pull request's
manifest changes against known vulnerabilities and stays; it says nothing about
a dependency that was already in the tree.

## The inventory of what a release contains

Not done. Nothing has been released, and no route generates an inventory:

    gh release list --repo iderex/messbuch --limit 5
    (no output)

It covers the corpus artifact as well as the binary's dependencies, since the
data is part of what ships and a corpus release with no statement of which
series and which revisions it holds cannot be read by the downstream packager
who most needs to. #53 owes the generation and #56 owes the publishing, and item
10 of `docs/release-checklist.md` is where a release is stopped for the absence.

## The self-audit, and what each of its findings is

The audit is `ossf/scorecard-action`, which runs on the default branch and
uploads its results to the code-scanning tab. Nothing triaged them until this
file, and an untriaged report is decoration.

Read from one run rather than from the tab's current state, because the tab
moves:

    gh run list --repo iderex/messbuch --workflow scorecard.yml --limit 1 \
      --json databaseId,headSha,conclusion,createdAt \
      --jq '.[] | "\(.databaseId) \(.conclusion) \(.createdAt) \(.headSha)"'
    31307467027 success 2026-08-09T10:04:57Z a942e253fc174e7e7fe83edebd4488ded48e161b

Six findings were open at that run:

    gh api "repos/iderex/messbuch/code-scanning/alerts?per_page=100&state=open" \
      --jq '.[] | "\(.tool.name)\t\(.rule.id)"' | sort
    Scorecard	BranchProtectionID
    Scorecard	CIIBestPracticesID
    Scorecard	CodeReviewID
    Scorecard	DependencyUpdateToolID
    Scorecard	FuzzingID
    Scorecard	MaintainedID

Each carries one of three states. Re-run the two commands above rather than
trusting this table: a finding that has gone, or a new one, makes the table
wrong and the commands right.

### Code-Review, score 0. Accepted, with a date on the acceptance

`Found 0/11 approved changesets`. It is correct. No pull request on this board
has carried an approving review, the ruleset requires no approvers, and the
person writing a change is the person merging it.

Accepted until the first release, and not beyond it. Entry 4 of #13 decided that
contributions from outside are not taken until then, so there is nobody to
approve a changeset, and requiring an approver would mean requiring a review
that cannot happen. What is not claimed is that the changes were reviewed: they
were not, and where a pull request body says so it says so plainly rather than
leaving the reader to infer it from a green check.

This is the finding with the largest gap between its score and its cost. A
corpus whose argument is that unchecked numbers propagate is a project that
cannot afford an unchecked mainline for long. The acceptance expires with
`docs/decisions/0019-outside-contributions-before-first-release.md`.

### Branch-Protection, score 3. Held elsewhere, and not by this file

Five warnings on `main`: stale review dismissal disabled, no required
approvers, no required codeowners review, last-push approval disabled, and
up-to-date branches disabled.

Four of the five are the review question above under other names, and the fifth,
up-to-date branches, is a merge-order setting. What is held here is only the
required-status-check half, and it is already tracked: item 2 of
`docs/release-checklist.md` says the set is configured and does not yet cover
everything that reports, and `docs/required-checks.md` holds the strings and
names the two that must never be required.

Changing the ruleset is the maintainer's, so this file records the finding and
takes no action on it.

### Fuzzing, score 0. Filed as its own issue, and the issue exists

`no fuzzer integrations found`. Correct, and it is #52, which owes fuzzing of
the record parser on the quality-parity milestone. Nothing further is owed here.

### Maintained, score 0. Accepted, and it resolves without action

`project was created within the last 90 days`. It is a fact about the
repository's age rather than a defect, and no change can make it false today. It
stops being reported when the repository is older than the window.

### CII-Best-Practices, score 0. Accepted

`no effort to earn an OpenSSF best practices badge detected`. The badge is a
self-assessment questionnaire. Filling it in before the first release would mean
attesting to practices this project has not performed yet, and several of its
questions are the same review and release questions that are open above.
Revisit at the first release, when the answers would be about something that has
happened.

### Dependency-Update-Tool, score 0. Accepted, with the condition that reverses it

`no dependency update tool configurations found`. Correct. There is no
Dependabot or equivalent configuration in the tree.

The tree has one Go dependency and thirty action pins, all of which are
already visible in a diff when they move, and the dependency-review check reads
manifest changes on every pull request. An update tool's value grows with the
size of the tree and this one is small.

It is accepted rather than fixed for a second reason that is not about value.
An update tool opens pull requests, and entry 4 of #13 decided this project does
not take contributions from outside before the first release; turning on a bot
that opens them is close enough to that decision to be worth the maintainer
taking rather than a builder assuming. Revisit when the dependency count grows
beyond what a person tracks, or at the first release, whichever comes first.

## What checks any of this

Nothing.

No check reads this file, nothing compares the pin count above against the
workflows, and nothing refuses a finding that has been open since the audit
started reporting it. `PROSE, NOT ENFORCEMENT`, issue #53. The pinning half is
the exception in one narrow direction only: the modules leg refuses a build list
that does not verify against `go.sum`, and that is a check, but it says nothing
about whether an action reference is a tag or a commit.
