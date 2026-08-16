# 0001 How a decision is recorded here

## Question

Where do the architecture decisions of this project live, and in what shape,
so that somebody reading them in five years can tell whether the reason behind
each one still holds?

## Options considered

Leave the decisions on the issue tracker only. Write them as free prose in the
readme. Write them as numbered files in the repository with a fixed set of
parts. Write them as numbered files with no fixed shape.

## What each option costs

Tracker only. Nothing extra to maintain, and the argument stays next to the
comments that produced it. The cost is that a tracker is a record of what was
argued and not of what the project currently is. A decision reversed three
comments later reads the same as one that stands, the search is by keyword
rather than by structure, and the record is gone if the tracker is.

Free prose in the readme. The lowest friction to write. The cost is that the
reasons compress away first. Prose gets edited toward the current state, the
option that was rejected disappears, and a year later nobody can tell whether
the choice was reasoned or inherited.

Numbered files with a fixed set of parts. Every record answers the same
questions in the same order, so a reader can find the reversal condition of any
decision without reading the whole file, and a reviewer can see a missing part
without knowing the subject. The cost is real: it is more to write, it makes
short decisions feel heavy, and a fixed shape invites a part being filled with
a sentence that says nothing rather than left visibly empty.

Numbered files with no fixed shape. Cheaper to write than the above and it
keeps the records in the repository. The cost is that completeness stops being
checkable. Whether a record states its reversal condition becomes a judgement
about the prose rather than a question about a heading, and the part that gets
dropped is always the same one.

## Choice

Numbered files in `docs/decisions/`, one decision per file, each with the seven
parts below, plus a committed template that a new record starts from and an
index that lists every record.

The seven parts, in this order. A record missing any of them is incomplete.

1. The question. What was actually undecided, written so that somebody who does
   not already know the answer can see what was at stake.
2. The options considered. The ones genuinely on the table, including the one
   that was chosen. An options list with a single entry is a decision that was
   not taken.
3. What each option costs. Per option, not a single paragraph of drawbacks. The
   cost of the chosen option is written as plainly as the cost of the rejected
   ones, because a record that makes the choice look free is not believed later.
4. The choice. What was decided, stated so that a reader can tell whether a
   given change conforms to it.
5. The reasons. Why this option beat the others, referring to the costs above
   rather than repeating them.
6. The date. The day the decision was taken, in `YYYY-MM-DD`.
7. The condition that would reverse it. What would have to become true for this
   decision to be looked at again. This is the part that gets dropped and the
   part that matters most. A decision taken because a library did not exist yet
   says so, so that the day it exists somebody knows to look again. Where no
   condition can be named, the record says that and says why, which is a
   different statement from leaving the heading off.

## Reasons

The three rules this project works under require that an asserted fact carries
the command that produced it, that a rule nothing refuses is marked as prose
rather than presented as enforcement, and that no guard ships without proof it
bites. All three are about the difference between a claim and a checkable
statement, and a decision record with no fixed shape cannot be checked at all.
With the seven parts, the check a reader performs is mechanical: the headings
are either there or they are not.

The fixed shape also costs the least where the cost is worst. The expensive
failure is not a heavy record, it is a decision whose reason was lost while the
decision survived, because that produces a project full of constraints nobody
can argue with. The reversal condition is the direct answer to that failure and
it is the part a free-shape record loses first.

The template is committed for one reason. A new record started from the last
record somebody happened to open inherits that record's structure, including
whatever it got wrong, and inherits its subject matter as a set of headings
that now need deleting. A template inherits nothing.

The records live in the repository rather than only on the tracker because the
tracker is where a thing is argued and the repository is what the thing
currently is. Both are kept. Neither replaces the other.

A record in this form refers to itself, and that is a convention of the form
rather than a slip of voice. Part 7 asks what would reverse THIS decision, part
4 asks what a change would have to conform to, and neither can be written
without the record naming itself as the thing being reversed or conformed to. So
sentences of the shape "this record", "this decision" and "the record says that
and says why" are expected here and are not a style defect, and a survey of the
fleet's prose that counts them should count them against this paragraph. The
convention stops at the seven parts: outside them, a document explaining why it
exists is the ordinary defect it is everywhere else, and this file carries no
such sentence.

Nothing in this repository refuses a record that is missing a part.
`PROSE, NOT ENFORCEMENT`. There is no validator over `docs/decisions/` today
and no check on any pull request reads these files, so a record dropping its
reversal condition passes every route here. The completeness of a record is a
thing a reviewer holding the template checks by eye. This paragraph is the
whole disclosure and the template repeating the parts does not soften it.

## The numbering rule

Records are numbered with four digits, starting at `0001`, and the number never
changes once the record is committed. The filename is the number, a hyphen, and
a short slug of the question in lower case, for example
`0001-how-decisions-are-recorded.md`.

A number is allocated when the record is written, not reserved in advance,
with two consequences that are stated rather than discovered.

Numbers are not reused. If a record is superseded, the new record takes the
next free number and the old file stays where it is with a line at the top
naming the record that replaced it. Deleting a superseded record deletes the
reason the project once had, which is the failure this whole document exists
against.

Gaps are permitted. Several people write records at once, and an issue that
names the filename of the record it owes has reserved that number by naming it
even if the record never lands. A gap in the sequence means a number was
reserved and not used. It does not mean a record was removed, and a reader
finding one should look for the issue that named it rather than assume a
deletion.

## Date

2026-08-07

## Reversal condition

Reverse this if the fixed seven parts are observed to be filled with empty
sentences rather than left visibly incomplete, because at that point the shape
is producing the appearance of a reason instead of a reason and is worse than
free prose. The signal is a reversal condition that restates the choice, or an
options list whose rejected entries were written after the decision to justify
it.

Reverse the file-per-decision layout if the project acquires a route that reads
these records mechanically and that route needs a different shape to do it. A
record shape chosen for a human reader and a record shape a machine can refuse
are not automatically the same, and if one has to give, this document is the
one that gives.
