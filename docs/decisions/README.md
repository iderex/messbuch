# Decision records

Every architecture choice that shapes this project is written down here with
its reasons, so that somebody reading it later can tell whether the reason
still holds. The shape of a record, the seven parts it must have and the
numbering rule are set out in
[0001 How a decision is recorded here](0001-how-decisions-are-recorded.md).

A new record starts from [template.md](template.md), not from a copy of the
last record somebody happened to open.

This index is part of a record's landing. Adding a record and not listing it
here leaves the index wrong, and nothing in this repository refuses that.

## Records

| Number | Title | Date |
| --- | --- | --- |
| 0001 | [How a decision is recorded here](0001-how-decisions-are-recorded.md) | 2026-08-07 |
| 0002 | [The implementation language and toolchain](0002-language-and-toolchain.md) | 2026-08-08 |
| 0003 | [The corpus storage format and file layout](0003-storage-format.md) | 2026-08-07 |
| 0004 | [The measurement record schema](0004-record-schema.md) | 2026-08-08 |
| 0005 | [How uncertainty is represented](0005-uncertainty.md) | 2026-08-07 |
| 0006 | [Quantity identity and the controlled vocabulary](0006-quantity-identity.md) | 2026-08-08 |
| 0007 | [Units, normalization and redefinitions](0007-units.md) | 2026-08-07 |
| 0008 | [Provenance, and what makes a record checkable](0008-provenance.md) | 2026-08-08 |
| 0009 | [Offline by default, and where the network surface lives](0009-offline-by-default.md) | 2026-08-07 |
| 0010 | [Headless and unelevated tests, and where the rest goes](0010-headless-tests.md) | 2026-08-08 |
| 0011 | [How an analysis names the corpus it ran on](0011-corpus-versioning.md) | 2026-08-07 |
| 0012 | [Where the history of a corrected value lives](0012-where-correction-history-lives.md) | 2026-08-07 |
| 0013 | [Which pooling model is the default](0013-pooling-default.md) | 2026-08-07 |
| 0014 | [The deviation statistic and the reference value it needs](0014-deviation-statistic.md) | 2026-08-08 |
| 0015 | [What counts as a tension, what counts as real, and what the probability is about](0015-what-counts-as-a-tension.md) | 2026-08-08 |
| 0016 | [The license of this repository](0016-repository-license.md) | 2026-08-08 |
| 0017 | [The license of the corpus, which is not the license of the code](0017-corpus-license.md) | 2026-08-09 |
| 0018 | [Whether a corpus release gets an archived identifier, and under what name](0018-archived-identifier.md) | 2026-08-09 |
| 0019 | [Whether contributions from outside are accepted before the first release](0019-outside-contributions-before-first-release.md) | 2026-08-09 |
| 0020 | [The public contact point for security and for conduct](0020-public-contact-point.md) | 2026-08-09 |

## Numbers reserved and not yet used

None today. The five that were reserved, `0002`, `0004`, `0006`, `0008` and
`0010`, are in the table above and their issues are the ones that named them.
The sequence has no gaps in it as a result, and that is a fact about today
rather than a property the numbering rule promises: gaps are permitted and a
future one means a number was reserved and not used, not that a record was
removed.

Reproduce the reservation set rather than trusting this paragraph:

    gh issue list --repo iderex/messbuch --state open --limit 200 \
      --json number,body \
      --jq '.[] | select(.body | test("docs/decisions/[0-9]{4}-")) | "#\(.number) " + ([.body | scan("docs/decisions/[0-9]{4}-[a-z0-9-]+")] | unique | join(" "))'

What that prints is wider than this table and has to be read rather than
counted. An issue that merely depends on a record names its filename too, and
the command cannot tell a dependency from a reservation. What it is good for is
the opposite direction: a number appearing there that is neither in this table
nor in the record table above is a reservation nobody wrote down.

## A number that was reserved and then taken

`0012` was reserved by one issue and written by another. The reservation was
missed.

    $ gh issue view 45 --json number,createdAt --jq '"#\(.number) \(.createdAt)"'
    #45 2026-08-07T19:33:55Z
    $ gh issue view 45 --json body --jq .body | grep -n '0012'
    24:Done when: docs/decisions/0012-deviation-statistic.md states the definition, the
    $ git log --format='%H %ad' --date=iso-strict origin/main \
        -- docs/decisions/0012-where-correction-history-lives.md
    4fb51c5546775ba0aee82c82005790b0160a751a 2026-08-07T23:35:01+02:00

So #45 held `0012` by naming it, and a different record was written onto that
number afterwards. The rule decides the rest and leaves nothing to taste: a
number never changes once the record is committed, so
`0012-where-correction-history-lives.md` keeps `0012`, and the record #45 owes
takes the next free number. The next free number is the one neither command
below claims:

    $ gh issue list --repo iderex/messbuch --state open --limit 200 --json body \
        --jq '[.[].body | scan("docs/decisions/([0-9]{4})-") | .[0]] | unique | join(" ")'
    0002 0004 0005 0006 0008 0009 0010 0012
    $ ls docs/decisions/ | grep -oE '^[0-9]{4}' | tr '\n' ' '
    0001 0002 0003 0004 0005 0006 0007 0008 0009 0010 0011 0012 0013

`0014`. Both outputs are what they printed on 2026-08-08, in the working tree
that added `0002`, `0004`, `0006`, `0008` and `0010`, and the first one still
shows those five because the issues that reserved them were open at the moment
it ran. Run them again rather than reading these; the answer they are for is a
number that moves.

The record is now written and it took `0014`, which is the number the rule above
gives it. It is `0014-deviation-statistic.md`, and it is in the table at the top
of this file. The filename in issue 45 is corrected to match, in the issue
rather than here, so that the Done-when and the tree name one file between them.
Read what the directory holds rather than this paragraph:

    git ls-tree --name-only origin/main docs/decisions/

What let it happen is that nothing reads either table, and the command above was
not run before the number was allocated. Running it is the whole guard, and this
paragraph is the only thing that says so.

## What is not decided here

Questions only the maintainer can answer are collected on issue #13 and are not
decided in these records. When one is answered the answer becomes a record and
appears in the table above.

Five of them have been answered and are `0016` through `0020`. Which record
answers which entry is listed on #13 rather than here, so that the mapping has
one home. One entry of that issue is open and has no record: the domain of the
conduct address, which `0020-public-contact-point.md` names in its own section
on what it does not decide.
