# Adding a measurement to the corpus

This is the guide to read before adding your first measurement. It is written
for a careful person who is not a programmer, and it assumes you can read the
paper you are transcribing from and nothing else.

One record is one published number, from one source, in one file. Most of the
work is reading and judgement rather than typing, and the judgements are what
this guide is about.

It does not list the fields of a record. That list lives in
`docs/decisions/0004-record-schema.md`, and a second copy here would drift
against it, which is the failure that hurts most in a guide people follow
literally. Where this guide names a field it is because the judgement is about
that field. Where you need the full list, go to the record.

What this guide cannot yet promise: that following it produces a record a
machine accepts. There is no validator in this repository today.

    git ls-files | grep -E '\.go$|go\.mod'
    (no output, exit 1)

It is owed by #24 and #25. Until it exists, everything below is checked by a
person reading your pull request, and a record that follows this guide and still
turns out wrong is a defect in the guide worth reporting.

## Choosing the quantity identifier

Look in `vocabulary/` first. Each file there is one quantity, and its
`definition` field is written to be precise enough to decide whether the number
in front of you is a measurement of that quantity or of something next to it.
Read the definition rather than the identifier. Identifiers are short and two
neighbouring quantities often have names that sound the same.

When two names might be the same quantity, one rule decides it, and it is in
`docs/decisions/0006-quantity-identity.md`: a new identifier is warranted when
two candidate values could differ for a reason that is not measurement error.
If the true value is the same number under both names, they are one quantity and
the second name is an alias.

Read that rule in the direction that is hard. The tempting move is to split a
quantity whenever two sets of measurements disagree. That move is always
available and always looks like caution, and it destroys the finding: a
disagreement between two ways of measuring one quantity is the most interesting
thing this corpus can hold, and splitting it produces two series that each agree
with themselves and no tension at all.

Two things follow from that and both are easy to get backwards. A measurement
technique is not a quantity, so two techniques aimed at the same defined thing
share one identifier and the technique goes in the record's own coded method
field. A change in the defining convention is not an alias either; the
vocabulary entry carries the regimes and conversion across them is refused.

If no entry fits, stop rather than inventing an identifier in your record. An
identifier that resolves to no vocabulary entry is a coded field with no code
set, which is exactly what the schema record refuses. Open an issue proposing the
entry, with the definition you would write and the neighbouring quantity it
excludes. The entry and the first record that needs it can land together.

## Which value to take when the paper gives several

This is the common case rather than the exception, and it is where corpora go
wrong.

Take the value the source reports as its own result, in its results section or
its results table. Then say which one you took, because the record carries that:
`statement_kind` in `docs/decisions/0008-provenance.md` distinguishes a primary
result from an abstract, a summary table, a combined fit, a figure and an
erratum, and the distinctions are not bookkeeping.

The abstract is the trap. Abstracts round, and the abstract value and the
results-table value from one paper are frequently different numbers. If you read
it from the abstract, say so; do not quietly upgrade it.

A combined fit is not a measurement. Where the source has folded its own result
together with other people's, the number that comes out contains those other
people, and no pooling analysis may treat it as independent. It is still worth
recording, and it is recorded as what it is.

A value read off a figure carries a reading error nobody quantified, including
you. Record it as read off a figure.

One record is one value. If the paper reports two measurements it presents as
its own, that is two records. If the paper reports one measurement and shows the
intermediate steps, that is one record. Never average two of a paper's numbers
yourself, and never take a number that only exists after arithmetic you did.

## Uncertainty stated in an unusual way

Nothing is combined on the way in. If the source gives a statistical and a
systematic term, both go in as separate components with their own intervals. Do
not add them, in quadrature or otherwise. Combining is an analysis step and it
happens later, where it can be argued with.

Absence, zero and a limit are three different states and they are stored as three
different things. `docs/decisions/0005-uncertainty.md` writes out seven cases in
full, including the asymmetric interval, the total-only quote, the confidence
level given as a percentage, and the upper limit. Work from those rather than
from this paragraph; they are written in the representation you will be typing.

One distinction there is the one most often collapsed and it is worth stating
twice. A source that printed no uncertainty at all and a source whose
uncertainty you have not transcribed yet are the same bytes on disk apart from
one field. `uncertainty_status` is what separates them, and getting it wrong
turns a gap in the literature into a gap in our work or the other way round.
If you have not read the uncertainty out yet, say that. It is an honest state and
the corpus is built to hold it.

If the source states an interval whose meaning you cannot pin down, that is the
unclear-paper case below rather than a reason to guess a coverage.

## Units, and the conversion you must not do by hand

The published value goes in exactly as the source printed it, in the unit the
source used. That block is the fact and nothing in the corpus may rewrite it.

The normalized value in the canonical unit is stored alongside it, together with
the conversion factor and where the factor came from, so that a reader can divide
one by the other and recover what the source printed.
`docs/decisions/0007-units.md` fixes the canonical unit per dimension and the
shape of that arithmetic.

Do not convert a value into modern units by hand and record only the result. A
converted number with no factor beside it cannot be checked against the paper by
anybody, ever, and the paper is the only thing that settles a transcription
dispute.

