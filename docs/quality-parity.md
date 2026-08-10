# Quality parity with the sso board's gate

The target is the gate on https://github.com/Flowfin/jellyfin-plugin-sso, which
is the most complete gate available to copy from. This document is the map: every
leg of that gate, and for this board either adopted as it stands, adapted, or not
applicable, with one line of reasoning for every deviation.

The reason the map exists is the dropped leg. A gate copied by eye keeps the legs
somebody remembered and loses the ones nobody thought about, and the loss leaves
no trace, because a check that was never copied looks exactly like a check that
was considered and rejected. Every leg below carries a verdict for that reason,
including the ones that end in nothing being built here.

Almost none of it is in place yet. The section near the end says which legs those
are, and it is the section to read first if the question is what this board's gate
can currently refuse.

## How the target gate was read

Two different things could be called that board's gate, and only one of them is a
measurement.

The list of workflow files carrying a `pull_request` trigger is what the tree
says. The set of check runs the protected branch actually requires is what the
platform enforces, and where the two disagree the second one is the gate. Both
were read, and the map below walks the second.

    gh api repos/Flowfin/jellyfin-plugin-sso/rules/branches/main \
      --jq '.[] | select(.type=="required_status_checks")
            | .parameters.required_status_checks[].context'
    build
    ABI floor build
    Package (JPRM) / Build package
    Package (JPRM) / Generate SBOM
    CodeQL
    Analyze (csharp)
    DCO sign-off
    Deterministic PR-hygiene checks
    Enforce greppable invariants
    Reject Trojan Source Unicode
    Audit workflows (zizmor)
    prettier
    dependency-review

Thirteen contexts. The workflow files behind them were read at
`e9cee021e95763e5240b44b8d7af16598df609ce`, and every quotation below is from
that commit rather than from a working copy. That commit was the tip of that
repository's default branch when this document was written, taken from the
remote rather than from a clone that might be behind it:

    gh api repos/Flowfin/jellyfin-plugin-sso/commits/main --jq '.sha'
    e9cee021e95763e5240b44b8d7af16598df609ce

The two lists differ in both directions and the differences carry information.

    for f in $(git ls-tree --name-only origin/main .github/workflows/ \
                 | sed 's|.github/workflows/||'); do
      if git show "origin/main:.github/workflows/$f" | grep -qE '^\s*pull_request:'
      then echo "$f"; fi
    done
    codeql.yml
    dco.yml
    dependency-review.yml
    dotnet.yml
    e2e-login.yml
    opengrep.yml
    pr-hygiene.yml
    prettier.yml
    unicode-guard.yml
    zizmor.yml

Ten files, thirteen contexts, and one file in the list that is required by
nothing. `dotnet.yml` alone supplies three of the required contexts from three
jobs, and `e2e-login.yml` reports on a pull request without being required, which
is the shape the last row of the map is about.

One context is not one leg either. `build` is a single required check run that
carries a locked restore, a build with warnings as errors, the test suite and a
coverage bar, in that order, in one job. The map splits it, because this board
will land those four at different times under different issues, and a row that
tracked the check run rather than the leg would go green here while three of the
four were still missing.

## The map

`Verdict` is one of adopted, adapted, or not applicable. `Lands here` names the
issue on this board that builds this board's version, or says the leg is already
in place with the pull request that landed it. `Reasoning` is empty where the leg
is adopted as it stands and carries one line wherever this board deviates.

