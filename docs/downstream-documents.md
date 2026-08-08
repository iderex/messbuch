# Which documents are downstream of which parts of the tree

The rule: when a change makes a document wrong, the fix lands in the same pull
request, or an issue is opened on the current milestone before the pull request
merges. There is no third option, because documentation debt is invisible and
the person who discovers it is always a user rather than a maintainer.

This board has a sharper version of the problem than most. The decision records
and the correction path describe things a contributor follows literally while
transcribing measurements by hand. A guide that names a field which has since
been renamed does not merely confuse somebody, it produces a batch of records
that have to be redone.

This map exists so that the question on the pull request template can be
answered honestly rather than guessed at. Read the rows covering what you
touched.

## The map

| If you change this | These go wrong |
| --- | --- |
| A job `name:` or a trigger in `.github/workflows/` | `docs/required-checks.md`, which names each check by the exact string the branch protection matches and states whether it reports on every pull request, and `docs/quality-parity.md`, whose section on which legs are in place is a statement about this workflow set |
| Which issue lands a check, or a check landing and closing its issue | `docs/required-checks.md`, whose `Built by issue` column is the same claim, and `docs/quality-parity.md`, which names the issue behind every adopted and adapted leg and goes stale the same way |
| The ruleset on the default branch | `docs/required-checks.md`, whose whole premise is what that ruleset currently requires, and `docs/release-checklist.md`, one of whose items is whether that ruleset requires anything at all |
| What a release publishes, when it is cut, or which issue owes an item of it | `docs/release-checklist.md`, where every item names the command, the document or the open issue that settles it, so an item whose owner moves leaves the list pointing at nobody |
| The set of labels on the tracker | `.github/ISSUE_TEMPLATE/wrong-number.yml`, which names a label in its front matter; a renamed or deleted label makes the form apply one that does not exist |
| The directory layout, the file naming rule or the tracked format under `record/` | `docs/decisions/0003-storage-format.md`, `docs/corrections.md`, and `record/_example/1900-example-01.toml`, which is the committed illustration of the layout |
| Any uncertainty field name or the `uncertainty_status` value set | `docs/decisions/0005-uncertainty.md`, which writes out all seven cases in those field names, `docs/decisions/0004-record-schema.md`, which lists them as fields of a record, and `docs/corrections.md`, which describes absent against zero in them |
| Any unit, conversion or normalization field name | `docs/decisions/0007-units.md`, which states the relationship between the published and the normalized value in those names, and `docs/decisions/0004-record-schema.md`, which lists them as fields of a record |
| Any field name, required-or-optional decision, or closed value set in the record schema | `docs/decisions/0004-record-schema.md` first, then every record that writes a field name out: `docs/decisions/0005-uncertainty.md`, `docs/decisions/0007-units.md`, `docs/decisions/0008-provenance.md`, `docs/decisions/0012-where-correction-history-lives.md`, `docs/corrections.md`, `record/_example/1900-example-01.toml`, and `docs/curation.md`, which names a field wherever the judgement it teaches is about that field |
| Which value a transcriber takes when a source gives several, how an unusual uncertainty is written, whether a conversion may be done by hand, what counts as a second reading, or how a group is chosen | `docs/curation.md`, which is followed literally by somebody transcribing measurements, so a rule that moves without it produces a batch of records that have to be redone rather than merely confusing a reader |
| Dropping an analysis issue from a milestone | `docs/decisions/0004-record-schema.md`, which names per field the analysis that needs it, so a dropped analysis leaves a field with no reason |
| The quantity identifier syntax, the shape of a vocabulary entry, or the `techniques` set of any quantity | `docs/decisions/0006-quantity-identity.md`, every file under `vocabulary/`, and `docs/decisions/0004-record-schema.md`, whose `measurement.quantity` and `method.technique` are coded against them |
| A provenance field name, the `confirmation` value set, the `statement_kind` set, or which half of the resolvability check runs in the gate | `docs/decisions/0008-provenance.md` and `docs/decisions/0004-record-schema.md` |
| The implementation language, the pinned toolchain version, or the path of the package permitted to reach the network | `docs/decisions/0002-language-and-toolchain.md`, and `docs/decisions/0009-offline-by-default.md`, which names that package and leaves its path to the language record |
| The name of the harness that needs the outside world, or what the gate says about it | `docs/decisions/0010-headless-tests.md` and `docs/decisions/0002-language-and-toolchain.md`, which rests its plotting argument on the same rule |
| Which package may reach the network, or the name of the check that refuses the rest | `docs/decisions/0009-offline-by-default.md`, which fixes both, and `README.md` and `NOTICE.md` once the personal-data statements land under #58 |
| A stamp field, the corpus version rule, or what a release is | `docs/decisions/0011-corpus-versioning.md` and `docs/decisions/0012-where-correction-history-lives.md`, which states the version consequence of a correction |
| The shape of a `[[correction]]` entry or the `kind` set | `docs/decisions/0012-where-correction-history-lives.md`, `docs/corrections.md`, and the wording in `.github/ISSUE_TEMPLATE/wrong-number.yml` about what happens to a report |
| Which pooling model an analysis runs by default, or which between-study variance estimator it uses | `docs/decisions/0013-pooling-default.md`, which fixes the default and requires the model, the estimator and the small-sample correction to be named in every output |
| The seven parts of a decision record or the numbering rule | `docs/decisions/0001-how-decisions-are-recorded.md`, `docs/decisions/template.md`, and `docs/decisions/README.md`, which carries both the record table and the reserved-number table |
| Adding, superseding or renumbering any decision record | `docs/decisions/README.md`. An index that does not list a record is wrong in the one place a reader goes to find records |
| The shape of a `group/` entry, the rule for choosing which group a measurement belongs to, or what a group entry may not carry | `group/README.md`, which is where all three are written, and `docs/decisions/0004-record-schema.md`, whose `group.id` fixes the path an entry lives at and the identifier syntax it takes. Every existing entry under `group/` goes wrong too, since a shape change is a change to files already written against the old one |

