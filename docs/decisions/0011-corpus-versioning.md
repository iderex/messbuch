# 0011 How an analysis names the corpus it ran on

## Question

Somebody is going to cite a number this tool produced. What does the output
have to carry so that the number stays reproducible while the corpus keeps
changing underneath it, and what does a corpus version number mean?

## Options considered

Carry nothing, and expect the reader to note what they ran.

Carry a timestamp and the tool version.

Carry a corpus revision identifier only.

Carry a full stamp: corpus version and revision, tool version and revision, the
options in force, and the record selection, with no wall-clock time in it.

## What each option costs

Nothing. Zero effort and the output stays clean. It costs the project the one
property that makes a cited number defensible. A pooled estimate with no
statement of what it was computed from cannot be checked by anybody, including
its author six months later, because the corpus that produced it no longer
exists in that state.

Timestamp and tool version. Looks like provenance and is not. The timestamp
names when, which does not identify the corpus unless somebody can map a time
onto a revision, and it makes the output impossible to reproduce byte for byte
because the bytes change on every run. That last cost is not incidental: it
removes the only mechanical check that a reproduction actually reproduced.

Corpus revision only. Identifies the data exactly and is cheap. It costs the
other half. The same corpus revision analysed with a different between-study
variance estimator, a different small-sample correction, or a different record
selection gives a different number, and none of that is recoverable from a
revision identifier.

Full stamp, no wall-clock time. Everything needed to re-run is present and the
output is reproducible byte for byte, so a reproduction can be checked by
comparing files rather than by reading them. It costs output space on every
result, it costs a discipline that every option default has to be printed
rather than assumed, and it removes the convenience of knowing when a result
was produced from the result itself.

## Choice

The full stamp, with no wall-clock time in it.

### What a corpus version number means

Three parts, `MAJOR.MINOR.PATCH`, and the rule is about what happens to a
number somebody already cited rather than about how much work went in.

`MAJOR` increases when a value that was already published as part of a release
becomes different. That covers a corrected transcription, an uncertainty
restated, a record withdrawn, a record's quantity identity reassigned, and a
schema change that makes an existing consumer's read break. What these have in
common is the only thing that matters: a reader holding the previous release
and a reader holding this one disagree about a number.

`MINOR` increases when records are added and nothing that was already there
changes. Adding a whole series is a minor change. This is the common case and
it is deliberately the cheap one.

`PATCH` increases when nothing analysis-visible changes at all: a note
corrected, a comment added, an alias added to a vocabulary entry, a
reformatting.

So correcting a transcription error and adding a series are not the same kind
of change, and the correction is the bigger one. That inverts the intuition
from software, where a fix is a patch and a feature is a minor, and it inverts
it for a reason. In software, a fix moves the program toward what it was always
supposed to do and the caller's contract is unchanged. Here, a fix changes a
fact the caller may have published. The version number is a warning to a
downstream reader, and it has to be loudest exactly where a previously correct
citation has become wrong.

### Whether a release is a tag or an artifact

Both, and the tag is the authority.

A corpus release is a git tag on this repository, named `corpus-v<version>`.
The tag is what a citation names, because a tag is a name for an exact tree that
survives the artifact being lost, re-uploaded or served from somewhere else.

The built artifacts are produced from that tag by the ordinary build, published
with checksums, and are a convenience for anybody who does not want to clone.
An artifact that disagrees with its tag is a defect in the release and the tag
wins. Publishing an artifact from anything other than a tag is not a release.

The tool version is separate and moves on its own schedule. A corpus release
does not imply a tool release and the reverse is not true either.

### The stamp fields

    stamp_version        the version of the stamp format itself
    corpus_version       e.g. 2.3.0, or "unreleased" when the revision is not
                         at or descended from a corpus tag
    corpus_revision      the full 40-character commit identifier
    corpus_state         "clean" or "dirty"
    corpus_dirty_digest  present only when corpus_state is "dirty"
    tool_version         the released version of the tool
    tool_revision        the full commit identifier the tool was built from
    tool_state           "clean" or "dirty"
    command              the subcommand that produced the output
    options              every analysis option in force, including the ones
                         left at their default, sorted by name
    selection            the record selection expression as given
    selected_count       records that entered the analysis
    excluded             per-reason counts of records the selection or the
                         analysis excluded

`options` carries the defaults as well as the values the operator typed. A
default that is not printed is a default that can change between versions and
silently change a published number, which is the failure this whole record
exists against. The same argument applies to `excluded`: an analysis of a
convenient subset and an analysis of a whole series produce the same-looking
number, and the excluded counts are what tells them apart.

