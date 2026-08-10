# 0006 Quantity identity and the controlled vocabulary

## Question

What names a quantity in this corpus, what does a vocabulary entry have to say
so that somebody can decide whether a given published number belongs under it,
and when is a new identifier warranted rather than another name for an
existing one?

Two failures decide this and both are silent. Two contributors ending up with
two identifiers for one quantity splits a series in half, and every analysis
then runs on half the data without saying so. One identifier covering two
things that are not the same quantity pools measurements of different objects
and reports the difference between them as a tension. The first loses a
signal, the second manufactures one, and this corpus exists to tell those
apart.

## Options considered

Opaque identifiers, allocated in sequence, with the human-readable name
carried only as a field.

Descriptive slugs in a single flat namespace.

Hierarchical identifiers, so that the discipline and the sub-field are part of
the name.

Adopt an external scheme's identifiers as this corpus's own.

## What each option costs

Opaque identifiers. Nothing about the identifier can go stale, since it says
nothing, and renaming a quantity is free because the name is only a field. The
cost is paid on every read. A directory listing under `record/<quantity-id>/`,
fixed by `docs/decisions/0003-storage-format.md`, becomes unreadable, every
diff in a contribution shows a number instead of a subject, and a contributor
cannot tell whether an entry for their quantity already exists without a
lookup. That last one is the failure this record is about: the cost of
checking is what decides whether people check, and a scheme that makes
checking expensive produces duplicate identifiers.

Descriptive slugs, flat. Readable everywhere it matters, and the check for
whether an entry exists is reading a directory. The costs are two. A slug
looks like a definition and is not, so somebody will file a record under a
slug whose words match their quantity without opening the entry, which is
exactly the second failure above. And a slug chosen badly is stuck, because
the identifier is a directory name that every existing record's path contains,
so renaming is a corpus-wide move.

Hierarchical identifiers. A reader sees where a quantity sits, and related
quantities sort together. The cost is that it requires a taxonomy of the
sciences that is agreed before the first record is filed. This corpus is
cross-disciplinary by design, the boundaries between disciplines are contested
by the people inside them, and a quantity that sits in two places has to be
put in one. Every hierarchy also invites the question of where a new branch
attaches, which is a discussion with no evidence in it.

An external scheme. The definitions come with authority, the boundaries have
been argued by people who know the field, and cross-referencing is free. The
costs are decisive. No single external scheme covers the range this corpus
claims, so adopting one means adopting one per discipline and the corpus has
no identifier space of its own. External schemes also renumber, retire and
merge entries on their own schedule, and this corpus's identifiers are paths
in a version-controlled tree that must not move.

## Choice

Descriptive slugs in a single flat namespace, with a vocabulary entry per
identifier, and external schemes referenced rather than adopted.

### Identifier syntax

An identifier is one to sixty characters of lower-case ASCII letters, digits
and hyphens. It starts with a letter, it does not end with a hyphen, and it
contains no two consecutive hyphens. There is no other structure: no dots, no
slashes, no namespace prefix and no encoded discipline.

An identifier may not start with an underscore, which is not reachable under
the syntax above and is stated anyway, because
`docs/decisions/0003-storage-format.md` reserves a leading underscore for
directories that are not quantities and the two rules have to agree in writing
rather than by coincidence.

An identifier is a directory name and a file name on every platform a
contributor might use, which is why the character set is this narrow.

An identifier never changes. Not when the preferred name of the quantity
changes, not when the slug turns out to read badly, and not when a better one
is thought of. Every record's path contains it, every published analysis names
it, and the cost of a rename lands on people outside this project. A better
name is added as an alias.

### The vocabulary, and where it lives

One file per quantity, at `vocabulary/<quantity-id>.toml`, in the tracked TOML
format fixed by `docs/decisions/0003-storage-format.md`.

