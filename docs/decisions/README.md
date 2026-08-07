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
| 0003 | [The corpus storage format and file layout](0003-storage-format.md) | 2026-08-07 |
| 0005 | [How uncertainty is represented](0005-uncertainty.md) | 2026-08-07 |
| 0007 | [Units, normalization and redefinitions](0007-units.md) | 2026-08-07 |
| 0009 | [Offline by default, and where the network surface lives](0009-offline-by-default.md) | 2026-08-07 |
| 0011 | [How an analysis names the corpus it ran on](0011-corpus-versioning.md) | 2026-08-07 |

## Numbers reserved and not yet used

The gaps above are not deletions. Under the numbering rule an issue that names
the filename of the record it owes has reserved that number by naming it. These
five files do not exist in the tree today, and this table is a pointer to the
issue that holds each one, not a claim that anything is written.

| Number | Slug named by the issue | Issue |
| --- | --- | --- |
| 0002 | `language-and-toolchain` | #3 |
| 0004 | `record-schema` | #5 |
| 0006 | `quantity-identity` | #7 |
| 0008 | `provenance` | #9 |
| 0010 | `headless-tests` | #11 |

Reproduce that list rather than trusting it:

    gh issue list --repo iderex/messbuch --state open --limit 200 \
      --json number,title,body \
      --jq '.[] | "#\(.number) " + ([.body | scan("docs/decisions/[0-9]{4}-[a-z0-9-]+")] | join(" "))'

## What is not decided here

Questions only the maintainer can answer are collected on issue #13 and are not
decided in these records. When one is answered the answer becomes a record and
appears in the table above.
