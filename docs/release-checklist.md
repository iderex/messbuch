# The release checklist

What has to be true before a tag is cut, in one list, so that a release is a
decision somebody makes rather than a thing that happens.

Every item names the command or the document that settles it. Where neither
exists yet, the item names the issue that owes it and says so, because an item
with no way to answer it is an item that gets waved through, and a checklist
whose entries cannot be checked is worse than none: it produces a signed-off
release nobody actually verified.

Nothing has been released:

    gh release list --repo iderex/messbuch --limit 5
    (no output)

    git tag | wc -l
    0

So this list has never been walked. The first release is the first walk, and the
answers from it are recorded against the tag.

## The items

### 1. The gate is green on the commit being tagged, and the run is named

Not answerable yet. There is no gate.

The single local command is owed by #14 and the workflow wrapping it by #16.
Until both land, nothing here builds or tests anything, and the five workflows
that do run read commit trailers, dependency manifests, Unicode and the workflow
files.

When it lands, the answer is the run rather than the workflow: a green run
identified by its own id, on the exact commit being tagged, not on a branch tip
that has moved since. `docs/required-checks.md` is where the check names are
fixed.

### 2. The required check set is configured on the protected branch

Answerable now, and the answer today is that it is not.

    gh api repos/iderex/messbuch/rules/branches/main --jq '[.[].type] | join(" ")'
    deletion non_fast_forward pull_request

No `required_status_checks` rule, so every check on this board reports without
blocking anything. Until that changes, the gate is advice, and a release cut
against advice is a release whose checks were optional. `docs/required-checks.md`
holds the strings to require and names the two that must never be required.

### 3. Every decision the release depends on has its record

Answerable now.

    git ls-tree --name-only origin/main docs/decisions/ | wc -l
    15

That count includes the index and the template, so it is a starting point rather
than an answer. `docs/decisions/README.md` is the index and carries both the
record table and the reserved-number table, and a reserved number with no record
behind it is the thing to look for. The question to ask against it is not how
many records exist but whether anything the release behaves according to is
decided in a pull request body instead of in a record.

### 4. The corpus validates, and its record count is stated

Half answerable now. The count is:

    git ls-files 'record/*' | grep -v '^record/_example/' | wc -l
    0

`record/_example/1900-example-01.toml` is excluded because it is not a corpus
record; it is the committed layout illustration and its numbers are invented.

The validation half is owed by #24 for structure and #25 for meaning. A count
with no validation behind it says how many files are in a directory, which is not
the same statement, and the release notes may not present it as one.

### 5. The reference number fixtures are green

Not answerable yet. Owed by #43.

A release of an engine whose estimators are unpinned is a release of numbers
nobody has checked. The fixtures compare this board's estimators against
published reference results, and they are the only thing standing between a
plausible number and a correct one.

### 6. The reproducibility claim is verified or withdrawn

Not answerable yet. Owed by #28.

The two states are not equally easy and that is deliberate. Verified means the
artifact was built twice and the bytes compared, with the command in the release
record. Withdrawn means the release notes do not claim reproducibility at all.
There is no third state where the claim is made because it was made last time.

### 7. The changelog is complete, including corrected records

Not answerable yet. Owed by #55, which also decides whether the tool and the
corpus release together or separately.

The corrected records are the entries that matter most. Somebody may have
published a number computed from a value this release changes, and the changelog
is the only place they can find that out. `docs/corrections.md` is where the
correction path is written and `docs/decisions/0012-where-correction-history-lives.md`
is where the shape of a correction entry is fixed.

### 8. The quickstart has been walked on each supported platform

Not answerable yet. Owed by #57.

Walked means followed literally on a clean machine with no network, by somebody
reading it rather than by somebody who knows the answer. A quickstart verified by
its author is a quickstart nobody has tested, in the same way and for the same
reason as the curation guide's acceptance test in #31.

Which platforms are supported is part of what #57 settles, and a platform nobody
walked is named as not walked rather than left out of the list.

### 9. The license is present

Answerable now, and the answer is that it is present.

    git ls-files | grep -iE '^LICEN[SC]E'
    LICENSE

AGPL-3.0, taken by the maintainer on 2026-08-08 as entry 1 of #13. Whether the
file is tracked and what the platform reads it as are two questions, and the
second is answered from the default branch rather than from a branch:

    gh api repos/iderex/messbuch --jq '.license.spdx_id'

This item blocked the release outright while it was open. Nobody outside could
use, fork or package anything here, and the sign-off check on every pull request
asked contributors to certify that they may submit their work under a license
that did not exist. Both of those end with the file.

The license on the corpus is a different question and is still open. It is entry
2 of #13, and the archived identifier question there also touches what a release
publishes, so this item being answered does not release the two below it.

### 10. The inventory is published

Not answerable yet. Owed by #53 for generating it and #56 for publishing it with
the release.

It covers the corpus artifact as well as the binary's dependencies. A corpus
release with no statement of which series and which revisions it holds is an
opaque blob, and the person who needs to know is a downstream packager who cannot
unpack it to find out.

## What this release deliberately does not claim

These go in the release notes rather than into a correction six months later. A
limitation stated in advance is a scope; the same limitation stated afterwards is
a retraction, and it costs the reader's trust in everything beside it.

The corpus covers a handful of quantities and a small part of published science.
It is not a survey of physics and no absence in it is evidence of an absence in
the literature.

The calibrated prior rests on the reference class stated in its own output. A
probability derived from this corpus is a probability given that reference class,
and it does not transfer to a quantity the corpus does not cover.

Where a record is unconfirmed, one person has read it. The confirmation breakdown
is printed by every analysis for that reason, and the release notes state the
breakdown of the corpus being shipped rather than leaving it to be looked up.

## How the walk is recorded

The answers are recorded against the tag, not in this file. This file is present
tense and says what has to be true; a record of one release's answers is past
tense and belongs with that release.

An item answered as not applicable carries the reason beside it. An item skipped
carries the word skipped rather than an empty line, for the same reason
`docs/decisions/0005-uncertainty.md` writes absence down instead of leaving a
field out: a skipped item and an item nobody reached look identical afterwards.

## What checks this

Nothing.

No check in this repository reads this file, and nothing refuses a tag that was
cut without walking it. Two of its items are about mechanisms that do not exist
yet, and one is about a decision no mechanism can take. This is a rule about how
a release is made rather than a property of the tree, so no check here could
decide it, and none is claimed.

The items that carry commands are the ones a later reader can re-run, and where
an item's command disagrees with what this file says the answer was, the command
is right.
