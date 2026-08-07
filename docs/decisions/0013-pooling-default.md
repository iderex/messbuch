# 0013 Which pooling model is the default

## Question

When several published measurements of one quantity are combined into a single
estimate, which model does this tool use unless told otherwise, and why is the
scale factor approach used inside particle physics not it?

This is a decision about a default rather than about which methods exist. All of
them can be asked for. The question is which one answers when nobody asked.

## Options considered

The fixed effect model, weighting each measurement by the inverse of its
variance and treating all scatter beyond the quoted errors as chance.

The random effects model, allowing the underlying value to differ between
experiments and estimating that between-study variance from the data.

The scale factor approach, taking the fixed effect central value and inflating
its interval when the measurements disagree by more than their quoted errors
allow.

No default at all, refusing to pool until the operator names a model.

## What each option costs

The fixed effect model. It is the most efficient estimator available if its
assumption holds, it is the one every reader recognises, and it has no tuning.
Its cost on this corpus is fatal and specific: the assumption it makes is the
proposition under test. This corpus was assembled to study the case where the
quoted errors are wrong, and a model whose premise is that all scatter beyond
the quoted errors is chance answers that question before it is asked. Its
interval is also too narrow whenever the premise fails, so the failure makes the
result look more precise rather than less, which is the wrong direction for a
default to fail in.

The random effects model. It estimates the thing this project wants to measure
rather than assuming it away, and its interval widens when the measurements
disagree. It costs several real things. The between-study variance is badly
determined when there are few studies, which is the regime most series here sit
in. Which estimator is used for it changes the answer, so the choice cannot be
silent. The interval needs a small-sample correction that matters most in
exactly that regime. And it moves weight toward the less precise measurements
relative to the fixed effect model, which is a hazard rather than a virtue when
the less precise measurements are the ones most likely to be biased. There is
also an interpretive cost that has to be paid in words rather than in
arithmetic, set out under the reasons below.

The scale factor approach. It is the incumbent inside one field, its published
values are what a reader of that literature will compare against, and it is
conservative in the sense that it widens an interval rather than narrowing one.
Its costs are the reason this record exists. It keeps the fixed effect central
value, so it assumes that whatever caused the disagreement did not move the
centre, which is an assumption about a mechanism nobody has identified. It has
no model of the between-experiment variance, only a repair applied to the
interval, so it cannot produce a prediction interval for what the next
measurement will find, and the prediction interval is the statement a reader of
this corpus actually wants. It folds the disagreement into the error bar, which
makes the disagreement disappear from the output as a quantity. And it is
defined with conventions local to one field, including when it is applied at all
and which measurements are admitted first, so a cross-disciplinary corpus that
adopted it would be importing those conventions along with it.

No default. The most honest option in one narrow sense and unusable in every
other. It makes the tool require a methodological decision before it produces
anything, which most operators will resolve by copying whatever the first
example in the documentation uses, so the default returns as folklore instead of
as an argued position. It also makes two series incomparable unless the operator
happened to choose identically, and comparing series is the whole point.

## Choice

The default is the random effects model.

The fixed effect model is available, is never the default, and every output
names which model produced it.

More than one estimator for the between-study variance exists, the output names
which one was used, and the output says whether the small-sample interval
correction was applied. None of these is inferable from the number, so none of
them may be left out of the stamp, which
`docs/decisions/0011-corpus-versioning.md` already requires for every option in
force including the ones left at their default.

The scale factor approach is available as a comparison view and is never the
interval this tool reports as its own. Its purpose here is to reproduce what the
incumbent method would have said about the same records, so that the difference
between the two traditions is a number this corpus can show rather than a claim
it makes. An output produced that way is labelled as such and is not presented
beside a random effects interval without that label, because two intervals of
different provenance sitting in one column is how a comparison becomes a
substitution.

The output states, in words and not only in a field, what the between-study
variance is being read as. See below.

## Reasons

Every rejected option fails on the same axis, which is what the default is
allowed to assume about the thing being studied. Fixed effect assumes the
scatter is chance. The scale factor assumes the centre is unaffected and the
excess belongs in the error bar. No default assumes an operator will make the
choice deliberately, which is the assumption least supported by how defaults
actually get used. Random effects assumes least: it lets the data say how much
the experiments disagree beyond their quoted errors, and that quantity is the
subject of this corpus rather than a nuisance in it.

The interpretive cost has to be paid explicitly, because leaving it unpaid would
make the tool say something false about physics. In the medical tradition where
this model comes from, the between-study variance is usually read as real
variation in the underlying effect, because the effect genuinely differs between
populations. For a physical constant there is one value and it does not vary
between laboratories. So the between-study variance estimated here is not
variation in the quantity. It is the variance of whatever the experiments got
wrong and did not report, which is exactly the unquantified systematic error this
corpus exists to look at. That reading is stated in the output rather than left
to a reader's background, because a reader from either tradition will otherwise
supply the wrong one, and the wrong one from the physics side reads as a claim
that a constant is not constant.

Keeping the scale factor approach as a comparison rather than dropping it is
deliberate. This project's stated position is that the medical tradition is
being imported in place of home-built scale factors, and a substitution argued
against a method the tool cannot reproduce is an assertion. Being able to print
both from the same records is what makes the argument checkable, and it is also
the honest way to be wrong in public if the incumbent turns out to be better
behaved on this corpus than expected.

Nothing in this repository implements or refuses any of this. `PROSE, NOT
ENFORCEMENT`, and here it is the whole of it rather than a residual. There is no
source tree, no estimator and no corpus; #38 holds the models, #43 holds the
check against published worked examples, and #44 holds the surface that would
print the labels required above. Until those land, this record is an argued
position and not a description of any behaviour.

## Date

2026-08-07

## Reversal condition

Reverse the default to the fixed effect model if the corpus itself shows that
the between-study variance is consistent with zero across series once the
seed corpus is large enough for that statement to mean anything. At that point
the fixed effect model is no longer answering the question by assumption, it is
answering it correctly, and defaulting to the heavier model would be its own
kind of thumb on the scale. This is a measurement, not a judgement: the trigger
is the distribution of the estimated between-study variance across series, and
the tool that produces it is the one this record is about.

Reverse the choice of default between-study variance estimator, which is a
smaller decision inside this one, whenever a better-behaved estimator in the
small-study regime becomes standard. The regime matters more here than the
literature's general advice, because most series in this corpus have few
measurements.

Revisit the comparison view if it is observed being cited as this project's own
result. A labelled comparison that gets quoted without its label has stopped
being a comparison, and at that point the cost of publishing it exceeds what it
buys.