| Leg | Where it is on the sso board | Verdict | Lands here | Reasoning for the deviation |
| --- | --- | --- | --- | --- |
| Locked dependency restore, where a resolution differing from the committed pins fails rather than resolving quietly | `dotnet.yml`, in the `build` job: `dotnet restore --locked-mode` | Adopted | #14 for the lockfile and the restore, #53 for the pinning rule across dependencies and actions | |
| Build with warnings as errors | `dotnet.yml`, in the `build` job: `dotnet build --no-restore --warnaserror` | Adapted, and in place | In place, landed by #16 as `.github/workflows/build-and-test.yml`, running the single command #14 created | The Go compiler refuses rather than warns, so there is no flag with this spelling. What stands in for it is `go vet`, inside the command's format-and-lint leg, and #16 added nothing further: a stricter analyser reachable only from a workflow would refuse on a pull request what no local run refuses |
| The test suite | `dotnet.yml`, in the `build` job: `dotnet test --no-build --verbosity normal` | Adopted, and in place | In place, the harness landed by #15 and the check that runs it by #16 | |
| A coverage bar on the surface that decides refusals, failing closed on an unreadable or empty report | `dotnet.yml`, the `Enforce the security-surface coverage bar` step, which runs `scripts/check-coverage.py` over a Cobertura report | Adapted, and in place | In place, landed by #50 as the `coverage-floor` leg of the single command | There the surface is the code that authorises a login; here it is the validator's refusal paths and the estimators, because the equivalent of letting the wrong person in is admitting a wrong number or computing one. The estimators are not on the surface because no package here produces a number yet, which `coverage-floor.toml` says of itself. The bar is a leg of the one command rather than a workflow of its own, so a contributor meets it locally and `Build and test` carries it on a pull request |
| A formatter check that reports rather than applies | `prettier.yml`, job id `prettier` with no `name:` | Adopted, with the repair command in the same code path | In place, landed by #17 as `.github/workflows/format-and-lint.yml` | The check and the fix are one function here rather than two tools that agree until they do not, so `go run . fmt` writes exactly what the check demands. The prose half is trailing whitespace and a final newline rather than a prose formatter, because a Markdown formatter would add a runtime this tree does not carry, and the line-ending question is settled in `.gitattributes` rather than in a formatter setting |
| Static analysis of the source, reporting into the code scanning tab | `codeql.yml`, contexts `CodeQL` and `Analyze (csharp)` | Adapted, and in place | In place, landed by #18 as `.github/workflows/codeql.yml`, context `Analyze (go)` | Two deviations, both forced by what this tool reads. The threat model is local rather than the default remote, because this tool opens no socket and under the remote model no source reaches any sink here, so the analysis would have been green by construction. And the analysis is limited to the source: this analyser also reads the workflow files and zizmor already gates that subject under its own name |
| Dependency review of newly introduced and upgraded dependencies | `dependency-review.yml`, job id `dependency-review` with no `name:` | Adopted, and already in place | In place, `.github/workflows/dependency-review.yml` | |
| The Trojan Source unicode guard | `unicode-guard.yml`, `Reject Trojan Source Unicode` | Adopted, and already in place | In place, `.github/workflows/unicode-guard.yml` | |
| The workflow audit | `zizmor.yml`, `Audit workflows (zizmor)` | Adopted, and already in place | In place, `.github/workflows/zizmor.yml` | |
| The sign off check | `dco.yml`, `DCO sign-off` | Adopted, and already in place | In place, `.github/workflows/dco.yml`, with the `DCO` file it certifies against | |
| The pull request hygiene check | `pr-hygiene.yml`, `Deterministic PR-hygiene checks` | Adopted, and already in place | In place, landed by #72 | |
| A greppable invariants lint, run as token patterns rather than through a language parser, where a finding exits non-zero | `opengrep.yml`, `Enforce greppable invariants`, running `opengrep scan --config tools/opengrep/rules.yml --error .` | Adopted, with different invariants and a different reader | In place, landed by #65 as `.github/workflows/no-network-imports.yml` | The invariants are this board's, not that board's: nothing here authorises a login, and the property worth refusing is what the corpus and the validator may reach. The first one is also not decided by token patterns. A grep for `net/http` reads a comment, a string and a vendored copy alike, and it cannot see a package that reaches a socket three imports away without naming it; the leg reads the import graph the compiler would build instead. A later invariant that is genuinely a token pattern can still arrive as a second leg |
| An application binary interface floor build, so the artifact stays loadable inside the oldest supported host | `dotnet.yml`, `ABI floor build` | Not applicable, with an equivalent | #61 | There is no host application to stay loadable inside; the equivalent breakage here is a schema change that makes an existing corpus unreadable, so the parity leg is a format compatibility check rather than a build |
| A package build of the shipped artifact | `dotnet.yml`, `Package (JPRM)`, calling `build.yml`, context `Package (JPRM) / Build package` | Adapted | #56 | A release here is a data artifact as well as a binary, so the package step has two outputs rather than one |
| An inventory of what shipped, generated at release time | `build.yml`, `Generate SBOM`, context `Package (JPRM) / Generate SBOM` | Adapted | #53 for the inventory, #56 for publishing it with the release | The inventory has to cover the corpus artifact as well as the binary's dependencies, since a corpus release with no statement of which series and which revisions it holds is an opaque blob |
| An end to end harness against a live external service | `e2e-login.yml`, which carries a `pull_request` trigger with a `paths:` filter and is required by nothing | Not applicable in the gate, with an equivalent outside it | In place, landed by #15 as `test/online/`, with #11 holding the record that fixes where it goes | The equivalent is the network integration harness `docs/decisions/0010-headless-tests.md` names as `test/online/`, which stays outside the gate for the same reason that one is not required there: a leg that depends on somebody else's service reds on their outage |
| A translation catalog guard | Nowhere. See the section below | Not applicable | Nothing lands it | This board is English only, and the leg is also not in the target gate to be adopted from |

