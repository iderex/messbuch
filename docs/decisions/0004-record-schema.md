# 0004 The measurement record schema

## Question

One record is one published value of one quantity by one experiment. Which
fields does it carry, which of them are required, and what does each one mean?

The schema decides what questions the corpus can ever answer, and it becomes
irreversible the moment there is data in it. A field that is missing cannot be
added retrospectively, because adding it means going back to every source and
reading it again, which is the whole of the work. A field that is present and
optional is worse than absent in a specific way that this record has to guard
against.

This record fixes field names. Several earlier records propose names in
passing and say that this one decides them; where they disagree with what is
written here, this record wins and the other is amended.

## Options considered

A minimal schema: quantity, value, uncertainty, date, source. Anything else
added later when an analysis needs it.

A maximal schema: every field anybody might want, most of them optional, on
the argument that a field nobody fills costs nothing.

A schema derived from the analyses on this board, where every field names the
analysis that needs it and a field no analysis needs is not carried.

A schema adopted from an existing bibliographic or dataset standard and
extended.

## What each option costs

Minimal. Cheapest to specify, cheapest to fill, and the corpus grows fastest.
Its cost is the one that cannot be paid later. Adding the blinding field a
year in means re-reading every source in the corpus, and the people who
transcribed them have moved on. The fields this project would add later are
exactly the ones that answer its own question, so a minimal schema is a corpus
that can hold data and cannot test the thing the data was collected for.

Maximal. Nothing is lost and no re-reading is ever needed. Its cost is the
trap this project has to name explicitly: a field that is optional in practice
is absent in most records, and a field absent in most records cannot be used
in any analysis, because an analysis over the subset that has it is an
analysis over a subset selected by whoever bothered. So a maximal schema does
not buy the questions it looks like it buys, and it buys a real contributor
cost per record plus a validator surface per field.

Derived from the analyses. Every field has a reason that can be checked, the
optional set is small enough to argue about individually, and the schema is
falsifiable: a field whose analysis is dropped is a field to drop. Its cost is
that it binds the corpus to the analyses this board has thought of, and an
analysis nobody has thought of yet is exactly the kind of thing an open corpus
should enable. That cost is real and it is paid.

An existing standard, extended. Interoperability, and somebody else has
already argued about the hard parts. It costs the fit. No standard this
project could adopt models a measurement's uncertainty as a component list
with coverage statements, models the difference between a value read from a
paper and one read from a review, or models the difference between a
measurement and a combined fit, and those three are the corpus's subject. The
extension would be larger than the standard, and the result would be a
standard-shaped thing that is not the standard, which is worse for
interoperability than a clean local schema plus an export.

## Choice

The schema derived from the analyses, with the field list below.

Every field names the analysis that needs it, by issue. A field that names
none is not carried, and the section on refused fields shows the rule biting
rather than being asserted.

Every optional field states why it is optional. The reason has to be a fact
about the sources rather than a judgement about effort, because the trap above
is precisely that effort-optional fields become absent fields.

Anything that will be grouped, filtered or counted is coded against a closed
set. Free text is permitted in exactly one field, it is named below, and no
analysis reads it.

### Absence is written where absence is a fact

Three fields take an explicit value for "the source did not say" rather than
being left out. `uncertainty_status` from
`docs/decisions/0005-uncertainty.md`, `blinding` below, and
`data_taken_status` below.

The pattern is the same each time and it is the most important structural
decision in this record. A field left out because the source said nothing and
a field left out because nobody has read that part yet are the same bytes on
disk, so a corpus that leaves both out cannot tell a gap in the literature
from a gap in its own work. Every count of "how many historical measurements
were blinded" would silently include this project's backlog. Writing the
absence costs one line per record and is what makes those counts mean
anything.

### The fields

Paths are TOML paths inside the file. The file's own path carries the identity
of the record, per `docs/decisions/0003-storage-format.md`, so no field
repeats it.

`schema_version`, required. An integer, the version of this field set that the
record was written against. It means which reading rules apply to this file.
It does not mean the corpus version, which is
`docs/decisions/0011-corpus-versioning.md` and is a property of a release
rather than of a record. Needed by #61, which refuses a schema change that
makes an existing corpus unreadable, and by #28, which requires the built
artifact to be reproducible. This is a structural field and not an analysis
field; see the section on fields kept without one.

`status`, required. One of `active` and `withdrawn`, per
`docs/decisions/0012-where-correction-history-lives.md`. It means whether this
record is part of the corpus as a live statement. It does not mean the record
is correct; a wrong value that has been corrected is `active` with a
correction entry. Needed by #38, which must exclude withdrawn records and
count them, and by #48, whose report states what was excluded.

