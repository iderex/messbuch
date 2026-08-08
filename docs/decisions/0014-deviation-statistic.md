# 0014 The deviation statistic and the reference value it needs

## Question

The by-product this project exists for is a statement about how far published
measurements sat from the truth, in units of the uncertainty their own authors
quoted. Producing that statement needs three things fixed before any number is
computed from them: what the statistic is, what stands in for the truth, and
what happens to a measurement whose uncertainty the source did not quote.

The stand-in is the contestable part and it is the reason this record exists.
There is no truth available to compare against, only later numbers, and every
later number is entangled with the earlier ones in some way. Choosing the
stand-in after seeing what each one does to the answer would make the result
worthless, so the choice is made here and the sensitivity to it is published
rather than resolved privately.

This record states no result. What the deviation distribution actually looks
like is the measurement #46 asks of the corpus, it has not been made, and
nothing below anticipates it.

## Options considered

Four stand-ins for the truth.

The modern accepted value: the current recommended value from whichever body
adjusts that quantity, or the current published consensus where no body does.

A later high-precision measurement, chosen so that it is independent of the
series being scored.

A pooled estimate over the series with the measurement being scored left out.

The resolved value, under a restriction of the whole analysis to quantities
that were later pinned to a precision far beyond anything in their historical
series.

Across all four, a fifth question: whether one of them is fixed in the code, or
whether the tool supports several and the reader is shown the difference.

## What each option costs

The modern accepted value. It is defined for nearly every quantity this corpus
will hold, it is precise, and it is what a reader means informally by the true
value. It is also what the prior literature on this effect used, so it is the
only one of the four under which a result here can be compared with the results
this project is answering. Its cost is a correlation and the correlation is not
uniform. A recommended value is usually itself a weighted combination whose
weight sits almost entirely on the most recent and most precise members of the
series, so the late members of a series are scored against a number they
largely determined, and their deviations are pulled toward zero. That shrinks
exactly the tail this project is trying to measure. The cost has a second half
that is easy to miss: a recommended value is a moving object with its own
revision history, so a deviation computed against one vintage and a deviation
computed against the next are not the same statistic, and a number quoted
without its vintage cannot be reproduced.

A later independent high-precision measurement. It removes the correlation
cleanly where it can be had. It is unavailable for many quantities, and the
word independent is doing work that the literature usually does not support:
independence of apparatus, of input constants and of the people involved is a
judgement a transcriber would have to make, and no field of this corpus records
it. Choosing which later measurement counts is itself a selection with no rule
behind it, which is the shape of decision this record exists to avoid. Worse
for this project specifically, the quantities where the option fails are the
ones that motivated the corpus: a quantity whose two method families still
disagree has no later measurement that both sides accept, so the series most
worth scoring are the ones this option cannot score.

A pooled estimate with the scored measurement left out. It is available for
every series with more than one member, it removes the scored measurement's own
weight from its own reference by construction, and it needs no external
authority. It inherits the pooling model's assumptions whole, and under
`docs/decisions/0013-pooling-default.md` the default model estimates a
between-study variance which is the quantity this corpus was assembled to
study, so the reference and the statistic would rest on the same premise. Its
real hazard is quieter than that. A systematic shared by every member of a
series moves the leave-one-out reference by nearly the same amount as it moves
the measurement, so the deviations come out small while every measurement in
the series was wrong together. That failure leaves no trace in the output. It
also does not remove the correlation it is credited with removing so much as
soften it: the early members of a series are still scored against a pool
dominated by the later precise ones.

The resolved value under a restricted analysis. It is the cleanest of the four
and the one whose reference is closest to deserving the name. It shrinks the
sample hard, and the survivors are not a random sample of anything. A quantity
gets a high-precision follow-up because somebody thought it was worth
resolving, which correlates with somebody suspecting a problem, and a
distribution built only from those is a distribution of the cases where a
problem was suspected.

Fixing one reference in the code. It makes every number in the output
comparable with every other and removes an option nobody would know how to set.
Its cost is that the single most contestable decision in the project becomes
invisible at the point where the result is read, which is where it matters, and
that a reader who disagrees with it has no way to see what their own choice
would have given.