Two conversion cases need a decision from you. A conversion fixed by definition
imports no uncertainty and is marked exact. A conversion running through a
measured constant imports the constant's uncertainty, and the normalized block
carries that as its own component. Where the imported term is small enough to
make no observable difference, it is recorded as negligible rather than dropped,
with the two numbers that produced the judgement, because a silently omitted
negligible term and a silently omitted large one look identical on disk. The
threshold and the arithmetic behind it are in the units record.

## Where you read it: the locator

`locator` says where inside the source the number is, and a reader will use it to
find the number again. It is structured rather than free text and carries
whichever of page, table, equation, figure and section apply, with at least one
always present.

Use the source's own labels. A table the paper numbers `III` is recorded as
`III`, because that is what a reader will be looking for on the page.

Use the printed page, not the page of the PDF. They differ for almost every
scanned journal article, and a locator that sends a reader to the wrong page is
worse than one that sends them nowhere, because they will believe they are in the
right place and conclude the transcription is wrong.

## The source, and whether you opened it

The source is identified by a bare identifier under one of the schemes in
`docs/decisions/0008-provenance.md`. Bare means the identifier itself, not a
resolver URL wrapped around it.

Older material often has no identifier in any scheme. That is the ordinary case,
not a problem: the record carries a print block with what a librarian needs, and
it states that the source is not resolvable rather than leaving that to be
inferred from an empty field. A transcriber who had no identifier and a
transcriber who did not look are different states and the corpus keeps them
apart.

Then the question this guide exists to make unavoidable: did you open the source?

If you read the value out of the source itself, the record is primary. If you
took it from a review, a compilation or a database that cited the source, the
record is secondary and it names what you actually read, with its own identifier
and its own locator. Both are required, because a secondary record without them
claims a chain of custody it cannot show.

A secondary record is weaker in a specific way rather than in a vague one. It is
a record of what the review said. It is never evidence about what the paper said,
because the review's own transcription is a step nobody here checked, and reviews
contain transcription errors for the same reason everything else does.

Do not include a value from a source you have not opened and cannot name the
route to.

## The group

Every record names the group that made the measurement, and that name has to
resolve to an entry in `group/`. The granularity is the one at which two
measurements share systematics and share a prior, which is not the same as an
institution and not the same as an author list. `group/README.md` is where the
shape and the choosing rule are written.

If no entry exists yet, it lands in the same pull request as the record that
first needs it. That is written into `group/README.md` and it is not optional.

If the source does not settle which group a measurement belongs to, argue it in
an issue rather than guessing. A guessed group is invisible afterwards and it
biases exactly the pooling the corpus is built to do.

## What a second reading is

Every record starts unconfirmed. One person has read the number out of the thing
named, and that is all the corpus claims.

A confirmation is a second person independently reading the same locator in the
same source and getting the same value. Independently means without looking at
the first reading first. Nothing can check that, which is why it is written here
rather than left to be assumed: the value of the second reading is destroyed the
moment it becomes a check that the first reading was copied correctly.

Where the second reading disagrees and the disagreement is not resolved, the
record is marked disputed. It stays in the corpus, it is excluded from analyses,
and the disagreement goes down the path in `docs/corrections.md`.

Confirmation and directness are different questions and neither substitutes for
the other. Two people reading the same review and agreeing gives a confirmed
secondary record, and it is still a statement about the review. Opening the paper
to check a review's number is not a confirmation of that record at all; it is a
new primary record, and the two then either agree or produce a correction.

## When the paper is unclear

Record the ambiguity. Do not resolve it silently.

The schema has explicit values for what the source did not say, and
`docs/decisions/0004-record-schema.md` fixes which fields take them. The pattern
is always the same: a field left out because the source said nothing and a field
left out because nobody has read that part yet are the same bytes on disk, so the
absence is written down instead of implied.

Where the ambiguity is real and no field can carry it, the record waits. Put the
question in an issue with the passage quoted, and let it be answered before the
number enters the corpus. A guessed reading is indistinguishable from a read one
afterwards, and the corpus has no way to find it again.

## The four things not to do

Do not convert a value into modern units by hand and record only the result.

Do not fill in a missing uncertainty from a later paper. A later paper's
uncertainty is a different measurement's uncertainty, and the older record's
missing one is data about how that literature was published.

Do not include a value from a source you have not opened, unless the record says
plainly that you did not open it and names what you read instead.

Do not guess a group. The registry entry is cheap to add and a wrong group is
expensive to find.

## After you have written it

One record per file, in the layout and under the filename rule in
`docs/decisions/0003-storage-format.md`. `record/_example/1900-example-01.toml`
is the committed illustration of that layout, and its numbers are invented and
may not be cited.

Open a pull request. The template asks two questions and both want a sentence
rather than a tick. The second one asks which documents your change makes wrong,
and `docs/downstream-documents.md` is the map to answer it from.

If a number already in the corpus turns out to be wrong, that has its own path
and it is not this guide: `docs/corrections.md`, and the report form under
`.github/ISSUE_TEMPLATE/`.

## What checks this

Nothing.

No check in this repository reads this guide, compares it against the decision
records it summarises, or refuses a record that ignored it. The validator that
would refuse a malformed record does not exist yet and is owed by #24 and #25;
the meaning checks that would catch a plausible but wrong record are #25's.
Until then a record is checked by whoever reads the pull request.

The guide is a summary of decisions taken elsewhere, and where it disagrees with
a decision record, the record is right and this file is a defect. It names the
records rather than restating them wherever the detail matters, for that reason.
