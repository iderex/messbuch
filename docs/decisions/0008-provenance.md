# 0008 Provenance, and what makes a record checkable

## Question

A number in this corpus is worth what its provenance is worth. What does a
record have to carry so that a reader can open the source it names and see the
same number without hunting, so that a value read out of a review is
distinguishable from one read out of the paper, and so that a transcription
nobody has checked twice is distinguishable from one that has? And what does
the validator do about a source it cannot resolve, given that the validator
runs in a gate with no network?

## Options considered

A source identifier and nothing else, on the argument that the identifier
resolves to the paper and the paper contains the number.

An identifier plus a free-text citation, so a human can find the value.

An identifier plus a structured locator inside the source, plus a statement of
what kind of claim the source was making at that locator, plus a record of
whether the value was read from the source or from something citing it, plus a
confirmation state.

The above, with unconfirmed records refused entry until a second reader has
agreed.

## What each option costs

Identifier only. One field, and nothing to argue about. It costs the property
the corpus is for. A paper is thirty pages and a number appears in an
abstract, in a results table, in a figure caption, in a combined fit and in a
conclusion, often as five different numbers. A reader handed a resolvable
identifier and a value has to find the value before they can check it, and
finding it is the expensive half. Worse, when they find a number that differs
they cannot tell whether the transcription is wrong or whether they are
looking at a different statement in the same paper, so the check ends
inconclusively, which is the outcome that trains people to stop checking.

Identifier plus free-text citation. Cheap to write and genuinely helpful to a
human. It costs everything at analysis time and most things at validation
time. Free text cannot be grouped, cannot be counted and cannot be refused, so
the question "how many records in this corpus were read out of a review rather
than out of the paper" has no answer, and that question is one of the ways
this corpus checks itself.

The structured option. It carries the case the literature actually produces
and every field on it is a field some analysis or some check reads. It costs
the contributor real effort per record, at the exact moment they are least
willing to spend it, which is after they have already found the number. It
also costs the schema and the validator more surface, and every closed
vocabulary in it is a thing that will meet a case it does not have a value
for.

Refusing unconfirmed records. It makes the corpus's confirmation state
trivially true, since every record in it is confirmed. It costs the corpus its
growth rate, because entry now runs at the speed of the second reader rather
than the first, and it costs the project its own visibility: work that has
been done once and not twice stops existing anywhere, so the backlog is
invisible and its size is unknown. A project that cannot see its own backlog
cannot decide whether to recruit a second reader, which is the decision the
backlog exists to inform.

## Choice

The structured option, with unconfirmed records admitted and marked.

### The source, and how it is identified

Every record carries a `[source]` block. The source is identified by at least
one entry in a closed set of schemes: a digital object identifier, an arXiv
identifier, a bibliographic-service code, an ISBN for a book, and a URL for
material that has none of the above and lives at a stable address.

Identifiers are stored bare. A digital object identifier is stored as the name
itself and not as a resolver URL, and the same holds for every other scheme
that has a canonical bare form. A resolver prefix is a routing decision that
belongs to whoever is doing the resolving, and storing it bakes today's
resolver into every record in the corpus.

This record does not restate the syntax of any external scheme. The validator
checks each identifier against its scheme's own published syntax, and the
per-scheme shape lives with the checker rather than here, because a syntax
restated in a document drifts against the code that decides it and the
document is then the thing people trust.

A source with no identifier in any of those schemes is the ordinary case for
the older literature, and it is not refused. Such a record carries a
`[source.print]` block with the bibliographic fields needed to find the item
in a library: journal or publisher, volume, issue, page range, year, and the
author list as published. It also carries `resolvable = false`, which is
written rather than inferred from the absence of an identifier, so a record
whose transcriber had no identifier and a record whose transcriber did not
look are different states.

### The locator

`locator` is required on every record and says where inside the source the
value was read. It is structured rather than free text, and it carries
whichever of these apply, with at least one always present:

- `page`, the page of the source as printed, not the page of a PDF.
- `table`, the table's own label as the source gives it.
- `equation`, the equation's own number as the source gives it.
- `figure`, the figure's own label, used when the value was read from a figure
  or its caption.
- `section`, the section's own number or title.

The labels are the source's, not ours. A table numbered `III` is stored as
`III`, because a reader is going to look for `III` on the page.

`page` is separated from a PDF page count deliberately. The two differ for
almost every scanned journal article, and a locator that sends a reader to the
wrong page is worse than one that sends them to no page, because they will
believe they are in the right place.