## Choice

### The statistic

For a measurement with published central value `x` and a reference `r`:

    z = (x - r) / s

Four things about that formula are fixed here.

The denominator is the measurement's own quoted uncertainty and nothing else.
The reference's uncertainty is not combined into it. The question this corpus
asks is whether the interval printed in the paper means what it claims, so the
denominator has to be that interval rather than a repaired version of it.
Combining the two would move every deviation toward zero by an amount that
depends on the reference, which would make the result partly a statement about
the reference rather than about the literature.

`s` is the total the source quoted where the source quoted a total, and
otherwise the quadrature sum of the components it quoted, and which of the two
was used is recorded on every scored measurement. The distinction is not
pedantry. A quadrature sum over components an author did not intend to be
combined is our arithmetic and not theirs, and it assumes an independence
between statistical and systematic parts that the source did not state. Where
a component carries `correlation_group`, the analysis says whether it treated
the group as shared, because `docs/decisions/0005-uncertainty.md` stores the
group as a name and refuses to invent a matrix, and an analysis that assumed
independence there has assumed something.

An asymmetric interval is used on the side facing the reference. If `x` is
below `r`, `plus` is the denominator; if above, `minus`. Averaging the two
sides would erase the asymmetry in precisely the cases where the author put it
there deliberately.

`z` keeps its sign. The signed distribution is what shows whether published
values approached the modern value from one direction, which is the bandwagon
question, and the distribution of the absolute value is a view derived from it
rather than the stored statistic.

### The reference

The default reference is the modern accepted value, and the vintage of that
value is recorded with every number computed from it.

A measurement is admitted to the headline distribution only where the reference
interval is at most one third of the measurement interval. The reason is
arithmetic rather than convention, and it is the same arithmetic in both
directions:

    awk 'BEGIN{for(f=2;f<=5;f++) printf "ref 1/%d as wide: sqrt(1 + 1/%d^2) = %.4f\n", f, f, sqrt(1+1/(f*f))}'
    ref 1/2 as wide: sqrt(1 + 1/2^2) = 1.1180
    ref 1/3 as wide: sqrt(1 + 1/3^2) = 1.0541
    ref 1/4 as wide: sqrt(1 + 1/4^2) = 1.0308
    ref 1/5 as wide: sqrt(1 + 1/5^2) = 1.0198

At one third, treating the reference as exact overstates the precision of the
comparison by about five per cent, which is small against the effect sizes this
corpus is about. The same rule does the second job the modern accepted value
needs done: a measurement precise enough to have dominated the recommended
value it is being scored against is, by construction, a measurement whose
interval is comparable to the reference interval, so the admission rule removes
the members where the correlation bites hardest without anyone having to know
how the adjustment was weighted. The ratio is carried on every measurement and
the excluded count is printed with the result, so the rule is visible rather
than silent.

The tool supports all four references and hard-codes none. Every output names
which reference was used and its vintage, which
`docs/decisions/0011-corpus-versioning.md` already requires of every option in
force including the ones left at their default. Any headline number is
published with the same statistic recomputed under the other three beside it.
Where those four disagree materially, that disagreement is the result and is
reported as one rather than resolved by preference.

### Measurements the statistic is not defined for

`docs/decisions/0005-uncertainty.md` fixes what an analysis may assume when a
field is absent, and this record does not restate that list. What it settles is
what this statistic does at each of those states.

An `uncertainty_status` of `none-in-source` has no denominator. No `z` is
computed, the measurement is excluded, and the uncertainty is never imputed
from a later paper or from its neighbours. A statistic in units of a quoted
uncertainty cannot be computed for a measurement that quoted none, and
supplying one would put a number this project invented into the distribution
this project is measuring.

An `uncertainty_status` of `not-transcribed` is excluded and counted
separately from the above, because one is a fact about the literature and the
other is a fact about how far the transcription has got.

A `coverage` of `unstated` is excluded by default. Assuming a coverage is an
option, and the assumed value appears in the stamp with the result.