`superseded_by`, required when `status = "withdrawn"` and a replacement
exists, refused otherwise. The path of the replacing record. It means this
file was withdrawn and that file is what a consumer holding a citation to it
should read. It does not mean a later publication improved on it, which is
`lineage` below and is a different thing entirely. Needed by #48.

`measurement.quantity`, required. An identifier from the controlled
vocabulary, per `docs/decisions/0006-quantity-identity.md`. It means which
vocabulary entry's definition this value is a value of. It does not mean the
technique, which is coded separately, and does not mean the field of study.
Needed by every analysis on the board; #38 groups by it.

`measurement.published.value`, required unless the record is a limit, in which
case it is refused. The value as the source printed it, in the source's own
unit. It means the fact. It does not mean the comparable number, which is the
normalized block. Needed by #38.

`measurement.published.unit`, required. The unit as the source printed it. It
means what the published value is expressed in. Needed by #38 through the
conversion, and by any reader checking the record against the paper.

`measurement.published.uncertainty_status`, required, one of `reported`,
`none-in-source` and `not-transcribed`, per
`docs/decisions/0005-uncertainty.md`. Needed by #38 and #39, which must
exclude records with no usable uncertainty and print the two counts
separately, and by
#45, whose deviation statistic divides by an uncertainty.

`measurement.published.uncertainty`, a list of component blocks, required when
`uncertainty_status = "reported"` and refused otherwise. The component shape,
its `component`, `plus`, `minus`, `coverage`, `interval_kind`,
`correlation_group` and `rescaling` fields, is fixed by
`docs/decisions/0005-uncertainty.md` and is not restated here. Needed by #38,
#39, #40 and #45.

`measurement.published.limit`, required when `measurement.published.value` is
absent, refused otherwise. Its `direction`, `bound`, `coverage` and
`interval_kind` fields are fixed by `docs/decisions/0005-uncertainty.md`.
Means: the source reported a bound rather than a value. It does not mean the
measurement failed. Needed by #42, where a limit is part of a quantity's
history, and by the exclusion accounting in #38 and #48.

`measurement.normalized`, required unless `normalization_status` says the
record is not convertible, refused when it does. The same value and
uncertainty components expressed in the canonical unit, per
`docs/decisions/0007-units.md`. It means the comparable number. It does not
mean a better number; where they differ in precision the published one is the
fact. Needed by #38, which cannot pool values in different units.

`measurement.conversion`, required whenever `measurement.normalized` is
present. Its `factor`, `exact`, `imported_uncertainty` and source-of-factor
fields are fixed by `docs/decisions/0007-units.md`. It means the stored
multiplication relating the two blocks, so that the arithmetic is checkable
from the record alone. Needed by #38 and by the validator's meaning leg, #25.

`measurement.normalization_status`, required. One of `normalized` and
`not-convertible-across-redefinition`, per `docs/decisions/0007-units.md`. It
means whether a comparable number exists for this record. Needed by #38 and
#42 for the exclusion accounting.

`measurement.definition_epoch`, required. A string naming the definitional
regime the source published under, from the closed set the vocabulary entry
maintains, per `docs/decisions/0007-units.md`. It means which definition of
the quantity the number is a number of. It does not mean the publication date;
a paper can publish under a superseded regime. Needed by #42, whose
cumulative-by-year analysis must not pool across a regime boundary.

`measurement.denominator_quantity`, required for a dimensionless quantity,
refused otherwise. Either a vocabulary identifier or the literal
`definitional`, per `docs/decisions/0007-units.md`. It means what the value is
expressed relative to. Needed by #38.

`publication.date`, required. The publication date of the source, as precise
as the source states it, and no more precise. It means when this value entered
the literature. It does not mean when the measurement was made. Needed by
#42, and by #45 and #46, whose whole subject is how a series moves in
publication order.

`publication.submitted`, optional. The date the source states it was received
or submitted. Optional because many sources do not print one, which is a fact
about the literature and not about transcriber effort. It means the last date
by which the authors' own analysis was fixed. It does not mean when the data
were taken. Needed by #45 and #46: what an author could have seen before
committing to a number is the mechanism the bandwagon question is about, and
the publication date overstates it by however long the journal took.

`publication.data_taken_status`, required. One of `reported`, `none-in-source`
and `not-transcribed`, with the same meaning the uncertainty status field has.
Needed by #46, which has to distinguish a literature that does not say when
its data were taken from a corpus that has not finished reading.

