# 0009 Offline by default, and where the network surface lives

## Question

The tool reads data an operator may not be allowed to disclose. What is the
default network behaviour, where may a socket be opened at all, and what makes
that boundary true by construction rather than by promise?

## Options considered

No network code anywhere in the project.

Network capability anywhere, with an opt-out flag or configuration setting that
turns it off.

Network capability anywhere, off by default, on by flag.

Network capability confined to one named component that the operator invokes
deliberately, with every other package refused the capability by a check on the
import graph.

## What each option costs

No network anywhere. The strongest possible statement and the easiest to
verify. It costs two features that are genuinely useful and that somebody will
otherwise build badly outside the project: resolving a source identifier
against a bibliographic service while transcribing, and submitting a
contribution upstream. Refusing both means every contributor does the first by
hand in a browser and the second by a route the project does not control.

Opt-out flag. Cheapest to build and it matches what most tools do. It costs the
guarantee entirely. A default that reaches the network is a default that
reaches the network on the first run, before the operator has read anything,
and the data at risk is at risk on that first run. It also makes the promise
untestable, because the property becomes a runtime configuration state rather
than a fact about the program.

Off by default, on by flag, capability anywhere. Better, and still a runtime
property. Every package can open a socket, so the guarantee rests on every code
path checking the flag, forever, including code written by somebody who did not
know the flag existed. The failure is silent and the operator has no way to
detect it short of watching the host's traffic.

One named component, enforced on the import graph. The property becomes
structural: a package that cannot import anything network-capable cannot open a
socket regardless of what its code says or what flags are set, and the check
that decides this reads the source rather than the running program. It costs a
check that has to be built and kept accurate, it costs an architectural
constraint that will occasionally be inconvenient, and it is only as good as
the list of what counts as network-capable, which is a list that has to be
maintained against the language's standard library and the dependency set.

## Choice

Offline by default, enforced structurally.

The default. Running an analysis opens no socket. There is no telemetry, no
crash reporting, no update check, no fetching of a schema from a server, and no
background resolution of an identifier. None of these is a setting. The
analysis path has no code that could do them.

The one component. Exactly one package in the tree may import a
network-capable API. Its name is `net`, it lives at the path the language
decision fixes for a package of that kind, and everything that reaches the
network in this project lives inside it. Two behaviours are expected to live
there: resolving a source identifier against a bibliographic service, and
submitting a contribution upstream. Both are useful, both are invoked as their
own subcommand by an operator who typed it, and neither runs as a side effect
of anything else.

Before it sends anything, that component prints what it will send and where it
will send it, and it does so on every invocation rather than on the first. A
disclosure the operator has to remember from a previous run is not a
disclosure.

The check. A check named `No network outside the net package` reads the import
graph of the tracked source, computes the set of packages that transitively
reach a network-capable API, and refuses when that set contains anything other
than `net` and its own dependencies. It reports on every pull request and it
fails closed: an import graph it cannot compute is a refusal, not a pass. The
list of what counts as network-capable is data in the check rather than a
condition in its code, so adding a name to it is a reviewable one-line change
rather than an edit to a matcher.

The check exists, and what it does and does not cover is stated where its
result is printed. It is the `no-network-imports` leg of the single local gate
command, in `internal/gate/netimports.go`, and it reports on every pull request
from `.github/workflows/no-network-imports.yml` under the exact string above.
#65 landed it. That paragraph read `PROSE, NOT ENFORCEMENT` while there was no
source tree to compute a graph over, and the change from an intention to a
property is the one thing about this record that moved.

It is not required by the branch protection. Five checks are required on the
default branch and this is not one of them:

    gh api repos/iderex/messbuch/rules/branches/main \
      --jq '.[] | select(.type=="required_status_checks")
            | [.parameters.required_status_checks[].context] | join("; ")'
    DCO sign-off; dependency-review; Deterministic PR-hygiene checks; Reject Trojan Source Unicode; Audit workflows (zizmor)

So a red run here reports and does not refuse a merge. Adding the string to that
set is the maintainer's configuration to make, and `docs/required-checks.md` is
where the argument for it is. Reporting and refusing are different words and
this record uses the one that is true.

The check's own limit, stated here because it will otherwise be discovered
later and read as a defect. An import-graph check refuses a package that can
reach a socket. It does not refuse a package that shells out to a program that
opens one, and it does not refuse a dependency that opens a socket from inside
code the graph shows as reached for another reason. The first is covered by a
separate refusal on subprocess execution outside the same component; the second
is not covered by anything here and is what the dependency pinning and the
release inventory on the quality milestone are for. Saying the check is
sufficient would be a stronger claim than it can carry.

## What this means for the corpus itself

The published measurements and their citations are not personal data in any
interesting sense. They are numbers from papers, and the authors' names in a
citation are already public in the form the corpus repeats them.

The contributors are a different matter. A contribution to this repository puts
a name and an address into a public version history permanently, and this
project cannot promise to rewrite that history later. That is a property of how
the work is done and it is a thing to be honest about in the documentation
rather than a thing somebody discovers after their first commit. Anyone who
needs to contribute under a different identity has to know before they start.

An operator pointing the tool at their own unpublished series is the case the
default exists for. That data may be commercially sensitive, may be under an
agreement forbidding disclosure, and may identify the people who took the
measurements, since a series carries who measured, when and where. None of it
leaves the host as a side effect of an analysis. It is written where the
operator says to write it and nowhere else.

The documentation half of this, meaning what the readme and the operator
documentation say in the operator's own words, is a separate issue on the
release milestone and is not discharged by this record.

## Reasons

The confined-component option is chosen because it is the only one of the four
that turns the guarantee into something a check can decide. The other three
make it a property of runtime state, and a property of runtime state cannot be
verified by a reader, cannot be verified by a reviewer, and can only be
verified by an operator who is watching their own network interface, which is
not a reasonable thing to ask.

Refusing the network entirely was close, and it lost on a practical judgement
rather than a principled one. The two features are things people will do
anyway, and doing them inside a component with a stated disclosure is safer
than doing them outside the project with none.

The cost that is paid knowingly is the check's maintenance. A list of
network-capable APIs goes stale, and a check that has gone stale reads as
enforcement while enforcing less than it claims. That is why the list is data
and why the check fails closed on a graph it cannot compute: both make staleness
noisier than silence.

There is a legal edge to this beyond the engineering one. An operator in a
regulated setting has to be able to state what the tool does with the data they
give it, and the only answer that survives an audit is one backed by something
other than the project's word. That is the same reason the check's limits are
written down here: an assurance with an unstated boundary is worse in an audit
than a narrower one with a stated boundary.

## Date

2026-08-07

## Reversal condition

Reverse the confinement if the implementation language turns out to have no way
to compute a reliable import graph over the tracked source, because at that
point the check cannot be built and keeping the architecture while admitting
the enforcement is impossible would be the runtime-flag option wearing this
record's clothes. The honest fallback in that case is the no-network-anywhere
option, not a promise.

Revisit the two permitted behaviours if either turns out to need to run during
an analysis rather than as its own invocation. That would be a request to make
the network path implicit, and it is refused by this record; the reversal
condition is written so that the request is argued here rather than
accommodated quietly.

Revisit the whole record if the project ever ships a hosted service, because
every sentence above is about a tool running on the operator's own machine and
none of it transfers.
