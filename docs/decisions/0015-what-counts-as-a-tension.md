# 0015 What counts as a tension, what counts as real, and what the probability is about

## Question

The sentence this board exists to be able to write is that a published three
sigma tension has historically been real with some probability. Three
definitions sit between the deviation statistic and that sentence, and none of
them is settled by arithmetic.

What event is a tension. Two measurements of one quantity disagreeing by three
combined standard deviations is not the same event as one measurement sitting
three of its own standard deviations from an accepted value, and the two have
different base rates.

What makes a tension real. That can only be decided by what happened
afterwards, and for a large share of historical tensions nothing conclusive
happened at all. The rule for those cases decides the answer more than the rule
for the clear ones does.

What the probability is about. A probability without a reference class is not a
number, and the honest class here is narrow: tensions of this size, in this
corpus, in these disciplines, in this period.

`docs/decisions/0014-deviation-statistic.md` supplies the statistic and the
reference underneath all of this and is not restated here. This record states
no result, and no probability has been computed from anything.

## Options considered

### For the event

The deviation event: one measurement sitting at least `k` of its own quoted
standard deviations from the reference, taken directly from the statistic in
`0014`.

The pair event: two measurements of one quantity whose intervals disagree by at
least `k` combined standard deviations.

The split event: two families of measurements, usually two method families,
whose pooled estimates disagree by at least `k`. This is the shape of the
disagreements that motivated the corpus.

One of the three, or several computed and kept apart.

### For real

The later value falls on the side the tension pointed to, judged against a
later measurement admitted under the precision rule in `0014`.

The tension was closed by an identified error, meaning somebody published what
went wrong, which the correction path in
`docs/decisions/0012-where-correction-history-lives.md` can carry.

The modern accepted value falls outside the interval of the measurement the
tension was against.

And, for the cases where none of the above ever happened: drop them and report
the probability over what is left; count them as not real; or treat them as
censored and report what the answer would be at both ends.

### For the reference class

Stated once in the documentation and left out of the output. Stated in the
output as free text. Or carried as required structure that no output format can
omit.

## What each option costs

The deviation event. It is computable for every scorable measurement in the
corpus, it needs no pairing, and it is exactly the statistic `0014` already
defines. It is also the event furthest from what a reader means by a tension.
A reader saying three sigma tension is usually describing a disagreement
between two results, not a distance from a table value, and a probability
computed on the deviation event and quoted against the reader's meaning would
be answering a different question with a confident number.

The pair event. It is what most readers mean and it is what most published
claims of tension are. Its costs are combinatorial and selective. A series with
`n` scorable measurements has `n(n-1)/2` pairs, the pairs are not independent
of one another, and a single badly wrong measurement generates a tension with
every other member of its series. Counting those as separate historical cases
would let one bad measurement dominate the base rate.

The split event. It is the closest to the disagreements that motivated this
corpus and the closest to how a field actually experiences a tension. It needs
the families to be defined, which means it inherits the method coding entirely,
and where the coding is loose the event is not well defined. It also produces
very few cases, since most quantities in the corpus will not have two
well-populated method families.

Computing several and keeping them apart. It costs output surface and it costs
a reader who wants one number. It buys the only thing that makes any of them
quotable, which is that the number arrives attached to the event it was
computed on.

Real by the later value falling on the tension's side. It is the closest thing
to the plain meaning of the question and it is computable wherever a later
admissible measurement exists. Its cost is that it inherits the reference
choice and the admission rule from `0014`, so a tension can change state when
the reference vintage moves.

Real by an identified error. It is the strongest evidence available, because
somebody found and published the mechanism. It is also the rarest, it is
recorded unevenly across fields, and its availability correlates with how much
attention a quantity received, which is the same selection that runs through
everything else here.

Real by the modern value falling outside the earlier interval. It is easy to
compute and it collapses the question into `0014` again, which means it cannot
distinguish a tension that was real from a measurement that was simply wrong in
a direction nobody argued about.

Dropping the unresolved cases. It produces the cleanest arithmetic and the most
misleading number of the three. The cases where nothing conclusive happened are
not a random subset: a tension that got resolved is one somebody thought worth
resolving, and dropping the rest reports a probability conditioned on that
attention while presenting it as a probability about tensions.

Counting the unresolved as not real. It is defensible in one narrow reading and
it silently answers the question in one direction. It converts our ignorance
into evidence against, at whatever rate ignorance happens to occur.

Treating the unresolved as censored. It refuses to answer where the data does
not, and it produces a range rather than a number, which is harder to quote and
is the point. Its cost is that where the unresolved share is large the range is
wide enough to be uninformative, and an uninformative range is an unwelcome
result rather than a wrong one.

The reference class in the documentation only. It costs nothing to build and it
fails in the one situation that matters, which is a number that has travelled
away from its document.

The reference class as free text in the output. Better, and it decays: free
text gets truncated, reflowed into a caption, or dropped by a format that has
nowhere to put it.

The reference class as required structure. It costs every output format the
work of carrying it and it forbids the formats that cannot. That is the cost
being bought rather than avoided.

## Choice

### The event

Three events are computed and never merged into one number. Every output names
which event it is reporting on, in the same structure as the number and not in
a caption.

The deviation tension, from `0014` directly, at threshold `k`.

