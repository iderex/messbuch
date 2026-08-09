# 0018 Whether a corpus release gets an archived identifier, and under what name

## Question

A citable corpus wants a permanent identifier that still resolves if this
repository disappears. Does the project deposit its releases with an external
archive to get one, and if so, what name does the deposit carry?

The second half is not a detail. A deposited version is immutable, so whatever
name is on it is on it permanently and publicly.

## Options considered

For the identifier: deposit each release with an external archive and take a
persistent identifier per release; deposit nothing and let the repository and
its tags be the only reference.

For the name on the deposit: the handle this repository is published under, or
the maintainer's own name.

## What each option costs

Depositing each release. The corpus becomes citable by somebody who needs a
stable reference, which is most of the audience this project is aimed at. The
costs are an account and a standing relationship with a third party, publication
under whatever terms that archive requires, and immutability: a deposited
version cannot be withdrawn, so a transcription error that ships is permanent.
That immutability is the point of the exercise and the risk of it at the same
time.

Depositing nothing. No third party, no account, nothing that cannot be undone.
The cost is that the corpus is not citable by anyone who needs a reference that
outlives a repository, which removes the audience the corpus exists for.

The handle on the deposit. The maintainer's name stays off a permanent public
record. The cost is that the work is attributable only to a handle, and a handle
in a citation is a claim rather than a credential.

The maintainer's name on the deposit. The work is attributable to its author, and
in a curriculum vitae, a funding application or a citation a name carried by an
author identifier is evidence. The cost is that the name and the attribution
stand permanently and publicly in every deposited record, and a deposit cannot
later be turned back into a pseudonym. This is the direction that does not
reverse.

## Choice

Each release is deposited and takes a persistent identifier of its own. A
`CITATION.cff` file accompanies it so that a reader citing the corpus does not
have to construct the citation by hand.

The deposit is tied to a release rather than to every state of the branch,
which is what keeps the immutability cost bounded: a transcription error is
permanent only once it has been through a release rather than the moment it is
pushed.

The deposit runs under the maintainer's own name, Nils Lehnen, rather than
under the handle. The pseudonymous option was on the table and was set aside
the same day.

## Reasons

Citability is the whole reason for depositing at all, and it is only worth
having if the identifier resolves after this repository does. Nothing about the
tags on a branch gives that.

Attaching the deposit to a release rather than to every state is the answer to
the one cost that cannot be undone. It does not remove the risk that a wrong
number ships permanently; it reduces the number of occasions on which it can.
`docs/corrections.md` is where a wrong number that has already shipped is
handled, and a corrected value never overwrites the record of the old one.

On the name, the two costs point in opposite directions and neither is small.
The one that decided it is that a citation carrying a handle asserts authorship
and a citation carrying a name with an author identifier evidences it. The
permanence was seen and accepted rather than overlooked.

## What this does not do, and what is still owed

None of it exists in the tree:

    git ls-files | grep -Ei 'CITATION'
    (no output, exit 1)

    gh release list --repo iderex/messbuch --limit 5
    (no output)

    git tag | wc -l
    0

Three things follow from the choice and are work rather than further questions.
An author identifier belongs with the name, otherwise the name connects this
archive's deposits to nothing outside it. `CITATION.cff` has to carry the same
spelling of the name as the deposit, because two spellings in two places are two
people to anything citing automatically. The copyright line in `LICENSE` already
carries the name, and it and the deposit must not drift apart.

`PROSE, NOT ENFORCEMENT`. Nothing here refuses a release cut without a deposit,
nothing compares a `CITATION.cff` against the copyright line, and no archive is
named in this record because none has been chosen. `docs/release-checklist.md`
is where a release is walked, and it is a list a person reads.

## Date

2026-08-09

## Reversal condition

Reverse the deposit itself if the chosen archive's terms turn out to conflict
with the corpus license in `0017-corpus-license.md`, since a deposit made under
incompatible terms cannot be withdrawn afterwards. Check that before the first
deposit rather than after it.

The name on the deposit has no reversal condition, and that is a statement
about the decision rather than a missing part of this record. Once a version is
deposited under a name it stays deposited under that name, so there is nothing a
later decision could return to.