The vocabulary is separate from `record/` because it is a different kind of
thing. A record is a transcription of something somebody published. A
vocabulary entry is this project's own decision about what a name means, and
mixing the two under one tree would put a decision of ours in a directory
whose contents are supposed to be other people's facts.

### The required fields of a vocabulary entry

- `id`. The identifier. It matches the file name, and a mismatch is a refusal.
- `name`. The quantity's usual name in running prose, capitalised as a person
  would write it.
- `definition`. Prose, and the field the whole record is about. See below.
- `dimension`. The dimension of the quantity, spelled as
  `docs/decisions/0007-units.md` fixes. There is no set to choose from: that
  record gives a rule for the canonical unit rather than a list of dimensions,
  and a compound dimension is as admissible as a one-word one. This sentence
  named a set until #104, and two of the first entries written under it carried
  a dimension no such set held.
- `canonical_unit`. The coherent SI unit for that dimension, per the same
  record. The entry is the authority for this quantity's dimension and canonical
  unit and that record is the authority for how the unit is chosen and how both
  are written down.
- `includes`. A list of boundary cases that are this quantity, each a short
  sentence.
- `excludes`. A list of boundary cases that are not, each a short sentence and
  each saying what they are instead.
- `techniques`. The closed set of measurement techniques admitted for this
  quantity, each with an `id` under the identifier syntax above, a `name` and a
  one-sentence definition of what the technique does. This is the codebook the
  record's method field is coded against, and it is per quantity because a
  technique name means nothing outside the quantity it measures. A technique
  entry may carry an optional `family`, naming a technique family shared across
  quantities, for the question of whether a drift is per family rather than per
  technique. There is no catch-all value: a technique the set does not have
  arrives as a one-line addition to this file in the same pull request as the
  record that needs it, which is a reviewable change rather than a free-text
  field growing quietly.
- `aliases`. A list of other names the quantity is published under. May be
  empty, and empty is written rather than the field being omitted.
- `definition_epochs`. The closed set of definitional regimes for this quantity,
  as `docs/decisions/0007-units.md` requires. Each entry carries an `id` and,
  where the regime began on a nameable date, a `from`. The initial regime of a
  quantity whose definition has never changed omits `from`, and that absence is
  what marks it as initial; inventing a start date for it would put a
  manufactured fact in the file whose job is to say what is known. A quantity
  with one regime says so with one entry rather than by omission.
- `status`. `active` or `superseded`. A superseded entry carries
  `superseded_by`, keeps its file and keeps its records, for the same reason a
  withdrawn record does in
  `docs/decisions/0012-where-correction-history-lives.md`.
- `external`. A list of entries, each with a `scheme` and an `id`, pointing at a
  settled identifier for this quantity in an external scheme. May be empty.
  Empty means no external identifier has been recorded, not that none exists,
  and the two are different statements.

### What `definition` has to do

It has to decide membership. That is a higher bar than describing the
quantity, and it is the bar because the definition is what a contributor holds
a candidate paper against.

A definition passes when a reader who knows the field can take a published
number, read the definition, and reach an answer without asking anybody. It
fails when the answer depends on what the reader already assumed. The
`includes` and `excludes` lists exist because prose alone rarely reaches that
bar: the lists are where the cases that actually caused an argument get
written down, and an entry gains an entry in one of them every time a case is
decided.

`excludes` says what the excluded thing is instead. An exclusion that only
says no leaves the contributor with a number and nowhere to put it, and a
contributor with nowhere to put a number invents an identifier.

### When a new identifier is warranted rather than an alias

One rule decides it.

A NEW IDENTIFIER IS WARRANTED WHEN TWO CANDIDATE VALUES COULD DIFFER FOR A
REASON THAT IS NOT MEASUREMENT ERROR. If two names refer to something whose
true value is the same number, they are one quantity and the second name is an
alias. If the two could disagree because they are defined against different
things, they are two quantities.