The pair tension, counted per unordered pair of measurements of one quantity
that are both scorable under `0014`. To stop one bad measurement from
generating a whole series of cases, the pair count is reported alongside the
count of distinct measurements involved and the count of distinct quantities,
and no probability is reported on pairs without all three counts. The pairs are
not independent and the output says so next to the interval rather than in this
record.

The split tension, over method families as coded in the corpus, reported only
for quantities where both families meet a stated minimum count, with that
minimum printed.

`k` is not fixed at three. The headline is reported as a function of `k` over a
stated range, because a single threshold is what turns this into one quotable
number, and one quotable number travelling alone is the failure this whole
record is designed against. Three is where a reader will look and it is a point
on a curve rather than the result.

### Real

Every tension gets exactly one of four outcomes, and the four are stored and
reported separately.

Confirmed. A later measurement admitted under the precision rule in `0014`
falls on the side the tension pointed to, outside the interval of the value the
tension was against.

Refuted. A later admissible measurement agrees with the value the tension was
against, within that value's interval.

Explained. The tension was closed by a published identified error in one of its
inputs, carried through the correction path in
`docs/decisions/0012-where-correction-history-lives.md`. An explained tension is
counted as confirmed or refuted according to which side the identified error
puts it on, and it is also counted separately, because how often a tension was
closed by somebody finding the mistake is a finding in its own right.

Unresolved. No later measurement meeting the admission rule exists. This is
recorded as an outcome and never as an absence.

The probability is reported as three numbers and never as one.

The resolved-only proportion, confirmed over confirmed plus refuted, with a
binomial interval on it, and with the count it rests on.

The censoring bounds, which is the same proportion computed twice over the
whole set: once counting every unresolved case as refuted, and once counting
every unresolved case as confirmed. Those two numbers bracket what the answer
would be under any assumption about the unresolved cases whatsoever, and the
distance between them is the honest size of what is not known.

The unresolved share, printed as a count and a proportion. Where it is large,
the censoring bounds are wide, and a reader should be able to see why in the
same place.

### The reference class

The probability is not a scalar anywhere in this tool. It is a structure whose
required members are the number, its interval, the count of historical cases it
rests on, the event kind, the threshold `k`, the reference choice and vintage
from `0014`, the admission rule, the set of quantities, the set of disciplines,
the period covered, and the unresolved share.

An output format that cannot carry that structure does not carry the
probability. It carries the corpus, the counts and a pointer, and it says which
it is. This is the concrete form of the requirement that no output format can
emit the probability without its conditions, and it is a design rule rather
than a warning in a document, because a warning is what a format author skips
under deadline.

The sensitivity to the reference choice is part of the same structure. The
probability is computed under all four references from `0014` and the four
results travel together. Where they disagree materially, that disagreement is
the result and is reported as one.

## Reasons

Keeping the three events apart is the decision everything else rests on. They
have different base rates, they are what different readers mean by the same
word, and a single number computed on one of them and quoted against another
would be wrong in a way nobody could detect from the number. Merging them would
also require weighting them against each other, and there is no principled
weight.

The censoring bounds are the answer to the part of the question that has no
data. Every alternative supplies a number where the historical record supplies
none. Dropping the unresolved conditions the result on the attention a tension
received, which is precisely the selection this project keeps finding
everywhere else; counting them as not real answers in one direction at whatever
rate our ignorance happens to occur. Bounds refuse to answer and say how much
they are refusing, and if that makes the headline uninformative then the
uninformative headline is the true state of the evidence. The resolved-only
proportion is still reported, because it is the comparable number and it is
what a reader of the prior literature will want, but it is never reported
without the bounds beside it.

Making the probability a structure rather than a number is the only part of
this record that is a mechanism rather than a definition, and it is aimed at
the specific failure the issue behind this record names: a single quotable
number travelling without its conditions. A rule that says the conditions
should accompany the number is broken by the first person who copies the number
into a slide. A representation in which the number does not exist apart from
its conditions makes that copy an act somebody has to perform deliberately.

Reporting the headline as a function of `k` follows from the same reasoning one
step earlier. Three sigma is the number in the reader's question, and answering
only at three sigma invites the answer to become a constant of nature. A curve
cannot be quoted as one number without visibly discarding the rest of it.

Nothing in this repository computes, stores, refuses or checks any of this.
`PROSE, NOT ENFORCEMENT`, and it is the whole of the rule rather than a
residual. There is no source tree, no corpus and no output format. #47 holds the
calculation and the output surface this record constrains, #44 holds the command
line surface it would be printed from, #45 holds the implementation of the
reference choices whose sensitivity is required above, and #46 holds the
distribution the probability is read off. Until those land this record is an
argued position and not a description of any behaviour.

## Date

2026-08-08

## Reversal condition

Revisit the four outcome states if the corpus shows a common case that none of
them fits. The likeliest candidate is a tension that was closed by a change in
the definition of the quantity rather than by a measurement, where neither side
was wrong and the question stopped being the same question.

Revisit the censoring bounds as the headline if the unresolved share turns out
to be small enough that the bounds and the resolved-only proportion say the same
thing. At that point the wide presentation is costing clarity and buying
nothing, and the trigger is a count that the output already prints.

Revisit the structure requirement if it is observed forcing a useful output
format out of existence rather than forcing it to carry the conditions. The
requirement is there to make a bare number hard to produce, not to make the tool
unusable, and if a format that a reader genuinely needs cannot exist under it
then the rule has overshot.
