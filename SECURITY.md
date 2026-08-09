# Security

## Reporting a vulnerability

Use this repository's private vulnerability reporting form:

    https://github.com/iderex/messbuch/security/advisories/new

It is open. Read that rather than trusting this line:

    gh api repos/iderex/messbuch/private-vulnerability-reporting
    {"enabled":true}

There is no published email address, and that is a decision rather than an
omission. An address in a file gets scraped the day it lands, and the form does
the same job without leaving anything to scrape. It was taken on 2026-08-09 and
is recorded on issue #13, entry 5.

Do not open a public issue for a vulnerability. Everything else belongs on the
public tracker, including the case below.

## What to expect

No response time is promised here, because nobody can hold a promise this
document cannot enforce. What the form gives you is a private channel with a
record, and a place for a fix and an advisory to be published together.

Nothing has been released:

    gh release list --repo iderex/messbuch --limit 5
    (no output)

So there is no supported-versions table and no back-porting policy yet. The only
version is the default branch.

## What is in scope

The interesting surface of this project is code that reads input somebody else
produced, and there is exactly one shape of that.

The record parser and the validator read files that arrive by pull request and
files a user assembled on their own machine. A crafted record that makes the
parser allocate without bound, recurse without bound, read outside its input, or
take pathological time is in scope, and it is the defect class this project
expects to have.

The analysis code reads a corpus that may not be this one. An operator can point
the tool at their own directory, so every assumption the analysis makes about
its input is an assumption about a file somebody else wrote.

The build and release path is in scope: anything that would let a change reach a
published artifact without going through the ordinary route.

A finding does not have to be reachable today. Most of the code above is not
written yet, and a defect in what exists is in scope whether or not a caller
reaches it.

## What is not a vulnerability here

A wrong number in the corpus. That is a data defect, it is expected, and the
measure of this project is how one gets reported and corrected rather than
whether one exists. It belongs on the public tracker, through
[the form for it](https://github.com/iderex/messbuch/issues/new?template=wrong-number.yml),
which asks three things and nothing else. `docs/corrections.md` is what happens
to it afterwards. Filing such a report privately slows the correction down for
no gain, because there is nothing confidential about a number that is already
published in a paper.

A record whose source does not say what the record claims is the serious version
of the same thing and is still not a vulnerability. It is the case
`docs/corrections.md` treats separately, and it goes on the public tracker too.

Something the project already says it does not do. `docs/decisions/0009-offline-by-default.md`
puts every network-capable import in one package and states plainly that the
paragraph describes an intention and that no run has been observed. A report
that the intention is not yet a proven property is welcome and is a defect
report on the public tracker rather than a vulnerability, because the
documentation does not claim otherwise.

## What this document does not do

Nothing here is enforcement. No check in this repository reads this file and
none could: whether a report was handled, and how fast, is not a property of the
tree. The form is a real channel and the rest of this document is a description
of judgement.