Read the rule in the direction that is hard rather than the one that is easy.
The tempting move is to split whenever two sets of measurements disagree, and
that move is always available, always looks like caution, and always destroys
the finding. A disagreement between two ways of measuring one quantity is the
most interesting thing this corpus can hold. Splitting it into two identifiers
converts it into two series that agree with themselves, and the tension stops
existing.

Two consequences follow and both are stated so the rule is not read as a
preference for lumping.

A measurement technique is not a quantity. Where two techniques target the
same defined thing, they share an identifier and the technique lives in the
record's own coded method field, which is what
`docs/decisions/0004-record-schema.md` requires it for. The question "does the
value drift per method" is then askable, and it is one of the questions this
corpus exists to answer.

A change of the defining convention is not an alias. Where the definition
itself moved, `definition_epochs` carries the regimes and
`docs/decisions/0007-units.md` refuses conversion across them. That keeps one
identifier and still prevents pooling across the boundary.

### The neutron lifetime, decided

ONE IDENTIFIER, `neutron-lifetime`, with the technique coded on each record.

The beam technique counts the decay products emerging from a beam of known
neutron density. The bottle technique confines neutrons and counts how many
survive after a known interval. Both are aimed at the same defined thing, the
mean lifetime of the free neutron, and neither definition mentions the
apparatus. Under the rule above they are one quantity.

The disagreement between the two sets of results is the reason this decision
matters and the reason it goes the way it does. It is one of the clearest live
examples of the effect this corpus was assembled to study, and it is
expressible only if both sets sit under one identifier with their technique
coded. Under two identifiers there is no tension, only two well-behaved
series, and the corpus would have destroyed its best case while looking
tidier.

The condition that would reverse it is written in the reversal section rather
than here, and it is a real one: the two techniques are not obviously
sensitive to the same thing, since counting decay products and counting
survivors differ if neutrons disappear by a route that produces no counted
product. If that turns out to be the explanation rather than an error, they
are two quantities and this decision was wrong.

### The proton charge radius, decided

ONE IDENTIFIER, `proton-rms-charge-radius`, with the technique coded on each
record.

Muonic hydrogen spectroscopy, ordinary hydrogen spectroscopy and electron
scattering are three routes to the same defined thing, the root-mean-square
charge radius of the proton. Each route reaches it through different theory,
and the theory input is part of what a value depends on, which is a fact about
the extraction and not about the quantity.

That theory dependence is recorded rather than used to justify a split. The
record carries the technique in its coded method field and the theory input in
the transcriber's note, so an analysis can split by route and can state that
the extraction assumed something. A split into separate identifiers would put
the same information into the identifier, where no analysis can recombine it,
and would again remove a tension by renaming it.

### A quantity that became defined rather than measured

The third case in this class is a quantity whose defining convention changed
so that older values are not comparable to newer ones. `definition_epochs`
carries it, and `docs/decisions/0007-units.md` decides what normalization may
do across the boundary, which is nothing.

The sharpest version is a quantity that stopped being measurable at all
because its value was fixed by definition. The speed of light in vacuum is the
standard example: after the metre was defined in terms of it in 1983, a
published value is not a measurement of the same kind of thing. Where that has
happened, the entry says so with the date, and measurements published after it
are not measurements of that quantity. That is a vocabulary statement and it
belongs in the entry rather than in an analysis's exclusion list.

### A different number under a different convention is not an alias

The neutron half-life is not an alias of `neutron-lifetime`. It is a different
number. Under the rule above the two cannot differ for a reason that is not
measurement error, since they are related by the exactly defined factor `1 /
ln 2`, so they are one quantity, and the half-life is a convention rather than
a name.

A published half-life is therefore admitted under `neutron-lifetime`. It is
stored as published, in the unit as published, and the conversion to the
canonical form is the stored exact factor that `docs/decisions/0007-units.md`
already requires, with `conversion.exact = true` because the factor is exactly
defined and imports no uncertainty.

That is the first use of that record's conversion block for something other
than a unit factor, and it is named here rather than discovered later. If the
schema record concludes that the conversion block must hold unit factors only,
this case needs its own field and this paragraph is the thing to revisit.

