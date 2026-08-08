# 0002 The implementation language and toolchain

## Question

This project ships three things that run: a validator that refuses a malformed
record on every pull request, an engine that computes random-effects estimates,
heterogeneity statistics and publication-bias diagnostics, and a command an
operator who is not a programmer installs and runs on a machine with no
network. What language are they written in, which toolchain version is pinned,
and what numerical dependencies does that leave the project carrying?

The three pull in different directions. The validator wants to be fast and to
have an exact refusal message. The engine wants a numerical library ecosystem,
and the ecosystem for the methods this project imports is not in a compiled
language. The operator wants one file and no interpreter. Anything chosen here
gives one of the three something less than it wants, and the record says which.

## Options considered

One compiled language for both the validator and the engine, in Go.

One compiled language for both, in Rust.

Python, with the scientific stack the meta-analysis literature already uses.

A split: the validator in a compiled language and the engine in Python.

## What each option costs

Go. The operator install is one static file with no runtime to install, the
build is reproducible with flags the toolchain already carries, and the
import graph over the tracked source is computable by the toolchain itself,
which is what the offline guarantee in
`docs/decisions/0009-offline-by-default.md` rests on. Compilation is fast
enough that the gate stays a thing people run rather than avoid. The costs are
real and there are three. The meta-analysis reference implementations are not
in Go, so every estimator is written from a published definition rather than
inherited, and each one has to be pinned against published numbers before it is
believed. There is no plotting library the project would want to depend on, so
the funnel and forest plots emit a vector document directly, which is more work
and, as it happens, is the same shape the headless-test rule wants anyway.
Third, arbitrary-precision arithmetic exists in the standard library but the
numerics of this project will be `float64`, so the reproducibility of a cited
number is a property the tests have to establish per platform rather than one
the language hands over.

Rust. The same operator story and a stronger one on the numerics: the type
system removes a class of defect the validator is exposed to, since it parses
files that arrive from people nobody has met. The costs are compile times that
are a tax on every gate run, a much smaller pool of contributors for a project
whose contributors are mostly going to be scientists transcribing papers rather
than systems programmers, and the same missing estimator ecosystem as Go, so
the largest cost of the compiled route is paid either way with no offsetting
gain. Its import-graph story for the offline check is also less direct: the
capability boundary would be expressed against crate dependencies and module
boundaries rather than against a package graph the toolchain prints.

Python. scipy and statsmodels exist, the methods this project imports are
implemented there and have been checked by many people, and a large body of
worked examples exists that a reviewer can hold an estimator against. The costs
land on the two things this project promised. The operator now needs an
interpreter and a dependency tree on a machine that may have neither and no
network to fetch them, which is either an install story the quickstart cannot
honestly claim is one command or a bundling exercise that reintroduces the
compiled-artifact problem with less control. And the reproducibility of a cited
number becomes a property of a resolved dependency graph rather than of a
pinned toolchain, so a number published from this tool is reproducible only by
somebody who can reconstruct that graph, which is the weaker of the two claims
by a wide margin.

The split. The validator gets speed and the engine gets the ecosystem, and
neither compromises. It pays both sets of costs at once: two toolchains in the
gate, two dependency-pinning stories, two formatting and lint configurations,
two static-analysis jobs, and a contributor who has to install both to run the
gate locally. It also puts the corpus schema in two places, because a validator
and an engine that read the same records both encode the record shape, and the
two encodings drift. The benefit only appears if the engine genuinely cannot be
written in the validator's language, which is a claim about the estimators and
not about the ecosystem around them.

## Choice

Go, for the validator, the engine and the operator command alike. One language,
one toolchain, one gate.

### The pinned version

`go1.26.5`. That is the version present on the machine this record was written
on, and it is what the toolchain file added by #14 pins:

    $ go version
    go version go1.26.5 windows/amd64

That command is a fact about one machine and not a claim about what the latest
release is. The pin exists so that the version is a decision somebody took
rather than whatever a runner happened to have, and #14 is where the pin
becomes a file the build reads. Raising it later is a change to that file with
its own reason, not a drift.

The pin is a floor and a ceiling in the sense the toolchain already supports:
the module declares the language version and the toolchain line, and the gate
runs with the resolved toolchain printed, so a run under a different version
says so rather than passing quietly.

### Numerical dependencies

None. Every estimator in the engine is written from its published definition,
and each one is pinned by a golden-number test against a published reference
result before it is used for anything. Issue #43 on the meta-analysis milestone
is where that obligation lands.

This is the cost of the chosen option stated plainly rather than a benefit.
Writing a random-effects estimator from a definition is how a wrong estimator
gets into a project, and the only thing standing between this project and that
outcome is the golden-number tests. A cross-check against an existing
statistics package is welcome and is kept as static fixtures committed to this
repository, never as a second toolchain the gate invokes. The difference
matters: a fixture is a number somebody read out of a named tool at a named
version and committed, which a reviewer can check, and a second toolchain in
the gate is an install every contributor pays for.

### The one decoding dependency, and what it is allowed to be

The tracked corpus format is TOML, fixed by
`docs/decisions/0003-storage-format.md`, and the standard library does not
parse it:

    $ go list std | grep -ci toml
    0