## The leg that is not there

The list of legs to walk includes a translation catalog guard, and the walk did
not find one. Nothing in that repository's workflow set mentions a translation
catalog, a localization file or a message catalog:

    git grep -ilE 'translation|localization|\.po\b|catalog' origin/main \
      -- .github/workflows/
    (no output, exit 1)

The verdict is unchanged, since this board is English only either way. What
changes is the reason, and the reason is the part a later reader acts on. A leg
recorded as not applicable reads as a leg somebody looked at and declined. This
one could not have been adopted, because there is nothing at the other end to
adopt, and a future round of parity work would otherwise go looking for it.

## Legs that board runs outside its gate

Two of that board's workflows are deliberately not gate legs, and both have an
issue on this milestone. They are recorded here rather than left out, because an
issue with no row is the same invisible drop the map exists against, and because
in both cases the sso board's reason for keeping them out of the gate is a
decision this board inherits rather than one it gets to retake quietly.

Mutation testing. `stryker-mutation.yml` has no `pull_request` trigger, so it
cannot become a required check by accident, and its break threshold is zero so a
low score never exits non-zero. The score is a test-quality signal read by a
person, and an infrastructure failure that produces no report at all still shows
red. #51 carries this board's version, over the validator and the estimators. The
default it inherits is that the score is reported and not enforced.

Fuzzing. `fuzz.yml` runs on a schedule and on manual dispatch, and not on a pull
request. #52 carries this board's version, over the record parser, and it is in
place: `.github/workflows/fuzz.yml` here has the same trigger shape for the same
inherited reason, that a fuzz run's length is not a thing a merge should wait on.
The targets are in `internal/validate/fuzz_test.go`.

Two things about this board's version are not inherited and are worth reading
before it is changed. The targets assert properties rather than waiting for a
panic, because the defects this validator can ship are quieter than a crash: two
runs over one input agreeing, every refusal naming a catalogue site and saying
what was found and what was expected, and acceptance implying that the file
decoded and named a schema version this tree carries. And one of the three
targets splices fuzzed values into a legal record instead of mutating a whole
file, because a whole-file mutator spends its budget on bytes that do not decode
and leaves the parsing underneath the decoder out of reach. That was measured
rather than supposed, and the measurement is in the pull request that landed it.

When it last ran is not a thing this document can say, and is a thing the runs
page can:

    gh run list --repo iderex/messbuch --workflow fuzz.yml --limit 10

Neither is a deviation from the gate, because neither is in the gate. They are
recorded so that a reader comparing this milestone's issue list against this map
does not find two issues the map never mentions and conclude the map is stale.

## Which legs are not yet in place

Plainly, and this is the important section. Of the seventeen legs in the map,
nine are in place on this board and one needs nothing built at all. The other
seven name work this board still owes, and five of the seven were waiting on a
source tree to build against, which now exists.

Seven of the eight in place are checks reporting on a pull request. The eighth
is the network integration harness, which is in place by never being a check at
all, and is listed separately below for that reason.

In place today:

    git ls-files .github/workflows/
    .github/workflows/build-and-test.yml
    .github/workflows/codeql.yml
    .github/workflows/dco.yml
    .github/workflows/dependency-review.yml
    .github/workflows/format-and-lint.yml
    .github/workflows/fuzz.yml
    .github/workflows/no-network-imports.yml
    .github/workflows/pr-hygiene.yml
    .github/workflows/scorecard.yml
    .github/workflows/unicode-guard.yml
    .github/workflows/validate-corpus.yml
    .github/workflows/zizmor.yml

Run on the branch that lands the fuzz targets. Before it, that list is the same
without `fuzz.yml`, and before the branch that landed the structural half of the
corpus validator it was also without `validate-corpus.yml`.

Two of those files report on no pull request at all and this heading used to
claim they all did. `scorecard.yml` never did, and `fuzz.yml` is deliberately
outside the gate in the section further down. The rest report on every one.

That is the build and test check, the static analysis, the sign off check,
dependency review, the formatter check, the pull request hygiene check, the
unicode guard, the workflow audit and the invariants leg. `scorecard.yml` is not a gate leg and
`docs/required-checks.md` says why it must never be required.