A record that is a limit has no central value and gets no `z`. Whether the
reference falls inside the excluded region is a different statement about a
different object, and folding it in would put two statistics in one column.

A component of zero width is a data defect to be checked against the source,
per the same record, and not a denominator.

All four counts are printed with the result. How many published measurements
cannot be scored at all, and why, is a finding about the literature rather than
housekeeping, and a distribution reported without them invites the reader to
assume the excluded set was empty.

### The selection effects, stated with the result and not only here

Five, and every one of them narrows what the final probability statement is
about.

Series that were later resolved are not a random sample of series. Resolution
tends to follow interest, and interest tends to follow trouble.

Quantities that received a high-precision follow-up are not a random sample of
quantities.

The admission rule above is itself a selection. It keeps the older and less
precise members of a series preferentially, and those are the members most
likely to deviate, so the distribution it produces is not the distribution over
all published measurements and may not be described as one.

Which series were transcribed first is a selection made by this project. A
series is chosen partly because it is known to misbehave, and the seed corpus
was chosen on exactly that basis.

A measurement that was never published is not in the corpus at all. That
missingness is what the funnel and selection diagnostics on the analysis
milestone address, and this statistic does not see it.

## Reasons

The modern accepted value wins on the axis that separated the four, which is
whether the failure mode of the reference is visible in the output. Its failure
is a correlation with the most precise members of the series, it acts in a known
direction, it is bounded by the admission rule, and it can be measured directly
by recomputing the same statistic leave-one-out, which the tool does anyway. The
leave-one-out estimate's worst failure, a systematic shared across a whole
series, produces small deviations and no signal that anything went wrong; a
reference whose failure is silent is worse for this project than one whose
failure is loud, because the whole deliverable is a claim about how often the
literature was quietly wrong. The independent later measurement is the best
reference where it exists and cannot be the default, since the series it cannot
score are the ones the corpus was built around. The resolved-value restriction
is kept as one of the four rather than as the frame, because a result computed
only on resolved quantities is a narrower claim than the one #47 is trying to
make and should be shown as the comparison it is.

Supporting all four rather than one is not fence-sitting. The single number this
project will be quoted on rests on this choice more than on any other, and a
sensitivity that is computed but not published is a sensitivity the author
looked at and the reader did not. `docs/decisions/0013-pooling-default.md`
already takes the same position for the pooling model, and the argument is the
same argument.

The exclusions are the part most likely to be relaxed later by somebody in a
hurry, so the reason is written where they will meet it. Every one of the
excluded classes could be included by supplying a number: an assumed coverage,
an imputed uncertainty, a symmetrised interval. Each of those substitutions
makes the corpus look better behaved than it is, because each replaces a wide or
unknown quantity with a conventional one. A corpus assembled to test whether
quoted intervals can be trusted is the last place that may fill in the missing
ones with a convention.

Nothing in this repository computes, refuses or checks any of this. `PROSE, NOT
ENFORCEMENT`, and here it is the whole of the rule rather than a residual. There
is no source tree, no estimator and no corpus. #45 holds the implementation that
would support more than one reference, #47 holds the probability statement that
consumes it, #44 holds the surface that would print the stamps and counts
required above, and #43 holds the check that pins any of it against published
numbers. Until those land this record is an argued position and not a
description of behaviour anybody has observed.

## Date

2026-08-08

## Reversal condition

Reverse the default to the leave-one-out pooled estimate if the corpus shows
that the admission rule is discarding a large enough share of the scorable
measurements to make the headline distribution unrepresentative of the
literature. The trigger is a count rather than a judgement: the excluded share
is printed with every result, and the question is whether what survives is still
a sample worth describing.

Revisit the one-third admission ratio if the shape of the deviation distribution
turns out to be sensitive to it. The five per cent figure above is an argument
about a comparison of interval widths and not a measurement of what the ratio
does to the answer, and the second is what should decide it once there is a
corpus to measure it on.

Revisit the whole record if a reference class arrives that none of the four
covers, for instance a quantity whose accepted value was later found to be wrong
and revised by more than its own stated interval. That case scores every
historical measurement against a number the field itself retracted, and none of
the four options above says what to do with it.