So the project carries one third-party module for decoding, and it is not a
numerical dependency. Which module and which version is pinned by the lockfile
#14 owns, and this record fixes only what it has to be: pure Go with no cgo, so
the static single-file install and the cross-platform build survive; no network
access at decode time, so it cannot break the offline guarantee; and a
permissive licence, which is constrained by whatever answer entry 1 of #13
takes for this repository.

The direction of the dependency is one way. The corpus is decoded by this
project; nothing in this project is decoded by the module.

### Package layout, and the network package

Packages live under `internal/`, one directory per package. The single package
permitted to import a network-capable API is `internal/net`, which is the path
`docs/decisions/0009-offline-by-default.md` leaves to this record when it says
the component lives where the language decision puts a package of that kind.

The check that enforces it reads the import graph the toolchain prints. That
the graph is printable by the toolchain, transitively and without building
anything, is the property the offline guarantee depends on:

    $ go list -deps net/http | wc -l
    186

The check itself does not exist. `PROSE, NOT ENFORCEMENT`, issue #65, which is
the issue that owes it. Nothing in this repository computes an import graph
today, because there is no source tree to compute one over, and the paragraph
above says what will be possible rather than what is true.

### The rendering surface

Plots emit a vector document that the tests assert against, rather than pixels
or a screenshot. That follows from the headless rule in
`docs/decisions/0010-headless-tests.md` and it is recorded here because it is
also what closes the plotting gap in the cost list above. The absence of a
plotting library stops being a cost the moment the rendering target is a
document the project writes itself.

## Reasons

The split loses first and it loses on its own terms. It is the only option
whose benefit is conditional on a claim nobody has tested, that the estimators
cannot be written in the validator's language, and it pays its full cost from
the first day whether or not that claim turns out to be true. Two toolchains in
one gate is also the specific shape that produces a gate people stop running
locally, and a gate people do not run locally is a gate that catches things
after the pull request instead of before it.

Python loses on the two promises this project has already made. The operator
quickstart on #57 claims a clean machine, no compiler, no package manager and
no network after the download, and an interpreted stack cannot honour that
without a bundling step that is a compiled artifact wearing a different name.
The reproducibility of a cited number is the second: this corpus exists because
published numbers turn out not to mean what their intervals claim, and a
project making that argument cannot ship an analysis whose number depends on
which versions resolved on the day somebody ran it. Both of those outrank the
ecosystem, and the ecosystem is the only axis Python wins.

Between the two compiled options, Go beats Rust on three things and loses on
one. It wins on the import graph, which is not a general preference but the
specific mechanism the offline guarantee was written against; on compile time,
which is paid on every local gate run and therefore decides whether the gate is
run; and on the contributor pool, since the people who can check a
transcription of a 1940 paper are not selected for systems programming. It
loses on parser safety, and that loss is real because the validator's input is
untrusted by construction. What pays for it is the static analysis on #18 and
the parser fuzzing on #52, both of which exist on this board already and are
aimed at exactly that surface. A memory-safety property obtained from a type
system is stronger than one obtained from a fuzzer, and this record is choosing
the weaker of the two knowingly rather than pretending they are equivalent.

One cost of the chosen option is not fully mitigated and is written here so it
is not discovered later. Floating-point arithmetic in Go is `float64`, and
whether two machines produce the same last digit for the same estimator is a
property of the compiler and the target architecture rather than something this
record can assert. That an implementation may contract a multiply and an add
into a single fused operation, and that an explicit conversion to the target
type prevents it, is a claim about the language specification that was not
verified against the specification text on the route that produced this record,
and it is written as a claim. What the project does about it does not depend on
the claim being right: the golden-number tests on #43 run on every platform the
release claims to support, and a digit that moves between them is a defect that
surfaces there. Until those tests exist and have run on more than one
architecture, the reproducibility of a cited number is an intention.

Nothing in this repository refuses a violation of any of this.
`PROSE, NOT ENFORCEMENT`. There is no source tree, no module file and no
toolchain pin here today, so a pull request adding a Python script, a second
toolchain or a numerical dependency passes every check this repository runs.
The commands quoted above read a Go installation that happens to be on one
machine; they read nothing in this tree. What would make the language decision
enforceable is the build entry point on #14 and the checks on #16, #17, #18 and
#65, and none of them exists yet.

## Date

2026-08-08

## Reversal condition

Reverse the single-language choice if an estimator this project needs is
demonstrated to be unwritable in Go at the accuracy the golden-number tests
demand. Demonstrated means a specific estimator, a specific published reference
result, and a failing test, not an expectation that one will turn up.

Reverse the no-numerical-dependency rule if the golden-number tests start
failing in ways that trace to numerical technique rather than to the
definitions, for example an eigenvalue or optimisation routine written here
being less stable than the published one. The signal is a test that fails for
an implementation reason after the definition has been checked twice, and at
that point a well-known library is a smaller risk than a hand-written routine.

Reverse the Go choice in favour of Rust if the validator's parser produces a
memory-safety or unbounded-resource defect that the static analysis and the
fuzzing on this board did not catch first. That is the axis Rust won and this
record traded away, so it is the axis on which the trade is judged.

Revisit the pinned version whenever the pinned toolchain stops receiving
security fixes. That is a date, not a judgement, and it should move the pin
rather than be argued about.
