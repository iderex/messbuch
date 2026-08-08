# 0010 Headless and unelevated tests, and where the rest goes

## Question

Two kinds of work on this board do not obviously fit inside a test suite that
can run anywhere: drawing a plot, and resolving a source identifier against a
bibliographic service. What is the rule the gate's tests run under, where does
each of those two go instead, and what stops the second one from being read as
having passed when it did not run at all?

## Options considered

State the rule and rely on review to hold it.

State the rule and put everything that breaks it behind a runtime skip, so the
test exists in the suite and skips when the environment cannot carry it.

State the rule and put everything that breaks it in a separate harness outside
the gate, with its own invocation and its own reporting.

Have no rule and let each test take what it needs, on the argument that a
display or a socket is available on a developer machine and in a hosted runner
anyway.

## What each option costs

Review only. Nothing to build. It costs the property on the first busy day.
The rule is invisible at the moment it is broken, because a test that quietly
needs a display passes on the machine of the person who wrote it and on the
runner that already has one, and the failure surfaces later on somebody else's
machine as a test that hangs rather than as a test that says what it wanted.

Runtime skip. Cheap, familiar, and the test stays next to the code it covers.
It costs the thing this rule exists to protect. A skipped test and a passing
test are the same colour in every summary anybody reads, so a suite that
skipped its whole network half looks exactly like one that ran it. The
pressure is also in the wrong direction: once skipping is the normal way to
handle an environment that cannot carry a test, adding a test that needs the
network costs nothing, and the set of quietly-skipped tests grows without
anybody deciding it should.

Separate harness. The gate's suite has one environment and one meaning, and a
test needing the outside world cannot enter it by accident, because entering
it means moving a file. It costs a second invocation somebody has to remember,
it costs the harness its own reporting, and it costs the honest admission that
the harness will be run less often than the gate. That last cost is the one
worth naming: a test outside the gate is a test that runs when somebody
chooses, and choosing is a thing people stop doing.

No rule. Nothing to build and nothing to explain. The gate becomes something
that needs a particular machine, and a gate that needs a particular machine is
one contributors run with a shrug and skip when it is inconvenient. This
project also asks an operator in a regulated setting to believe that an
analysis opens no socket, and a test suite that opens sockets wherever it
likes is not a place that claim can be checked from.

## Choice

Every test that runs in the gate runs with no display attached, no elevated
privileges, no network and no external service. That is a birth requirement of
this project and not something retrofitted later.

The four, said precisely, because each is broken in its own way.

No display. No test opens a window, requires a graphical session, or asserts
on a screenshot or on pixels.

No elevated privileges. No test requires administrator or root, installs a
certificate into a trust store, registers a service or a scheduled task, edits
a firewall rule, or does anything else whose failure mode on a contributor's
own machine is a consent prompt. A test that needs elevation is not run under
elevation; it is not in the gate. Where such a test is skipped, the skip is
disclosed where the work is argued, and a workaround that obtains the
privilege by another route is not a repair.

No network. No test resolves a name, opens a socket, or reads a URL, including
to a service running on the same host.

No external service. No test needs a database, a container runtime, a message
broker or any other process the suite did not start itself, and none needs
credentials.

### The two classes that do not fit, and where each goes

Rendering. Funnel plots and forest plots are output, and the ordinary way to
test a plot is to look at it. That is refused here. The plotting code emits a
vector document, and the tests assert on the document: that the expected
points are present at the expected coordinates, that the axes cover the
expected range with the expected ticks, that an interval is drawn from the
value the record holds to the value the record holds, and that a marked study
is marked. No pixels, no rasterisation, no screenshot and no display anywhere
in the path.

This is not a compromise made to satisfy the rule. Asserting on a document is
the stronger test: a pixel comparison fails when a font renders differently
and passes when a point is plotted at the wrong value, which is the wrong way
round for a project whose subject is whether a plotted number is the number
that was published. The rule and the better test agree here, and it is worth
saying so, because the two do not always agree and this record would be less
trustworthy if it implied they did.

Reaching the outside world. Resolving a source identifier against a
bibliographic service genuinely needs the network, and a test of that path is
worth having, because the path exists and is the one place in this project
where something leaves the host.

That test goes into a separate harness. The harness is `test/online/`, it is
guarded by the build constraint `//go:build online` so that no file in it is
compiled into an ordinary build or an ordinary test run, and it is invoked
deliberately:

    go test -tags online ./test/online/...

The name says what it needs. It is not `integration`, not `e2e` and not
`slow`, because none of those names tells a reader that running it sends
traffic off the machine, and that is the only property of it that matters
here.

