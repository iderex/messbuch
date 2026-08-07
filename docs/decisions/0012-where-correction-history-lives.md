# 0012 Where the history of a corrected value lives

## Question

A corpus of transcribed numbers will contain wrong numbers. When one is
corrected, the old value does not simply disappear, because somebody may
already have used it. Where does the previous value live?

The deciding question, and the one every option below is judged against: can
somebody holding only a released artifact see that a value was corrected?

## Options considered

The version control history alone.

An append-only correction list inside the record file itself.

A separate corrections file, one per corpus release, that records refer to.

The changelog alone.

## What each option costs

Version control alone. Free, already exists, complete, and never gets out of
step with the data because it is the data's history. It fails the deciding
question outright. A consumer holding a downloaded artifact, or a single file
somebody sent them, has no history at all, and even a consumer with a clone has
to know which file to look at and has to read a diff to work out whether the
change was a correction or a reformatting. The people most likely to be holding
a stale number are exactly the people least likely to have a clone.

An append-only list inside the record. It travels wherever the record travels,
including into the built artifact and into a single file copied out of it, so
the deciding question is answered by construction. It costs record size on
every record that has ever been corrected, it costs the schema a structure that
has to be validated, and it creates a second place where a value can be wrong,
since a correction entry is itself transcribed by hand and can be entered
badly.

A separate corrections file per release. The records stay clean and there is
one place to read what changed between two releases, which is genuinely the
nicest thing to read. It costs the consumer a join, and it fails the deciding
question in the common case: a consumer who loaded the corpus and kept the
records has thrown the corrections file away, and a single record extracted
from the artifact carries nothing. It also drifts, because nothing ties an
entry in that file to the record it describes except a string.

The changelog alone. Human readable and already required for other reasons. It
is prose, so nothing downstream can act on it, and a consumer running an
automated comparison against a previous release cannot use it at all.

## Choice

The correction history lives in the record, as an append-only list, and it is
carried into the built artifact.

The changelog also carries corrections, as their own class of entry rather than
mixed in with additions, because a reader scanning a release wants to see
whether anything they cited moved. The changelog is the summary; the record is
the authority. Where the two disagree, the record is right and the changelog is
a defect.

Version control keeps the full history as it always does. Nothing about this
decision changes that, and nothing about the guarantee rests on it.

Shape of an entry, in the tracked TOML format:

```toml
[[correction]]
date = "2026-08-07"
kind = "transcription"
field = "measurement.published.value"
was = 4.774e-10
now = 4.803e-10
reason = "second digit after the point misread from the printed table"
reported_in = "#123"
```

`kind` takes one of five values, and the set is closed because the classes have
different consequences for a reader.

`transcription`. The source says one thing and the record said another, through
a copying error. The ordinary case.

`not-in-source`. The record claimed something the source does not say at all,
as opposed to saying it differently. This is the serious one. It means the
provenance failed rather than that a digit slipped, and it is visible as its own
class everywhere it appears rather than folded in with typographical errors,
because a reader deciding whether to trust the corpus needs to be able to count
them separately.

`duplicate`. The same published measurement entered the corpus twice.

`withdrawal`. The record should not be in the corpus and is not being replaced
by a corrected value.

`metadata`. A note, an alias or a comment changed, with no effect on any value,
uncertainty, unit, date or citation. This is the only kind that does not move
the major part of the corpus version.

Records are never deleted, including duplicates and withdrawals. A withdrawn
record keeps its file, gains `status = "withdrawn"` and a `superseded_by`
pointer where one exists, and stays in the built artifact so that a consumer
holding a citation to it can resolve what happened. Deleting it would turn a
correctable citation into a dangling one, which is a worse outcome than the
duplicate was.

Every correction moves the major part of the corpus version, per
`docs/decisions/0011-corpus-versioning.md`, except a `metadata` correction,
which moves the patch part. That is the same rule stated from the other side: a
version part exists to warn a reader whose citation may have moved.

## Reasons

Only one option answers the deciding question in the case that actually
matters, which is a consumer holding an artifact rather than a clone. That
consumer is the reason the correction path exists at all. Every other option is
nicer to read or cheaper to build and leaves that person unable to discover that
their number changed.

The cost that is paid knowingly is the second place a value can be wrong. A
correction entry carries the old value, which is itself transcribed, and a
mistyped `was` field is a new defect introduced by the machinery for fixing
defects. The mitigation is that the previous value is also in version control,
so the entry is checkable against the diff that made it, and the correction
review is where that check belongs.

The closed `kind` set exists so that `not-in-source` can be counted. A corpus
that reports how many of its corrections were misreadings of a source and how
many were fabrications is telling a reader something about its own
trustworthiness, and a single undifferentiated correction count hides exactly
the number a sceptical reader would want.

Nothing in this repository refuses any of this. `PROSE, NOT ENFORCEMENT`. There
is no schema, no validator, no changelog and no build, so a correction entered
without a `kind`, a withdrawn record deleted outright, or a major version that
did not move all pass every route here. The obligations above are checked by a
reviewer, and the checks that would refuse them are owed by the validator issues
on the corpus milestone and by the changelog issue on the release milestone.

## Date

2026-08-07

## Reversal condition

Reverse the in-record list if corrections become frequent enough that the
history dominates the record. The trigger is measurable and should be measured
rather than argued: the share of bytes in the built artifact taken by correction
entries. If that becomes the largest part of the artifact, the corpus has a
transcription quality problem that this decision is the wrong instrument for.

Reverse it also if a consumer-facing index appears that makes a correction
discoverable from a value alone, without a clone and without the record. That
would answer the deciding question by another route, and at that point the
record can go back to holding only the current value.

Revisit the closed `kind` set the first time a real correction does not fit any
of the five. A new class is a reason to extend the set, not a reason to file the
correction under the nearest wrong one.
