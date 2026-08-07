# 0003 The corpus storage format and file layout

## Question

How is the corpus stored in this repository: at what granularity, in what
format, under what directory layout and file naming rule, and is the artifact
that consumers download the same thing as the files contributors edit?

This decision fixes the layout and the serialisation. It does not fix the field
set of a measurement record, which is the record schema and is decided
separately. A file placed by this decision and a field named by that one are
different questions and are answered in different records.

## Options considered

Granularity: one file per published measurement; one file per quantity holding
the whole series; one file for the whole corpus.

Tracked format: TOML; YAML; JSON; CSV; a database file.

Artifact: ship the tracked files as they are; build a separate artifact from
them.

## What each option costs

One file per measurement. The diff of a contribution shows exactly the number
being added and nothing else, and two people adding measurements of the same
quantity never touch the same file, so the common concurrent case produces no
merge conflict. The cost is a very large number of very small files. A corpus
of ten thousand measurements is ten thousand files, directory listings stop
being readable, and reading a series means reading a directory rather than a
file.

One file per quantity. A series reads as a series, the history of a quantity
sits in one place, and the file count stays small. The cost lands exactly where
the work happens: two people transcribing different papers into the same series
collide on the same file, and the merge is a hand-resolved conflict in a table
of numbers, which is the worst place available for one. The review also loses
resolution, because a diff adding four lines to a hundred-line file does not
show which record a reviewer is being asked to check.

One file for everything. Unreviewable at any size the project cares about, and
it makes every contribution a conflict with every other. It is listed because
it is the shape a corpus drifts into if nothing decides otherwise.

TOML. Unambiguous, hand writable, comments allowed, and no surprise typing:
a value's type follows from its syntax and is not inferred from what it looks
like. The cost is that deep nesting reads badly, arrays of tables are the part
contributors get wrong, and it is less familiar than YAML to people arriving
from other data projects.

YAML. The most familiar of the four to most contributors and the most pleasant
to read at a glance. The cost is the type traps, which are not a theoretical
concern for this corpus in particular: an unquoted value can change type
silently, and a project whose entire subject is the exact form of a published
number cannot afford a format where an unquoted token silently becomes a
different type than the contributor meant.

JSON. Universally parseable and unambiguous. The cost is that it has no
comments, and a transcribed measurement very often needs a note saying what the
source actually said, why a value was read the way it was, or which of two
tables in the paper it came from. Pushing that into a data field makes it a
field the schema has to define and the validator has to police; keeping it as a
comment loses it entirely. JSON also has no trailing-comma tolerance, which
makes a hand edit likelier to produce a parse error than a wrong value, though
that particular cost is a mild one.

CSV. Compact, and every tool reads it. It loses structure the moment
uncertainty is more than one number, which for this corpus is the ordinary case
rather than the exception, and it has no place to put a note or a provenance
locator without inventing column conventions that no parser enforces.

A database file. Fast to query and impossible to review. A diff over a binary
file shows nothing, so every contribution would have to be trusted or
re-extracted, which removes the one property that makes a public corpus
checkable.

Ship the tracked files as they are. Nothing to build, and what a consumer holds
is exactly what the repository holds. The cost is that the consumer has to walk
a large directory tree and parse a hand-writable format, and that every
consumer reimplements the walk. It also removes the place where derived values,
normalisation and a version stamp would live.

Build a separate artifact. Consumers get one file they can load, and it can
carry derived and normalised values that would be noise in a tracked source
file. The cost is a build step that has to be correct and reproducible, a
second format to document, and the standing risk that the artifact and the
source drift apart if nothing checks that the artifact was regenerated.

## Choice

Granularity: one file per published measurement.

Layout: `record/<quantity-id>/<file>.toml`. One directory per quantity, named
by the quantity identifier from the controlled vocabulary. No nesting below
that.

File naming rule: `<year>-<first-author-slug>-<nn>.toml`, where `<year>` is the
four-digit year of publication of the source, `<first-author-slug>` is the
first author's family name reduced to lower-case ASCII letters and hyphens, and
`<nn>` is a two-digit sequence number starting at `01` that distinguishes
several measurements of the same quantity published in the same year by the
same first author. The sequence number is always written, including when there
is only one, so a second measurement never forces the first file to be renamed.

