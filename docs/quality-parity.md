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
| Build with warnings as errors | `dotnet.yml`, in the `build` job: `dotnet build --no-restore --warnaserror` | Adopted | #16, through the single command #14 creates | |
| The test suite | `dotnet.yml`, in the `build` job: `dotnet test --no-build --verbosity normal` | Adopted | #15 for the harness, #16 for the check that runs it | |
| A coverage bar on the surface that decides refusals, failing closed on an unreadable or empty report | `dotnet.yml`, the `Enforce the security-surface coverage bar` step, which runs `scripts/check-coverage.py` over a Cobertura report | Adapted | #50 | There the surface is the code that authorises a login; here it is the validator's refusal paths and the estimators, because the equivalent of letting the wrong person in is admitting a wrong number or computing one |
| A formatter check that reports rather than applies | `prettier.yml`, job id `prettier` with no `name:` | Adopted | #17 | |
| Static analysis of the source, reporting into the code scanning tab | `codeql.yml`, contexts `CodeQL` and `Analyze (csharp)` | Adopted | #18 | |
| Dependency review of newly introduced and upgraded dependencies | `dependency-review.yml`, job id `dependency-review` with no `name:` | Adopted, and already in place | In place, `.github/workflows/dependency-review.yml` | |
| The Trojan Source unicode guard | `unicode-guard.yml`, `Reject Trojan Source Unicode` | Adopted, and already in place | In place, `.github/workflows/unicode-guard.yml` | |
| The workflow audit | `zizmor.yml`, `Audit workflows (zizmor)` | Adopted, and already in place | In place, `.github/workflows/zizmor.yml` | |
| The sign off check | `dco.yml`, `DCO sign-off` | Adopted, and already in place | In place, `.github/workflows/dco.yml`, with the `DCO` file it certifies against | |
| The pull request hygiene check | `pr-hygiene.yml`, `Deterministic PR-hygiene checks` | Adopted, and already in place | In place, landed by #72 | |
| A greppable invariants lint, run as token patterns rather than through a language parser, where a finding exits non-zero | `opengrep.yml`, `Enforce greppable invariants`, running `opengrep scan --config tools/opengrep/rules.yml --error .` | Adopted, with different invariants | #65 lands the first invariant, no network capable import outside the one permitted package | The invariants are this board's, not that board's: nothing here authorises a login, and the properties worth grepping for are about what the corpus and the validator may reach |
| An application binary interface floor build, so the artifact stays loadable inside the oldest supported host | `dotnet.yml`, `ABI floor build` | Not applicable, with an equivalent | #61 | There is no host application to stay loadable inside; the equivalent breakage here is a schema change that makes an existing corpus unreadable, so the parity leg is a format compatibility check rather than a build |
| A package build of the shipped artifact | `dotnet.yml`, `Package (JPRM)`, calling `build.yml`, context `Package (JPRM) / Build package` | Adapted | #56 | A release here is a data artifact as well as a binary, so the package step has two outputs rather than one |
| An inventory of what shipped, generated at release time | `build.yml`, `Generate SBOM`, context `Package (JPRM) / Generate SBOM` | Adapted | #53 for the inventory, #56 for publishing it with the release | The inventory has to cover the corpus artifact as well as the binary's dependencies, since a corpus release with no statement of which series and which revisions it holds is an opaque blob |
| An end to end harness against a live external service | `e2e-login.yml`, which carries a `pull_request` trigger with a `paths:` filter and is required by nothing | Not applicable in the gate, with an equivalent outside it | #11 fixes where it goes, #15 builds the harness | The equivalent is the network integration harness `docs/decisions/0010-headless-tests.md` names as `test/online/`, which stays outside the gate for the same reason that one is not required there: a leg that depends on somebody else's service reds on their outage |
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
request. #52 carries this board's version, over the record parser. The same
inheritance applies: a fuzz run's length is not a thing a merge should wait on.

Neither is a deviation from the gate, because neither is in the gate. They are
recorded so that a reader comparing this milestone's issue list against this map
does not find two issues the map never mentions and conclude the map is stale.

## Which legs are not yet in place

Plainly, and this is the important section. Of the seventeen legs in the map,
five are in place on this board and one needs nothing built at all. The other
eleven name work this board still owes, and eight of the eleven were waiting on
a source tree to build against, which now exists.

In place today, all reporting on every pull request:

    git ls-tree --name-only origin/main .github/workflows/
    .github/workflows/dco.yml
    .github/workflows/dependency-review.yml
    .github/workflows/pr-hygiene.yml
    .github/workflows/scorecard.yml
    .github/workflows/unicode-guard.yml
    .github/workflows/zizmor.yml

That is the sign off check, dependency review, the pull request hygiene check,
the unicode guard and the workflow audit. `scorecard.yml` is not a gate leg and
`docs/required-checks.md` says why it must never be required.

Not in place as a check on a pull request, and no longer waiting on the source
tree or the dependency manifest they read, because both now exist:

    git ls-files | grep -cE '\.go$|go\.mod|go\.sum'
    11

The locked restore, the build with warnings as errors, the test suite, the
coverage bar, the formatter check, the static analysis, the greppable invariants
lint and the format compatibility check are all in that state. The first of them
to land was the toolchain pin and the single command, #14, and the rest wrap or
read what it creates.

Two of those eight now exist inside that command rather than merely being
startable. The locked restore is its `modules` leg, which runs `go mod verify`
against `go.sum` with `-mod=readonly` set, so a build cannot quietly rewrite the
pins to make itself work. The build is its `build` leg. Warnings as errors is
not part of either and is owed by #16 and #17, which decide the spelling.

Nothing on a pull request runs that command yet, and that distinction is the
whole subject of this section. Until #16 lands, a leg of the single command is
something a contributor runs and not something this board refuses a change over.
The command prints which of its own legs it did not run, and that output is the
authority for the covered set rather than this document:

    go run . ci

Not in place, and blocked on something other than the source tree: the package
build and the release inventory, which need a release to be published at all, and
which sit on the release milestone behind #56. That is ten of the eleven. The
eleventh is the network integration harness, which is owed by #15 and is the one
piece of parity work here that is deliberately never a gate leg.

Two of the legs already in place reported on a subject that did not exist, and
that is not the same as reporting clean. `dependency-review` was one of them,
which `docs/required-checks.md` already writes down, and `go.mod` and `go.sum`
now give it a manifest to review. The pull request hygiene check reads the pull
request rather than the tree, so it is the leg here that has been fully doing
its job throughout.

Nothing on this board's protected branch is required yet:

    gh api repos/iderex/messbuch/rules/branches/main --jq '[.[].type] | join(" ")'
    deletion non_fast_forward pull_request

So every check named above is advice until the required set is configured, which
is the first item on the release checklist for exactly this reason, and which
`docs/required-checks.md` is written to be read against.

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