There is no wall-clock time in the stamp. That is the field a reader expects
and it is left out deliberately, because its presence would make byte-for-byte
reproduction impossible and byte-for-byte reproduction is what turns a
reproduction claim into a check. Where a date is genuinely wanted, the commit
date of `corpus_revision` is fixed, is recoverable from the revision, and is a
better answer than when somebody happened to run the command.

### How the stamp is carried, per output format

JSON and JSON Lines. A top-level `stamp` object, or for line-delimited output a
first line that is a JSON object with a single `stamp` key. Structural, so no
consumer has to parse prose.

CSV. Leading lines each beginning with `#`, one field per line, before the
header row, and an identical sidecar file named `<output>.stamp.json`. The
sidecar exists because `#` comment lines are a convention rather than part of
any CSV specification, and a reader whose parser chokes on them would otherwise
have to strip the provenance to read the data. The lossiness of CSV is already
stated in the storage record and the stamp does not repair it.

Plain text, meaning what a terminal shows. A block at the end of the output,
introduced by a line reading `-- stamp --`, then one `key: value` per line in
the field order above. It sits at the end because the reader came for the
result and should not have to scroll past provenance to reach it.

The cost of that placement is stated rather than left to be discovered: an
output truncated by a pager or a pipe loses the stamp, and the plain-text
format is therefore the one where provenance is easiest to separate from the
number. That is a reason to cite from a machine-readable format, and the
documentation says so instead of pretending the text format is as safe.

### Running against a working tree with uncommitted changes

The tool does not refuse. It reports.

`corpus_state` is set to `dirty`, `corpus_dirty_digest` carries a digest over
the uncommitted content so that two dirty runs can at least be compared to each
other, and the tool prints a warning line on standard error saying the result
is not reproducible from the stamp. The output is still produced, because
refusing would make the tool useless during exactly the work it exists to
support, which is transcribing and checking records before they are committed.

The reproduction path refuses instead. Given a stamp with `corpus_state` of
`dirty`, the reproduce command exits non-zero and says the stamp does not
identify a tree it can check out. That puts the refusal where the claim is
being made rather than where the work is being done.

### The command that reproduces a stamped output

    messbuch reproduce --from result.json

It reads the stamp out of the file, checks out `corpus_revision`, verifies the
running tool's version and revision against `tool_version` and `tool_revision`,
re-runs `command` with `options` and `selection` exactly as recorded, and
compares the produced bytes against the file it was given. It exits zero only
on an identical byte sequence, and on any difference it prints the first
differing field rather than a diff of the whole file.

THIS COMMAND DOES NOT EXIST. `PROSE, NOT ENFORCEMENT`. There is no tool in this
repository, no command line surface and no analysis code; the command line
surface and the deterministic build are open issues on later milestones. The
paragraph above specifies what `reproduce` must do so that the stamp fields are
chosen against a real consumer rather than guessed, and it describes nothing
anybody has run. No documentation of this project may state the reproduction
property as a fact until the command exists and has been observed to fail on a
deliberately altered output.

## Reasons

The stamp is full rather than minimal because every field that was dropped in
the rejected options turned out to be a field that changes the number. A
revision alone leaves the estimator choice free, options alone leave the data
free, and a timestamp constrains nothing while costing the reproducibility
check.

Leaving the wall-clock time out is the decision most likely to be questioned
and it is the one the rest depends on. Once a time is in the output, two runs
of the same analysis on the same data produce different bytes, and every
statement about reproducibility becomes a statement a person has to verify by
reading rather than one a command can decide. The project's rule is that a
guard is proven by watching it bite, and a stamp containing a timestamp cannot
be guarded that way.

`MAJOR` for a correction is the other decision that will look wrong at first.
It is chosen because the version number's audience is a downstream reader
deciding whether their citation still holds, not a maintainer describing effort.
A reader who sees the major part move knows to check; a reader who sees a patch
bump does not, and under any other rule the corrected transcription is exactly
the change they would have needed to check.

## Date

2026-08-07

## Reversal condition

Reverse the no-timestamp rule if the reproduction check moves to comparing
parsed content rather than bytes, since at that point a varying field costs
nothing and the convenience is free. Do not reverse it before then, because the
byte comparison is the whole reason the rule exists.

Reverse the `MAJOR`-for-a-correction rule if corrections turn out to be so
frequent that the major part moves several times a year and stops carrying
information. At that point the signal has been diluted by volume and the right
answer is a separate corrections feed rather than a version part; the trigger is
a count of major releases per year, and it is a count the release process can
print.

Revisit the stamp field list whenever an analysis option is added that changes
a result and is not covered by `options`. That is a defect rather than a change
of mind, and the field list is amended without argument.