`publication.data_taken`, required when `data_taken_status = "reported"`,
refused otherwise. A start date and an end date, either of which may be a year
alone where that is all the source gives. It means when the measurement was
taken. It does not mean the analysis period. Needed by #46: the bandwagon
question is about the order values were published in and the physics is about
the order they were measured in, and a corpus that cannot separate the two
cannot tell a drift in the apparatus from a drift in the community.

`method.technique`, required. An identifier from the `techniques` set in the
quantity's own vocabulary entry, per
`docs/decisions/0006-quantity-identity.md`. It means how this value was
obtained. It does not mean the apparatus or the laboratory. Needed by #46,
which has to be able to split the pooled deviation distribution by method, and
by #40, where a funnel that separates by technique shows a different thing
from one that does not.

`method.note`, optional, and the one free-text field in the schema. Means:
what the transcriber needs a later reader to know that no coded field carries,
for example which theoretical input an extraction assumed. Does not mean
anything to any analysis: NO ANALYSIS READS THIS FIELD, and nothing may be
grouped, filtered or counted by it. Optional because most records need nothing
said. It is carried without an analysis and the section below says why.

`group.id`, required. A slug under the identifier syntax of
`docs/decisions/0006-quantity-identity.md`, naming the collaboration, group or
laboratory that produced the measurement, and resolving to an entry in
`group/<group-id>.toml`. It means who made this measurement, at the
granularity at which two measurements share systematics and share a prior. It
does not mean the institution's legal identity and does not mean the author
list. Needed by #38: a group publishing four times is not four independent
measurements, and pooling that treats it as four reports an interval that is
too narrow. Also needed by
#41, since a trim-and-fill run over a corpus with hidden within-group
correlation is measuring the wrong thing.

`lineage.supersedes`, optional, a list of record paths. It means the authors
present this value as replacing those earlier values of theirs. Optional
because most measurements supersede nothing, which is a fact about the
literature. Needed by #38, which selects one value per group where a group has
published several.

`lineage.superseded_by`, optional, a record path. The mirror of the above. It
means a later publication by the same group presents its value as replacing
this one. It does not mean this record is withdrawn from the corpus, which is
`status` and `superseded_by` at the top level; the two are separate fields
with separate meanings and the paths differ so that neither can be read as the
other. Optional for the same reason. Needed by #38 and #42.

`blinding`, required. One of `blinded`, `not-blinded` and `not-stated`. It
means whether the source says its analysis was blinded to the expected answer.
It does not mean the analysis was or was not blinded in fact; `not-stated` is
a statement about the paper. Needed by #46 and #45: whether the deviation
distribution differs between blinded and unblinded analyses is one of the
sharpest tests the corpus can run on its own central claim, and it is only
runnable if the field is on every record rather than on the ones somebody
remembered.

`source`, required, with the identifier, print, locator, `statement_kind`,
`directness`, `via` and `confirmation` fields fixed by
`docs/decisions/0008-provenance.md` and not restated here. Needed by #48,
which prints the confirmation breakdown, and by #38, which may not pool a
`combined-fit` value as an independent measurement.

`correction`, optional, a list of entries whose shape is fixed by
`docs/decisions/0012-where-correction-history-lives.md`. Optional because most
records have never been corrected. Needed by #48.

### Fields kept without an analysis, and why

The rule this record works under is that a field nobody can attach an analysis
to is dropped. Two fields are kept anyway and this section is the disclosure
rather than an exception hidden in the list.

`schema_version` and `status` are structural. They are read by the validator
and by the artifact build rather than by an estimator, and the issues that
need them are #61, #28 and #48. Calling those analyses would be stretching the
word, so they are named here instead.

`method.note` has no reader at all, machine or otherwise, beyond a person
checking a transcription. It is carried because
`docs/decisions/0003-storage-format.md` chose TOML partly so that a
transcriber could write down what the source actually said, and moving that
into a comment would put it outside the built artifact where no consumer ever
sees it. The risk is exactly the trap this record names: a free-text field is
where structure goes to die. What holds it is that nothing may read it. The
moment an analysis wants to group by something in this field, the answer is a
coded field and not a regular expression over notes.

### Fields proposed and refused

Each of these was on the list and none of them survived the rule. They are
written down so that the next person to propose one finds the argument rather
than having it again.

The value exactly as printed, as a string, for example the parenthesised-digit
form. It would let a reader check the transcription's formatting and would
preserve the significant digits the source chose. No analysis on this board
needs it. A digit-preference or rounding diagnostic would, and no issue here
asks for one, so the field is dropped rather than carried against a future
that may not arrive. If such an analysis is opened, this is the field it needs
and this paragraph is where it is already argued.