One file in that list corresponds to no row in the map above, and the mismatch
is the honest state rather than an omission. `validate-corpus.yml` reports under
`Validate the corpus` and was landed by #24. The board being copied from has no
counterpart to it: nothing there reads a corpus of transcribed numbers, so there
was no leg to adopt, adapt or decline, and putting a row in the map for it would
claim a comparison that was never made. Counting the workflow files and counting
the legs of the map therefore give different answers, and they are answers to
different questions.

Not in place as a check on a pull request, and no longer waiting on the source
tree or the dependency manifest they read, because both now exist:

    git ls-files | grep -cE '\.go$|go\.mod|go\.sum'
    46

The format compatibility check is what is left in that state. The coverage bar
came off this list with #50, which is a leg of the single command rather than a
workflow of its own, so it adds no check name and no row to
`docs/required-checks.md`. The number above was 11 when this section was written and nothing moved
it while three landings added source, which is the shape of drift a pasted count
has: it stays plausible. Re-run it rather than reading it. The first to land was the toolchain pin and the
single command, #14, and the rest wrap or read what it creates. The locked
restore, the build with warnings as errors and the test suite came off this
list with #16, which runs the whole of that command on every pull request, so
each of the three is now a leg the board refuses a change over rather than one
a contributor runs. The static analysis came off it with #18, which is not a
leg of that command: it needs a database built by an analyser the tree does not
carry, so it lives in a workflow of its own and a contributor does not run it
locally.

How each of the three that came off the list is spelled here. The locked
restore is the command's `modules` leg, which runs `go mod verify` against
`go.sum` with `-mod=readonly` set, so a build cannot quietly rewrite the pins to
make itself work. The build is its `build` leg and the suite is its `test` leg.
Warnings as errors has no direct spelling in this toolchain, since the Go
compiler refuses rather than warns; what stands in for it is the
`format-and-lint` leg's `go vet`, and #16 decided to add nothing beyond it.

Every leg of that command now runs on a pull request, which is what changed
here. `no-network-imports.yml` and `format-and-lint.yml` each invoke the single
leg they report under, so a red line under either names the property that
failed. `build-and-test.yml` invokes the command whole, so a leg with no
workflow of its own is still refused. The price is that a formatting defect
reds two checks and an import defect reds two checks, which is deliberate and
is argued in `docs/required-checks.md`. The command prints which of its own
legs it did not run, and that output is the authority for the covered set
rather than this document:

    go run . ci

Not in place, and blocked on something other than the source tree: the package
build and the release inventory, which need a release to be published at all, and
which sit on the release milestone behind #56. With the six above them, that is
the whole of what is owed.

The network integration harness is no longer among them. `test/online/` exists,
it carries the build constraint that keeps it out of every ordinary run, it
reports the number of tests it ran even when that number is zero, and the gate
names it as not run on every run rather than passing over it in silence. It is
the one piece of parity work here that is deliberately never a gate leg, so
"in place" for it means exactly that and nothing about a pull request.

Two of the legs already in place reported on a subject that did not exist, and
that is not the same as reporting clean. `dependency-review` was one of them,
which `docs/required-checks.md` already writes down, and `go.mod` and `go.sum`
now give it a manifest to review. The pull request hygiene check reads the pull
request rather than the tree, so it is the leg here that has been fully doing
its job throughout.

Five checks are required on this board's protected branch, and neither the
formatter check nor the invariants leg is among them:

    gh api repos/iderex/messbuch/rules/branches/main \
      --jq '.[] | select(.type=="required_status_checks")
            | [.parameters.required_status_checks[].context] | join("; ")'
    DCO sign-off; dependency-review; Deterministic PR-hygiene checks; Reject Trojan Source Unicode; Audit workflows (zizmor)

That is the five that were already reporting when the set was configured. A
check outside that list reports and does not refuse a merge, so `Format and
lint` and `No network outside the net package` are advice on the branch until
the maintainer adds their strings. Which strings belong in the set is what `docs/required-checks.md` is
written to be read against, and configuring it is the first item on the release
checklist.

## What checks this

Nothing.

No check in this repository reads this file. Nothing compares the map against
either board's workflow set, nothing refuses a leg that loses its issue, and
nothing notices when a required context is added or removed on the board being
copied from. Every row is a claim about another repository at one commit, and the
commit moves.

Re-derive rather than trust, with the three commands at the top of this document.
Where the output disagrees with the table, the output is right and the table is a
defect. Adding a row is the cheap half. Noticing that the gate on the other side
has grown a leg this map does not mention is the expensive half, and no machine
here does it.
