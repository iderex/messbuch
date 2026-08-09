# The machine readable schema

`record-1.toml` is version 1 of the field set a measurement record may contain.
It is the single authority. A program that needs a field's name, its type, its
presence rule or its value set reads that file; a document that repeated any of
them would drift against it and the document would be what people trusted.

The decision records under `docs/decisions/` are where each field was argued and
they stay the place to argue with one. What they are not is the place a program
reads.

## What reads it, and what still does not

`internal/schema` reads it and `internal/validate` applies it. The structural
leg of the gate is what a contributor meets:

    go run . ci validate-corpus

Nothing in that path restates a field name, a presence rule, a closed value set
or a pattern. All of it is read from this file, so a field added here is checked
without a line of Go moving, and a key written here that the loader does not
read is refused rather than skipped. That last one is the property worth
knowing: a rule written in the authority and applied by nothing would read as
enforced to everybody who opens the file, which is worse than an absent rule.

Half the file is still applied by nobody, and it is the half about meaning.
#25 owes the leg that opens a second file to decide the first, so
`resolves_to` is read by nothing: a `measurement.quantity` naming no entry
under `vocabulary/`, a `group.id` resolving to no file under `group/` and a
`lineage.supersedes` naming a record that does not exist all pass today. Two
conditions in this file say in their own text that no reading of a record
decides them, and the structural leg leaves both alone.

#26 owes the accounting per refusal site rather than per refusal name. The
suite behind the structural leg refuses a named refusal no fixture reaches;
what it does not do is notice a second branch inside one name that no fixture
reaches, and those two are not the same statement.

There is still no command that prints this file. #23 asks for one and it is not
here; what a contributor can do today is read it, which is why it carries a
comment per section and a `means` line per field. What is printed is the
validator's side of it, which is every refusal that reading this file can
produce:

    go run . refusals

## The shape of an entry

Each field is one `[[field]]` table addressed by its TOML path inside a record.

`path`, the TOML path. `type`, either a scalar named under `[scalar.*]` or a
composite named under `[type.*]`. `means`, one line saying what the field is
for. `does_not_mean`, present where a decision record drew the distinction,
because most of the expensive mistakes in a corpus are a field read as the
neighbouring one. `fixed_by`, the decision record the field comes from.
`needed_by`, the issues whose analyses need it, which is the rule
`docs/decisions/0004-record-schema.md` works under: a field naming no analysis
is not carried.

Composite types carry `[[type.<name>.member]]` entries with the same members,
minus `path`, plus `name`.

One type name in `record-1.toml` is neither, and the sentence above did not
predict it:

    grep -n 'identifier-or-literal' schema/record-1.toml
    197:type = "identifier-or-literal"

`measurement.denominator_quantity` is an identifier or one of the values in its
own `literals` list, and nothing under `[scalar.*]` or `[type.*]` carries that
name. The loader reads the suffix as meaning exactly that, so the alternatives
stay in the field's own entry rather than moving into the validator's source.
Whether the file should instead declare the type, or the paragraph above should
admit a third shape, is a question for `record-1.toml` and is not settled by
this note.

## The presence vocabulary

`presence = "required"`. Always present.

`presence = "optional"`. May be absent, and `optional_because` says why in terms
of a fact about the sources rather than about transcriber effort. That
distinction is `0004`'s and it is the difference between a field that is
usable and one that is filled in by whoever bothered.

`presence = "conditional"`. `required_when` and `refused_when` carry the
condition as data: a `field`, and one of `equals`, `not_equals`, `in`,
`present` or `absent`. Two of them do not read a field of the record, and they
say so rather than looking like the others:

- `measurement.denominator_quantity` reads `dimension` from the quantity's
  vocabulary entry, marked `vocabulary_field`. A validator needs the vocabulary
  loaded to decide it, which is decidable but not from the record alone.
- `superseded_by` and the `factor` inside a rescaling block depend on something
  no file states: whether a replacement record exists, and whether the source
  stated a factor. Both carry `machine_decidable = false` and a
  `condition_note`. A validator that silently treated them as ordinary
  conditions would refuse correct records, so the file says which ones it
  cannot decide instead of leaving a checker to find out.