### What kind of statement the source was making

`statement_kind` is required and takes exactly one value from a closed set.

- `primary-result`. The value the source reports as its own measurement, in its
  results section or its results table. This is the default case and the one an
  analysis wants.
- `abstract`. Read from the abstract. Abstracts round, and an abstract value and
  a results-table value from one paper are frequently different numbers.
- `summary-table`. Read from a table in which the source collects its own result
  alongside other people's. The value is the source's, but the table's rounding
  and unit are the table's.
- `combined-fit`. A value the source produced by combining its measurement with
  measurements from elsewhere. This is not an independent measurement and no
  pooling analysis may treat it as one.
- `figure`. Read off a figure rather than from a number the source printed. The
  precision is the reader's, not the source's.
- `erratum`. From a correction or corrigendum to the source.

`combined-fit` and `figure` are the two that change what an analysis may do
rather than only what a reader should expect. A combined fit double-counts
every input it contains, and a value read off a figure carries a reading error
nobody quantified. Both are stored rather than refused, because both are real
things the literature contains and a corpus that silently dropped them would
misreport what the literature holds.

### Read from the source, or from something citing it

`directness` is required and takes one of two values.

- `primary`. Somebody opened the source named above and read the value out of
  it.
- `secondary`. The value was taken from a review, a compilation or a database
  that cited the source. Nobody on this project has opened the source.

A `secondary` record carries a `[source.via]` block naming what was actually
read: its own identifier under the same scheme rules, and its own locator
under the same locator rules. Both are required. A secondary record without
them is a record claiming a chain of custody it cannot show, which is worse
than one claiming none.

The claim a secondary record makes is weaker and it is weaker in a specific
way. A review's own transcription is a step this project did not check, and
reviews contain transcription errors for the same reason everything else does.
So a secondary record is a record of what the review said, and it is never
evidence about what the paper said.

### The confirmation state

`confirmation` is required and takes exactly one of three values.

- `unconfirmed`. One person has read the value out of the thing named. This is
  where every record starts.
- `confirmed`. A second person independently read the same locator in the same
  source and got the same value. Independently means without looking at the
  first reading first, which is a thing the process asks for and nothing can
  check.
- `disputed`. A second reading disagreed and the disagreement is not yet
  resolved. The record stays in the corpus, is excluded from every analysis, and
  the disagreement is on the correction path in
  `docs/decisions/0012-where-correction-history-lives.md`.

A `confirmed` record carries `confirmed_by`, holding the confirming reader's
contributor identity in the form the version history already holds it. That
field adds no personal data the repository's history does not already carry,
which is the whole of the claim: it is not a promise that contributing here is
anonymous, and `docs/decisions/0009-offline-by-default.md` is where that is
said plainly.

`confirmation` and `directness` are orthogonal and neither substitutes for the
other. Two people reading the same review and agreeing is a `confirmed`
`secondary` record, and it remains a statement about the review. Reading the
paper as a second check on a review's number is not a confirmation of the
secondary record, it is a new `primary` record, and the two then either agree
or produce a correction.

### Unconfirmed records are admitted, and every analysis prints the breakdown

Unconfirmed records enter the corpus and are included in analyses by default.
Every analysis prints the confirmation breakdown of the set it used: how many
`confirmed`, how many `unconfirmed`, and how many `disputed` records were
excluded. An analysis that cannot print the breakdown does not run.

`--confirmed-only` restricts a run to `confirmed` records, and when it is used
it appears in the output stamp described in
`docs/decisions/0011-corpus-versioning.md`, along with the count it removed.

This is deliberately weaker than the exclusion-by-default rule that
`docs/decisions/0005-uncertainty.md` applies to an unstated coverage
convention, and the asymmetry is not an inconsistency. An unstated coverage
that an analysis fills in is a bias of known size in a known direction, so
including such a record moves the answer predictably and excluding it is a
correction. A transcription that has not been read twice is an error of
unknown size in no particular direction, so excluding unconfirmed records does
not correct anything and only removes data. What it would do is hide the size
of the unchecked share inside a number that looks clean, and printing the
breakdown is the response that actually addresses the risk.

### What the validator does about a source it cannot resolve

The validator cannot resolve anything. It runs in the gate, the gate has no
network by the rule in `docs/decisions/0010-headless-tests.md`, and reaching a
bibliographic service from the validator would put a network-capable import
into a package that `docs/decisions/0009-offline-by-default.md` forbids it in.
So resolvability is split into two questions answered in two places, and the
split is the rule.