Two rules hold inside the harness.

It must not run against real contributor data. The harness resolves
identifiers that are committed in the harness itself as fixtures, and it is
never pointed at a corpus an operator supplied or at anything under `record/`.
A network test that also reads real data is a route by which an operator's
unpublished series leaves their machine, and this project's whole offline
argument in `docs/decisions/0009-offline-by-default.md` is about not having
such a route.

Silence from a harness that did not run must not read like a pass. This is the
rule the harness exists to make possible and it is the one most easily lost.
The gate does not run the harness, and the gate says so on every run: a line
naming the harness, stating that it was not run, stating that it needs the
outside world, and printing the command above. A reader of a green gate run
therefore knows exactly what was not covered, and a run that covered less than
everything cannot be mistaken for one that covered everything and found
nothing.

### What the gate's configuration has to show

The requirement is that the gate's own configuration shows the harness
excluded rather than merely not mentioned. Not mentioning it would be the same
file a gate that had never heard of the harness would have, and the two states
have to be distinguishable by reading the configuration.

Excluded means the exclusion is written and printed rather than achieved by
the build constraint alone. The constraint keeps the harness out of the build.
It does not tell a reader that a harness exists, and a gate that is silent
about it is the runtime-skip option again with a longer path.

THE GATE'S CONFIGURATION DOES NOT EXIST TODAY AND THIS RECORD DOES NOT CREATE
IT. There is no build entry point, no test harness and no build or test
workflow in this repository, so no configuration here excludes anything. The
single command that is the local gate is owed by #14, the harness itself and
the headless rule's enforcement by #15, and the build and test check by #16.
Until those land, the paragraphs above describe what those issues have to
produce and not a property of anything that runs. This paragraph is the whole
disclosure and nothing later in this record softens it.

## Reasons

The separate harness beats the runtime skip on the one axis that decides this,
which is what a reader of a green run is entitled to conclude. Under a skip, a
green run means the tests that ran passed, and how many ran is a number nobody
reads. Under the harness, a green gate run means the whole gate ran, and the
one thing it did not cover names itself in the output. That is the difference
between a gate whose result is a fact and a gate whose result is a fact about
an unstated subset.

The rule is a birth requirement rather than a target because the direction of
travel is one way. Retrofitting it means going back through a suite and
finding which tests quietly acquired a dependency on their environment, and
that search is unbounded: a test that needs a display and a test that does not
look identical until it is run somewhere without one. Starting from the rule
costs one decision. Arriving at it later costs an audit.

Elevation is named separately from the other three because its failure mode is
not a red test. A test that wants a socket fails when the socket is refused. A
test that wants elevation raises a prompt on the machine of whoever ran it,
which takes the attention of a person who was doing something else, and on a
machine with a maintainer sitting at it that is a real interruption rather
than a build annoyance. That is why the rule is that such a test does not run,
rather than that it runs and is expected to fail.

The harness's name is fixed in this record for the same reason the offline
check's name is fixed in `docs/decisions/0009-offline-by-default.md`: the name
is what other things match on, and a name chosen later is a name that gets
changed later, quietly, leaving whatever referred to it pointing at nothing.

Nothing in this repository refuses a test that breaks any of this. `PROSE, NOT
ENFORCEMENT`. The repository has no test suite, so there is no test here to
refuse. #15 is the issue that owes the enforcement, and its condition is that
the headless rule is held by a check rather than by this document. What a
check can and cannot see is worth stating now so it is not overclaimed later:
a check can refuse a test file that imports a network-capable or
display-capable package, and it cannot refuse a test that shells out to a
program that opens a socket. That is the same boundary the import-graph check
carries, for the same reason, and it is stated in
`docs/decisions/0009-offline-by-default.md`.

## Date

2026-08-08

## Reversal condition

Reverse the separate-harness choice if the harness is observed not to be run
at all over a period in which the code it covers changed. At that point it is
not a harness, it is deleted tests with extra steps, and the honest response
is either to admit the network path is untested or to find a way to test it
that the gate can carry, not to keep a directory nobody invokes.

Reverse the assert-on-the-document rule for plots if a defect is found that a
document assertion structurally cannot see and a rasterised comparison would.
The candidate is a defect in the renderer rather than in the plot: correct
coordinates in the document and wrong output on the page. If one appears, the
answer is a rendering test that runs outside the gate under the same rules as
the online harness, not a display inside it.

Revisit the no-external-service rule if the project acquires a component whose
behaviour is genuinely a conversation with another process and cannot be
established by a test that starts that process itself. Nothing on this board
looks like that today, and the rule should not be relaxed in advance of one.