The rule has a boundary and it is the dimension. Two conventions that differ
in dimension are separate identifiers, whatever the arithmetic between them,
because an entry carries one `dimension` and one `canonical_unit` and cannot
hold two. The decay constant of the free neutron is the reciprocal of its
lifetime and has the reciprocal dimension, so it is its own identifier if the
corpus ever needs one, while the half-life shares the dimension of the
lifetime and does not.

### The entry committed with this record

`vocabulary/neutron-lifetime.toml` is committed with this record and is the
worked example the shape above is checked against. It carries no `external`
entries. NO EXTERNAL IDENTIFIER WAS VERIFIED ON THE ROUTE THAT PRODUCED THIS
RECORD, and an unverified external identifier in a vocabulary entry would be a
fabricated cross-reference in the one file whose job is to say what a name
means. The empty list is the honest state and it is written rather than left
out.

## Reasons

The flat descriptive namespace wins on the cost that decides duplicates, which
is the cost of finding out whether an identifier already exists. Every other
option makes that check more expensive, and the check being expensive is the
mechanism by which the first silent failure happens. The hierarchy loses on
top of that for a reason specific to this project rather than a general one: a
cross-disciplinary corpus cannot start by agreeing a taxonomy of the sciences,
and one that tried would produce no records for a long time.

The external scheme was the closest of the rejected options and it lost on
permanence rather than on coverage. Coverage could be worked around by using
several. Permanence cannot: an identifier here is a directory in a tree and a
string in somebody's published citation, and a scheme that retires an entry
would break both. Referencing external identifiers in a field gets the
cross-referencing benefit with none of that, and the field is where a settled
external identifier belongs.

The rule for splitting is written to be hard to satisfy because the failure it
guards is asymmetric. Merging two quantities that should have been separate
produces a visible mess: the series has a step in it and somebody asks why.
Splitting one quantity that should have been together produces two clean
series and no question, and the finding is gone with nothing left to notice. A
rule that errs has to err toward the failure that announces itself.

The two worked cases are in this record rather than left to the seed corpus
because both are the kind of decision that gets made implicitly by whoever
files the first record. Once a `neutron-lifetime-beam` directory exists, the
decision has been taken by a filename, and reversing it means moving records.
Deciding before the first record is what this milestone is for.

Nothing in this repository refuses any of this. `PROSE, NOT ENFORCEMENT`.
There is no validator here, so an identifier with a capital letter, an entry
with no `definition`, an entry whose `id` disagrees with its file name, and a
record filed under a directory that has no vocabulary entry all pass every
route in this repository today. The structural half of those refusals is owed
by #24 and the meaning half by #25, and the fixture behind each one by #26.
The committed entry below is checked by a reviewer holding this record, and by
nothing else.

## Date

2026-08-08

## Reversal condition

Reverse the single `neutron-lifetime` identifier if the beam and bottle
results are established to differ because the two techniques are sensitive to
different things rather than because one of them carries an error. That is the
case where they stop being one quantity under this record's own rule, and it
is a real possibility rather than a formality. The trigger is a settled
explanation in the literature, not a persistent disagreement, because a
persistent disagreement is the state this decision was taken to be able to
express.

Reverse the single `proton-rms-charge-radius` identifier on the same condition
and for the same reason.

Reverse the flat namespace if the vocabulary reaches a size at which finding
whether an entry exists stops being a directory read. That is a measurable
threshold rather than a feeling, and the measure is whether a contributor can
still answer the question by listing the directory. A few hundred entries is
still a listing. A few thousand is not.

Reverse the no-rename rule for one specific case and no other: an identifier
that is offensive or that names a person in a way the project should not
carry. Everything else lives with its slug.

Revisit the exclusion of external schemes as the identifier space if a single
scheme appears that covers every discipline this corpus reaches and commits to
permanence in writing. Neither half of that exists today and both are needed.
