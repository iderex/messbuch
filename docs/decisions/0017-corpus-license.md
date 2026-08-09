# 0017 The license of the corpus, which is not the license of the code

## Question

Under what terms is the corpus published, and what exactly is it that a license
over it would cover?

Giving a table of numbers a code license is a known source of confusion,
because concepts like linking and source form do not map onto it. An individual
measured value is a fact, and facts are not protected by copyright in most
jurisdictions. The selection and arrangement of a database can be protected as
a database work, and in the European Union a separate database right attaches
to the collection independently of copyright. So the subject of this decision
is the collection, not the number.

## Options considered

CC0 1.0. CC BY 4.0. ODbL 1.0. PDDL.

## What each option costs

CC0 1.0. A dedication to the public domain as far as the law allows, with
maximum reuse and no friction for anyone aggregating this with other datasets.
The cost is that there is no attribution obligation at all, so the corpus can be
absorbed into a product with no trace, and nothing obliges anyone to publish a
correction back.

CC BY 4.0. Attribution required. The cost is attribution stacking: a downstream
aggregator combining twenty sources carries twenty notices, and that is a real
reason aggregators leave a dataset alone.

ODbL 1.0. Share-alike for databases, so a derived database has to be published
under the same terms. The cost is that it is the most restrictive of the four
for reuse, and its definitions of derived work and produced work are hard to
apply with confidence.

PDDL. A public domain dedication aimed specifically at databases, so it
addresses the database right directly rather than by analogy. The cost is that
it is much less widely recognised than CC0.

## Choice

CC BY 4.0, over the collection.

The scope inside this decision is settled in the same breath, because leaving
it open would put holes in the corpus where nobody can say what terms its
contents are under. The data license covers the vocabulary and the
transcriptions alike. A transcription of a published number counts as a
contribution to the database rather than as a rights-free statement of fact.

`CONTRIBUTING.md` is where a contributor meets this before doing the work:

    git show origin/main:CONTRIBUTING.md | grep -n 'CC BY 4.0'
    87:`vocabulary/` is separate and is CC BY 4.0: what is licensed there is the

## Reasons

Attribution is the only one of the permissive options under which the work
stays visible. CC0 would have removed the attribution friction entirely and let
the corpus disappear into a product without corrections ever coming back, which
is the failure this project can least afford: a corpus is worth what its
provenance is worth, and a copy circulating with no trail back to the
transcription is a copy nobody can correct.

ODbL was set against the attribution stacking cost and lost on its own
definitions rather than on its strictness. Where a term is hard to apply with
confidence, a contributor cannot tell whether they are complying, and the
license stops being a rule and becomes a risk.

PDDL addresses the database right more directly than CC0 does, and the reason
it was set aside is the cost this record already states for it rather than a
further argument: recognition. Nothing beyond that was recorded, and nothing
further is supplied here.

`PROSE, NOT ENFORCEMENT`. No file in the tree carries the CC BY 4.0 text today,
and no check compares what `CONTRIBUTING.md` says about the corpus against what
this record says or against what any file under `record/` or `vocabulary/`
carries:

    git ls-files | grep -Ei 'LICENSE'
    LICENSE

One file, and it is the AGPL-3.0 text for the code. A corpus license stated in
prose and carried by no file is weaker than one a redistributor can copy, and
that gap is real rather than a formality.

## Date

2026-08-09

## Reversal condition

Reverse this if attribution stacking is observed to be the reason an aggregator
declined the corpus, because at that point the obligation is costing the
project the reach it was chosen to protect. The signal is a named aggregator
saying so, not a supposition about one.

Reverse it also if the database right the choice is built around is narrowed or
removed in the jurisdictions that matter here, since a license over a
collection that no longer carries a protectable interest constrains only the
people who read it carefully.
