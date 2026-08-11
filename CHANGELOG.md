# Changelog

Two version numbers live here and they are not the same thing.

The tool's version follows the ordinary meaning. A breaking change is one to the
command line surface or to the output format, and somebody who wrote a script
against the last version has to change it.

The corpus's version is defined by `docs/decisions/0011-corpus-versioning.md`
and its rule is about what happens to a number somebody already cited rather
than about how much work went in. Correcting a transcription is a MAJOR change
there and adding a whole series is a MINOR one, which inverts the intuition from
software and inverts it on purpose: a fix here changes a fact a reader may have
published, and the version number is the warning.

That is why a corrected record is the entry in this file that matters most.
Somebody may have computed and published a number from the old value, and this
file is where they find out.

## Released together or separately

Separately, and that is not decided here. It is already fixed by
`docs/decisions/0011-corpus-versioning.md`, which says the tool version moves on
its own schedule, that a corpus release does not imply a tool release, and that
the reverse does not hold either. A single stream would tie a data correction to
a tool release it has nothing to do with, and a reader watching for corrections
would have to read every tool release to find out whether one happened.

The cost is two release streams rather than one, which is the honest price and
is paid in this file by two sets of headings rather than one.

A corpus release is the tag `corpus-v<version>`. A tool release is the tag
`v<version>`. The tag is what a citation names, because it survives an artifact
being lost, re-uploaded or served from somewhere else, and an artifact that
disagrees with its tag is a defect in the release.

## What is checked

`go run . ci changelog` refuses a release tag with no entry here and an entry
here naming a version no tag exists for, in both directions, and it refuses this
file for having nowhere to write the next change. What it cannot judge is
whether an entry says anything useful, and no check here claims to.

Nothing has been released:

    git tag
    (no output)

So every heading below `Unreleased` is empty, and that is the state of the
project rather than a template left unfilled.

# The tool

## Unreleased

Nothing yet. The first tool release is the one the release checklist is walked
against, and it does not exist.

# The corpus

## Unreleased

Nothing yet. No record has entered the corpus:

    git ls-files 'record/*' | grep -v '^record/_example/' | wc -l
    0

When the first series lands it is a MINOR entry here, and every correction to a
record that has already been released is a MAJOR one naming the record, the old
value and the new.
