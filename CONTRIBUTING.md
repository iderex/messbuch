# Contributing

## Read this first: contributions from outside are not accepted yet

The maintainer decided on 2026-08-09 that this project does not take
contributions from people outside it until the first release. Nothing has been
released:

    gh release list --repo iderex/messbuch --limit 5
    (no output)

The reason is that a contribution made now would be written against a schema
that is still moving, and the review would land before the validator that is
supposed to carry it exists. The reason it is written here, at the top, is
separate and matters more: a refusal you discover after doing the work is the
refusal plus the work. That decision is recorded on issue #13, entry 4, and the
fleet-wide position from 2026-08-08 that outside contributions are welcome under
a sign-off is dated rather than cancelled by it. It applies from the first
release.

Two things you can do today, and both are wanted.

Report a number that is wrong. Use the form for it, which asks three things and
nothing else, and you do not have to be sure:
[A number in the corpus is wrong](https://github.com/iderex/messbuch/issues/new?template=wrong-number.yml).
What happens to a report afterwards is `docs/corrections.md`.

Argue with a decision. Every decision that shapes this project is written down
under `docs/decisions/` with its reasons and its reversal condition, and the
tracker is where they are argued. A decision record is easier to argue with than
a pull request.

The rest of this document is what a contribution will have to satisfy when the
gate opens, and it is written now so that it is not written in a hurry later.

## The one command

    go run . ci

That is the whole local gate. It is one command rather than a list of steps in a
document, because a list drifts against what actually runs and then misleads the
person following it most carefully. The command runs its legs in order, stops at
the first failure, and prints what each one examined. A leg that is not built
yet prints that it was not run and what is owed, so a run that covered less than
everything cannot be read as one that covered it and found nothing.

The legs are not listed here. Run it and read what it says.

Nothing to install beyond a Go toolchain at the version `go.mod` pins. The first
leg compares the release you are running against that pin and refuses a mismatch,
so a wrong toolchain is a named failure rather than a strange one.

One more command exists and it is not a gate:

    go run . fmt

It writes the bytes the formatting leg demands. Check and fix go through the
same function, so there is no gap between what the gate wants and what the fix
produces.

## What runs on a pull request

Not listed here either, and for the same reason. The set moves, and a copy in
this document would be wrong on the day it moves. Read it off a real pull
request:

    gh pr checks <number> --repo iderex/messbuch

`docs/required-checks.md` is where each check name is argued: what it refuses,
whether it reports on every pull request, and which two names must never be
required. It is a proposal for the branch protection rather than a list to
follow.

## Signing off a commit

Every commit carries a `Signed-off-by` trailer matching its author, and a check
refuses a pull request where one does not:

    git commit -s

The trailer is a certification, not a formality. Its text is the Developer
Certificate of Origin, in `DCO` at the root of this repository, and it says that
you wrote the change or otherwise have the right to submit it under the
project's license.

The license is AGPL-3.0, in `LICENSE`. The corpus under `record/` and
`vocabulary/` is separate and is CC BY 4.0: what is licensed there is the
collection rather than the individual measured value, and a transcription of a
published number counts as a contribution to the database rather than as a
statement of fact with no rights attached. Both decisions are on issue #13,
entries 1 and 2.

## Contributing is public, permanently

The name and address in your commits enter a public git history and stay there.
Rewriting that history later is not something this project can promise, and
anyone who needs to contribute under a different identity should decide that
before their first commit rather than after. This is a description of how git
works rather than a promise about anything this project builds, which is why it
is stated flatly and carries no mitigation.

## What a good record contribution looks like

`docs/curation.md` is the guide, and it is written for a careful person who is
not a programmer. It is not summarised here, because a summary of a guide people
follow literally is a second copy that drifts, and a rule that moves without it
produces a batch of records that have to be redone.

Two things from it are worth knowing before you open it. One record is one
published number from one source in one file. And every number comes from the
paper it was published in, with a locator inside that paper: a review is useful
for finding the papers and is not a source for the numbers.

## What the review checks

The gate decides whether the tree is sound. The review decides the things no
machine here reads, and they are the ones worth your attention:

- Whether the number matches the source, and whether the locator lands where it
  says. This is the whole point of the corpus and it is checked by a person.
- Whether an asserted fact carries the command that produced it, run at the
  commit being pushed rather than in a working tree. A number without its
  command is a claim, and it is written as one or not written.
- Whether the change makes a document wrong. `docs/downstream-documents.md` maps
  which documents are downstream of which parts of the tree, and the pull
  request template asks the question in a form that cannot be ticked without
  answering.
- Whether a guard ships with proof that it bites, for the reason it names. A
  check nobody has watched fail is a check nobody knows is wired up.
- Whether one pull request carries one topic.
- Whether a negative stayed negative. If a passage says something was not done,
  the admission survives the edit rather than being softened into a claim that
  it was.

## What is not enforced here

No check in this repository reads this document. Nothing compares it against the
gate, nothing refuses a pull request that skipped a step in it, and nothing
verifies that the command above is the one the workflows run. The review is
where a departure is caught.

Saying so is part of the document rather than a caveat at the end of it. A rule
that pretends to be enforced is worse than one that admits it is not, because a
reader who believes a machine is watching stops watching.
