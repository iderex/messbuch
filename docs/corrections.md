# Reporting and correcting a wrong number

This corpus contains wrong numbers. That is not a possibility to guard against,
it is a certainty to plan for, and the measure of the project is how a wrong
number is reported, confirmed and corrected rather than whether one exists.

## Reporting one

Open an issue from the template `A number in the corpus is wrong`. It asks for
three things and nothing else: which record, what the source actually says, and
where in the source it says it.

You do not have to be a contributor, you do not have to know the file layout,
and you do not have to be sure. A report that turns out to be a misreading costs
somebody a few minutes. A wrong number nobody reported stays wrong and gets
cited.

The three fields are required because a report without them cannot be checked by
anybody but its author, and nothing else is required because every extra
required field loses reports. Anything more you want to say goes in the free
field at the end.

## What happens to the report

It is checked against the source, by somebody reading the source, and the check
has one of four outcomes.

The report holds. The correction is made as below.

The report does not hold, because the source says what the record says. The
issue is closed with the locator that was read and what it showed, so that the
next person who notices the same thing finds the answer rather than filing
again.

The report is about the source rather than about the transcription, meaning the
paper itself is thought to be wrong. That is not a defect in this corpus. The
corpus records what was published, including values now believed to be wrong,
because a series of measurements that were later revised is exactly the thing
this project exists to study. The issue is closed saying so.

The report cannot be checked because the source is not reachable. The issue
stays open, says which source and what was tried, and is not closed by
assumption in either direction.

## Making the correction

A correction is not an ordinary edit. It changes a number somebody may already
have used.

The record keeps what it said before, what it says now, and why it changed, in
an append-only list inside the record file itself. That list travels into the
built artifact, so a consumer holding only a downloaded release can see that a
value moved without having a clone of this repository. The shape of an entry
and the reasons behind putting it in the record rather than in version control
alone are in
[0012 Where the history of a corrected value lives](decisions/0012-where-correction-history-lives.md).

The previous value is not deleted, and neither is a record that turns out not to
belong in the corpus. A withdrawn or duplicated record keeps its file, is marked
withdrawn, points at the record that survives where one does, and stays in the
built artifact. Deleting it would turn a citation somebody can still resolve
into one they cannot.

The changelog carries corrections as their own class of entry, separate from
additions, so that a reader scanning a release can see whether anything they
cited moved without reading the whole entry list.

Every correction except a pure metadata change moves the major part of the
corpus version, per
[0011 How an analysis names the corpus it ran on](decisions/0011-corpus-versioning.md).
That is deliberately the opposite of the software convention, where a fix is the
smallest kind of release. Here a fix is the change most likely to invalidate
something a reader published.

## The three cases that are not a typo

A duplicate. The same published measurement entered the corpus twice, usually
because one of the two was found through a review rather than through the
primary source. One record survives, the other is withdrawn and points at it,
and the correction is recorded on both.

A record whose source does not say what the record claims. This is the serious
case and it is recorded as its own class rather than as a transcription error.
The difference is not pedantry: a transcription error means a digit was copied
wrong, and this means the provenance failed, so the question is not only what
the value should be but how the record came to exist. Nothing else in the same
contribution is trusted until it has been checked.

A record whose uncertainty was reported as absent when the source gave one, or
as a number when the source gave none. Absent and zero are different values in
this corpus and neither is a guess, so this is a correction to the value rather
than to a note.

## What nothing checks

Nothing in this repository refuses a correction that skips any of the above. No
route reads this document, no check compares a corrected record against its own
correction list, and nothing requires the changelog entry to exist. The report
template can be bypassed by opening a blank issue.

What stands behind this path is a person reading a source and a reviewer reading
the diff. The checks that would refuse a malformed correction entry are owed by
the validator issues on the corpus milestone, and until they land the paragraph
above is the whole of the enforcement story.