## Rows that are owed and are not here yet

Most of this tree does not exist. The map covers what is in the repository and
it is not a plan for what a finished project's map would look like. The parts
below have no rows because they have no upstream to point at, and each becomes a
row in the pull request that creates it.

The build, the toolchain pin and the single local gate command. The contributing
document will name that command and nothing else as the local gate, so the
command and the document move together.

The machine readable schema. The field set itself now has a row, and the
generated schema file that #23 owes will be downstream of both. The curation
guide is downstream of the field set too and now has its own row above.

The `group/` registry now has a row, above, and it is a row about the shape
rather than about the entries. `docs/decisions/0004-record-schema.md` still
requires `group.id` to resolve to an entry and there is still no entry, because
no record in the corpus names a group. The pull request that files the first
record carrying a `group.id` lands the entry it resolves to, and
`group/README.md` is where that obligation is written.

The command line surface. Its help output, the quickstart and the operator
documentation are all downstream of it, and the quickstart is the one that fails
hardest, because it is followed literally on a clean machine.

The changelog. It is downstream of the corpus version rule and of the correction
`kind` set.

The measured shape of the deviation distribution. `README.md` opens by saying
what the deviations between experiments turn out to look like and what that
means for the plus-minus in the literature. Nothing in this repository has
measured that. It is the finding the corpus is being assembled to test, and #46
says in its own first line that the question is asked of the corpus rather than
assumed, so the front page currently states an answer the project has not
reached. When the analysis produces one, the paragraph is downstream of it in
both directions: if the corpus agrees, the sentence gains the command that
produced it, and if it does not, the sentence is wrong in the most public place
this project has. The row is owed by #46 and #47, and either of them lands it.

Adding a row is the cheap half. Noticing that one is owed is the expensive half,
and that is what the question on the pull request template is for.

## What checks this

Nothing.

No check in this repository reads this file. Nothing compares it against the
tree, nothing refuses a workflow rename that leaves `docs/required-checks.md`
stale, and nothing refuses a pull request whose template answer was left
untouched. This is a rule about how work is done rather than a property of the
tree, so no check here could decide it, and none is claimed.

The template makes the answer hard to skip silently and the review is where a
skipped one is caught. Saying so plainly is part of the rule: a rule that
pretends to be enforced is worse than one that admits it is not, because a
reader who believes a machine is watching stops watching.