`refused_when = { never = true }` appears twice, on `source.identifier` and
`source.print`. Each is required when the other is absent and neither refuses
the other, because a source can have both.

`absence_is_written = true` marks the fields where "the source did not say" is a
value rather than an omission. That is the structural decision `0004` defends
first: a field left out because the literature was silent and a field left out
because nobody has read that part yet are the same bytes on disk.

    grep -c 'absence_is_written = true' schema/record-1.toml
    4

Four, where `0004` names three. The fourth is `source.resolvable`, which comes
from `docs/decisions/0008-provenance.md` and is the same pattern under another
record: a transcriber who had no identifier and one who did not look are
different states. Marking it here is a reading of `0008` rather than something
`0004` says, and it is the only place the two records are joined under one flag.

## The version, and what happens to a record written against an older one

The version is in the file name and in `schema_version` at the top of it. A
record carries `schema_version` and that number names the file it was written
against.

A new version is a new file. `record-1.toml` is never edited into
`record-2.toml`, and no version's file is ever deleted, because a corpus is read
record by record and a record naming a version whose file is gone is a record
nothing can read.

Every change a record can observe increases the version by one. A new field, a
removed field, a value added to or taken out of a closed set, a changed presence
rule, a changed pattern, or a changed meaning. There is no minor part and no
patch part: an addition that looks harmless is exactly what an older tool meets
as an unknown field, and `0004` requires that to fail loudly rather than be
skipped. A typo in a `means` line is the only kind of change that does not move
the number, and it moves nothing else either.

Records are not rewritten when the version moves. They keep the number they were
written against, and the tool reads every version it has a file for. Migrating a
record to a newer version is a change to that record, travels the correction
path in `docs/decisions/0012-where-correction-history-lives.md` if it changes
anything a reader cited, and is never done in bulk by a script nobody reviewed.

#61 is the check that makes this unavoidable, and it does not exist yet. Until it
does, nothing stops a version bump that leaves an old corpus unreadable.

## Names that originate here

Ten names in `record-1.toml` are not written in any decision record. Each is
marked `named_here = true` so that the list can be produced from the file rather
than trusted from this paragraph:

    grep -c 'named_here = true' schema/record-1.toml
    10

They fall into two groups and both are cases where a decision record requires a
value and fixes no name for it.

`docs/decisions/0007-units.md` requires the factor's relative uncertainty to be
stored and requires the factor's authority to be named on the record, and names
neither field. They are `factor_relative_uncertainty` and `factor_source`.
`docs/decisions/0004-record-schema.md` points at `0007` for both, so the two
records between them require the values and fix no names.

`docs/decisions/0008-provenance.md` fixes the closed set of identifier schemes
and the bibliographic fields of a print source, and names no field for either.
They are `scheme` and `value` inside a source identifier, and `publisher`,
`volume`, `issue`, `pages`, `year` and `authors` inside a print block.

`0004` sets the rule for this case: where an earlier record proposes a name in
passing, the schema record decides and the earlier one is amended. That
amendment is owed to `0004`, `0007` and `0008` and is not made here, because it
is a change to `docs/decisions/` rather than to this directory. Until it lands,
those two records require a value and this file is the only place its name
exists.

One name is fixed against a disagreement rather than an absence.
`docs/decisions/0005-uncertainty.md` writes its examples at
`measurement.value` and `measurement.uncertainty`; `0004` writes the same fields
at `measurement.published.value` and `measurement.published.uncertainty` and says
in its own opening that it wins and the other is amended. This file follows
`0004`.

## What this version cannot express

A limit that has been normalized. `0004` says the normalized block carries the
same value and uncertainty components in the canonical unit and says nothing
about a bound, so `normalized.value` here is refused when the published block
holds a limit rather than a value, and there is no `normalized.limit`. A limit
published in a non-canonical unit therefore has no comparable form, and the only
honest way to file one today is
`normalization_status = "not-convertible-across-redefinition"`, which says
something that is not true of it.

That is a gap in the decision records rather than in this file, and inventing a
member to close it here would be this file deciding a question `0005` and `0007`
have not been asked. It is written down so the first transcriber to meet it
finds the argument rather than a silent refusal.