In the gate, the validator refuses a record on shape alone:

- No identifier in any scheme and no `[source.print]` block, or `resolvable`
  absent when there is no identifier. A source that can be neither resolved nor
  looked up in a library is not a source.
- An identifier that does not match its scheme's syntax.
- A `locator` block with none of its fields present.
- `directness = "secondary"` with no `[source.via]` block, or a `via` block
  missing its identifier or its locator.
- `confirmation = "confirmed"` with no `confirmed_by`.
- A `statement_kind` or a `confirmation` value outside its closed set.

A record that passes those checks is a well-formed record. It is not a record
whose source exists, and the validator makes no claim that it is.

Outside the gate, the online harness named in
`docs/decisions/0010-headless-tests.md` resolves the identifiers and reports
which ones did not resolve. Its result is not a merge gate and never becomes
one, because a merge that depends on a third-party service being up is a merge
that stops working when that service does. An identifier that fails to resolve
there is a defect report against the record, and it travels the correction
path like any other wrong value. A resolution failure never deletes a record
and never silently rewrites an identifier.

The one thing that must not happen is the harness's silence being read as
every identifier resolving. That is the same failure the harness's own
reporting rule in `docs/decisions/0010-headless-tests.md` exists against, and
it applies here in its sharpest form, because the claim at stake is the
corpus's central one.

## Reasons

The structured locator wins because the alternative fails at the only moment
that matters. A corpus like this is checked by somebody who is sceptical of
one number, and everything about a record should be aimed at making that
person's check cheap and conclusive. An identifier alone makes it expensive;
free text makes it cheap for that one person and impossible for every count
the project wants to run over the corpus as a whole.

`statement_kind` is the field most likely to be seen as bureaucracy and it is
the one that stops a specific, silent, fatal error. A combined fit pooled
alongside the measurements it already contains counts those measurements twice
and reports a tighter interval than any of the inputs, which is precisely the
false precision this corpus exists to detect. Without the field, a combined
fit and a primary result are the same row, and no analysis can tell them apart
afterwards.

Admitting unconfirmed records is the cost of the chosen option and the reason
is about seeing the project's own state rather than about the numbers. Refusal
would give a corpus in which every record is confirmed and nobody knows what
was left outside it. Admission gives a corpus whose unchecked share is a
number that appears in every result, and a number that appears in every result
is a number somebody eventually acts on.

Splitting resolvability across the gate and the harness is forced rather than
chosen. Every alternative either puts a network call inside the validator,
which two other records forbid for reasons that outrank this one, or drops the
resolvability check entirely, which would leave the corpus's most important
claim untested. Naming the split, and naming which half is a merge gate, is
what keeps the weaker half from being mistaken for the stronger one.

Nothing in this repository refuses any of this. `PROSE, NOT ENFORCEMENT`.
The source, locator, statement-kind, directness and confirmation shapes above
are written into `schema/record-1.toml` and nothing reads that file. There is no
validator. The harness that would resolve an identifier exists, at `test/online/`
behind the `online` build constraint, and it registers no test that resolves
anything. Every refusal listed under the validator above is a
refusal that is owed: the structural half is #24, the meaning half is #25, and
the fixture behind each refusal is #26. Until they land, a record naming no
source at all passes every route in this repository, and the field list above
describes a program nobody has written.

## Date

2026-08-08

## Reversal condition

Reverse the admission of unconfirmed records if the confirmed share of the
corpus stops moving while the corpus grows. The measure is the share rather
than the count, it is a number the validator can print, and the trigger is the
share falling over a period in which records were added. At that point
admission is not buying growth in a checkable corpus, it is buying growth in
an unchecked one, and refusing entry becomes the cheaper of two bad options.

Reverse the closed set of `statement_kind` values the first time a case
appears that none of them describes and that changes what an analysis may do
with the record. A case that only changes what a reader should expect is a
documentation matter and does not reverse this.

Revisit the bare-identifier storage rule if a scheme this corpus needs has no
canonical bare form, so that storing it without a resolver is genuinely
ambiguous. Today every scheme in the set has one.

Revisit the split between the gate and the harness if the project ever
acquires a locally held bibliographic index, because resolution against a
local file is not a network call and could then move into the gate. That would
be a change of what is possible, not a change of what is wanted, and the
wanted state has always been resolution inside the gate.
