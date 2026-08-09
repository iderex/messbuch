# 0016 The license of this repository

## Question

Under what terms is the code in this repository published, so that somebody
outside the project can use, fork or package it and know what they owe in
return?

The question blocked more than it looked. The sign-off check on every pull
request asks a contributor to certify a right to submit their work under the
project's license, and while there was no license the certification pointed at
nothing. The answer is also the hardest of any here to change afterwards, since
changing it requires the agreement of everyone who has contributed under the
old one.

## Options considered

MIT or BSD. Apache 2.0. GPL-3.0-only. AGPL-3.0-only.

## What each option costs

MIT or BSD. The shortest texts and the widest reuse, including inside closed
products. The cost is no patent grant, and no obligation on anyone who improves
the tool to publish the improvement.

Apache 2.0. Permissive in the same way, with an explicit patent grant and a
patent retaliation clause. The cost is a longer text, a notice obligation on
redistributors, and one-way incompatibility with GPL-2.0 code if any is ever
wanted.

GPL-3.0-only. Anyone distributing a modified tool has to publish their changes
under the same terms. The cost is that closed-source distributors will not
touch it, and some downstream ecosystems refuse it by policy.

AGPL-3.0-only. As above, and the obligation extends to somebody who runs the
tool as a hosted service without distributing it. The cost is that it is the
strongest deterrent to institutional adoption of the four, and several large
organisations forbid it outright.

## Choice

AGPL-3.0-only. The full text is at `LICENSE` at the repository root and its
copyright line reads `Copyright (C) 2026  Nils Lehnen`.

    git show origin/main:LICENSE | grep -n 'Copyright (C) 2026'
    633:    Copyright (C) 2026  Nils Lehnen

    gh api repos/iderex/messbuch --jq '.license.spdx_id'
    AGPL-3.0

This covers the code. The corpus under `record/` and `vocabulary/` is a
separate decision and is `0017-corpus-license.md`.

## Reasons

No reason was recorded with the decision, and none is supplied here on the
maintainer's behalf. What this record can state is the cost that is now carried
rather than avoided: of the four options this is the one institutions are most
likely to refuse, and that was known when it was taken.

One observation, and it is an observation rather than a reason anybody gave.
The clause that separates this option from GPL-3.0-only is the one that reaches
somebody running the tool as a hosted service, and running the analysis behind
a web form is the use the question itself named as the likely one. Nothing
states that this is why the option was chosen.

`PROSE, NOT ENFORCEMENT` for what the file at the root says. The platform reads
the file and reports an SPDX identifier, which is the command above, but no
check in this repository compares that identifier against this record, and none
refuses a change to `LICENSE` that leaves this record stating the old answer.

Two things the decision unblocked, and both are facts about the tree rather
than arguments. The sign-off certification now points at a file. Item 9 of
`docs/release-checklist.md`, which blocked a release outright while it was
unanswered, has an answer.

## Date

2026-08-08

## Reversal condition

Reverse this if the project wants code from a GPL-2.0-only host or ecosystem
that cannot take AGPL-3.0 material, since that is an incompatibility no wording
here resolves. Reverse it also if the deterrent this option's cost names turns
out to be the thing keeping the corpus from the institutions that hold the
published record, because the corpus is worth more to this project than the
tool is.

Reversal is not free after contributions arrive from outside, which is why it
is written down now while every commit has one author. The relicensing cost
grows with the contributor count and never shrinks.
