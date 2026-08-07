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
| A job `name:` or a trigger in `.github/workflows/` | `docs/required-checks.md`, which names each check by the exact string the branch protection matches and states whether it reports on every pull request |
| The ruleset on the default branch | `docs/required-checks.md`, whose whole premise is what that ruleset currently requires |
| The set of labels on the tracker | `.github/ISSUE_TEMPLATE/wrong-number.yml`, which names a label in its front matter; a renamed or deleted label makes the form apply one that does not exist |
| The directory layout, the file naming rule or the tracked format under `record/` | `docs/decisions/0003-storage-format.md`, `docs/corrections.md`, and `record/_example/1900-example-01.toml`, which is the committed illustration of the layout |
| Any uncertainty field name or the `uncertainty_status` value set | `docs/decisions/0005-uncertainty.md`, which writes out all seven cases in those field names, and `docs/corrections.md`, which describes absent against zero in them |
| Any unit, conversion or normalization field name | `docs/decisions/0007-units.md`, which states the relationship between the published and the normalized value in those names |
| Which package may reach the network, or the name of the check that refuses the rest | `docs/decisions/0009-offline-by-default.md`, which fixes both, and `README.md` and `NOTICE.md` once the personal-data statements land under #58 |
| A stamp field, the corpus version rule, or what a release is | `docs/decisions/0011-corpus-versioning.md` and `docs/decisions/0012-where-correction-history-lives.md`, which states the version consequence of a correction |
| The shape of a `[[correction]]` entry or the `kind` set | `docs/decisions/0012-where-correction-history-lives.md`, `docs/corrections.md`, and the wording in `.github/ISSUE_TEMPLATE/wrong-number.yml` about what happens to a report |
| The seven parts of a decision record or the numbering rule | `docs/decisions/0001-how-decisions-are-recorded.md`, `docs/decisions/template.md`, and `docs/decisions/README.md`, which carries both the record table and the reserved-number table |
| Adding, superseding or renumbering any decision record | `docs/decisions/README.md`. An index that does not list a record is wrong in the one place a reader goes to find records |

## Rows that are owed and are not here yet

Most of this tree does not exist. The map covers what is in the repository and
it is not a plan for what a finished project's map would look like. The parts
below have no rows because they have no upstream to point at, and each becomes a
row in the pull request that creates it.

The build, the toolchain pin and the single local gate command. The contributing
document will name that command and nothing else as the local gate, so the
command and the document move together.

The record schema and the machine readable schema. Every decision record above
that names a field is downstream of the schema, and so is the curation guide.

The command line surface. Its help output, the quickstart and the operator
documentation are all downstream of it, and the quickstart is the one that fails
hardest, because it is followed literally on a clean machine.

The changelog. It is downstream of the corpus version rule and of the correction
`kind` set.

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