The author list. It is already in the source's own bibliographic data and the
first author is already in the file name, per
`docs/decisions/0003-storage-format.md`. No analysis groups by author; the
independence question groups by `group.id`, which is the granularity at which
systematics are shared.

The country or the institution, separately from the group. No analysis on this
board asks whether the drift differs by laboratory or by country. It is a
reasonable question and it is not on this board, so the field is not carried.

A quality score, a reliability rating or a weight assigned by this project. It
would be this project's judgement stored as though it were data, in a corpus
whose entire argument is that judgements dressed as measurements are the
problem. Refused, and refused permanently rather than dropped for now.

Free-text keywords or tags. Refused by the coding rule: anything that will be
grouped is coded, and a tag field exists to be grouped.

### The size of the required set, said plainly

A record has eighteen or so required fields depending on which of the
conditional ones apply, and that is a lot to fill in for one number out of one
paper. The cost is real and falls on the contributor, at the moment after they
have already done the reading, which is when people stop. Nothing in this
record reduces it. What can reduce it is tooling that fills what can be
derived and prompts for the rest, and no issue on this board owes that today.

## Reasons

The derived schema wins because it is the only one of the four whose fields
can be argued about individually. Under the minimal schema the argument is
about whether to add a field, which is always answered later; under the
maximal one there is no argument at all, and the corpus fills up with fields
that are present and unusable. Naming the analysis per field turns each one
into a claim that can be checked and, more usefully, falsified: drop the
analysis and the field goes with it.

The cost of binding the schema to the analyses this board has thought of is
accepted rather than argued away. The mitigation is not a set of speculative
fields, it is that the corpus stores the published fact and its provenance
precisely enough that somebody can go back to the source. A question nobody
has asked yet is answered by re-reading sources, and re-reading is possible
only because `docs/decisions/0008-provenance.md` requires a locator precise
enough to find the number again. That is where the future-proofing lives, and
it is a better place for it than a column of empty optional fields.

Explicit absence values are the structural choice this record would defend
first. They cost one line per record and they are the only thing standing
between this project and a class of statement that would be quietly false:
every sentence of the form "N per cent of historical measurements did X"
computed over a corpus that cannot distinguish the literature's silence from
its own incompleteness.

`blinding` being required rather than optional is the sharpest case of the
same argument and is worth stating on its own. Blinding is the strongest
single predictor anybody would look for in a study of whether published
intervals mean what they claim. A blinding field filled in on the records
where somebody noticed it is worse than no field, because it produces a
comparison between blinded studies and studies nobody checked, and that
comparison looks like the real one.

The single free-text field is a deliberate hole with a rule around it rather
than a compromise. Trap two in the issue that asks for this record says free
text is where structure goes to die, and the way it dies is that something
starts reading it. Keeping exactly one such field and forbidding every reader
keeps the transcriber's note where a reader can see it and keeps it out of
every count.

Nothing in this repository refuses any of this. `PROSE, NOT ENFORCEMENT`.
There is no schema file, no validator, no vocabulary loader and no build here
today, so a record missing every required field above, or carrying a field
this record refuses, passes every route in this repository. The
machine-readable form of this field set is owed by #23, the structural
refusals by #24, the meaning refusals by #25, and the fixture behind each
refusal by #26. The `group/` registry that `group.id` resolves against does
not exist and nothing on this board owed it before this record; an issue is
opened for it on the seed-corpus milestone. Until all of that lands, the field
list above is a description of a program nobody has written, checked by a
reviewer holding this document.

## Date

2026-08-08

## Reversal condition

Reverse a field the day its analysis is dropped. That is the rule this record
is built on and it runs in both directions: if #46 is closed as not-doing,
then `publication.data_taken`, `publication.data_taken_status` and `blinding`
have lost their reason and the schema should shrink rather than carry them out
of habit.

Reverse the optionality of `publication.submitted` if the share of records
carrying it turns out to be high enough that the field is usable, since the
only reason it is optional is that many sources do not print one. That is a
count the validator can produce over the corpus, and it should be counted
rather than assumed in either direction.

Reverse the single free-text field if anything is observed reading it. At that
point the field has become a coded field with no code set, and the repair is
the coded field, not a stricter note format.

Revisit the whole field set the first time a source cannot be transcribed
without inventing a field. That is the expected way this record changes, and a
new field is the correct response rather than forcing the source into an
existing one, which is the same rule `docs/decisions/0005-uncertainty.md` sets
for its own shapes.

Revisit `schema_version` if #61 concludes that a per-record version is the
wrong place for the compatibility statement. This record carries it because a
corpus is read record by record, and if that turns out to be false the field
has no reader.