Tracked format: TOML.

Built artifacts, generated from the tracked source and never edited by hand:

- `messbuch-corpus-<version>.jsonl`, one JSON object per line, one line per
  record, lines sorted by a total order the build defines. This is the citable
  artifact and the one an analysis reads.
- `messbuch-corpus-<version>.csv`, a flat projection. It is lossy by
  construction, because the uncertainty representation does not fit a fixed
  column set, and it carries that statement in its own header rather than only
  in this document. It exists for a reader who wants to look at the corpus in a
  spreadsheet and it is not the artifact a citation names.

Reserved directory: a quantity directory whose name begins with `_` is not a
quantity. The build excludes it and the validator does not treat it as a
vocabulary entry. `record/_example/` is the only such directory today and holds
the layout example committed with this record. The exclusion is one rule with
one syntactic trigger so that it is mechanical rather than a list somebody
maintains.

## Reasons

One file per measurement is chosen for the concurrency cost, not for the diff.
Both matter, but the diff advantage is a convenience and the conflict
behaviour is a limit on how many people can work on this corpus at once. A
corpus whose contribution shape produces hand-merged conflicts in tables of
numbers has a ceiling on contributors that no amount of documentation raises,
and hand-merging a table of numbers is precisely the operation that introduces
the transcription errors this project exists to make findable. The file count
is a real cost and it is paid: it is absorbed by tooling, and tooling for
walking a directory is a solved problem in a way that merging numeric tables is
not.

TOML is chosen over YAML on the type question alone. Every other axis is close
enough that familiarity would decide it, and familiarity favours YAML. The
type question is not close. This corpus's whole content is the exact form of
published numbers and the exact form of their uncertainties, and a format in
which an unquoted token can silently become a different type than the
contributor intended fails at the one thing the corpus is for. JSON loses on
comments for a reason specific to transcription work: the note saying what the
source actually said is often the only thing that lets a later reader check the
transcription, and it belongs next to the value rather than in a field the
schema has to invent for it.

The tracked source and the built artifact are separated because their two
audiences want opposite things. A contributor wants a small file they can read
and edit by hand with a comment next to the number. A consumer wants one file
with a stable shape they can load in a line of code. Serving both from one
format punishes one of them, and it is the contributor who gets punished in
practice, because the consumer's requirement is the louder one.

JSON Lines is chosen for the machine artifact over a single JSON document
because it streams, because a diff between two releases is readable line by
line, and because a byte-for-byte reproducibility requirement is easier to hold
over a line-sorted file than over a nested document whose key order has to be
pinned everywhere. The sort order is defined by the build rather than left to
the language's map iteration, which is the specific defect the reproducibility
issue on the corpus milestone names.

Nothing in this repository refuses a file that breaks the layout or the naming
rule. `PROSE, NOT ENFORCEMENT`. The validator that would refuse a file in the
wrong place for what it claims to be is the structural validator, which is
open on the corpus milestone and does not exist today, so a record committed
at a path this document forbids passes every route here. Until it lands, the
layout is checked by a reviewer.

## Date

2026-08-07

## Reversal condition

Reverse the granularity if the file count becomes the binding cost rather than
a paid one. The signal is measurable and should be measured rather than argued:
a checkout or a validator run whose time is dominated by per-file overhead
rather than by parsing, at a corpus size the project actually has. A guess
about a corpus ten times larger than the one in hand does not reverse this.

Reverse the tracked format if a contributor-facing editor exists that writes
the records, because at that point the hand-writability of the format stops
being a constraint and the argument reduces to what the tooling parses most
safely.

Reverse the CSV projection, meaning drop it, if it is observed being cited. A
lossy projection that people cite is worse than no projection, and its header
warning is prose that nothing enforces.

Revisit the reserved-underscore rule if a real quantity identifier ever needs
to begin with an underscore. The vocabulary decides identifier syntax, and if
it permits a leading underscore then this rule and that one are in conflict and
this one gives.
