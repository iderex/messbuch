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

Correcting the filename in #45 belongs to #45 and is not done here; this section
exists so that a reader of the index is not the last to know.

What let it happen is that nothing reads either table, and the command above was
not run before the number was allocated. Running it is the whole guard, and this
paragraph is the only thing that says so.

## What is not decided here

Questions only the maintainer can answer are collected on issue #13 and are not
decided in these records. When one is answered the answer becomes a record and
appears in the table above.
