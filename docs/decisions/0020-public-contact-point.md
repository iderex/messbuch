# 0020 The public contact point for security and for conduct

## Question

A security policy and a code of conduct both need somewhere a person can write
that is not the public tracker. What is that somewhere, and is it one place or
two?

Whatever is chosen becomes permanently public and is scraped from the day it
lands, so the question is partly about how much surface to expose for the two
different things that need it.

## Options considered

A dedicated address for both. Two separate addresses, one for security and one
for conduct. The platform's private reporting form for security, with a
separate address for conduct only.

## What each option costs

One address for both. The least to publish and the least to maintain. The cost
is that a security report and a conduct complaint arrive in the same place, and
the two need different handling and different discretion.

Two separate addresses. The cleanest separation of the two kinds of message.
The cost is twice as much address published and twice as much scraped, for a
separation that can be had another way.

The reporting form for security, with an address for conduct. The security half
publishes no address at all: the form gives a private channel with a record, and
a place where a fix and an advisory can be published together. The cost is that
it binds the security channel to the platform this repository is hosted on, and
a reporter who will not use that platform has nowhere else to go.

## Choice

Security reports go through this repository's private vulnerability reporting
form. No address is published for them.

    gh api repos/iderex/messbuch/private-vulnerability-reporting
    {"enabled":true}

Conduct complaints need an address, and it is one address rather than two.

    git show origin/main:SECURITY.md | grep -n 'private vulnerability reporting form'
    5:Use this repository's private vulnerability reporting form:

## Reasons

The security half needs no published address because the form does the same job
without leaving anything to be scraped, and it does one thing an address does
not: it keeps the report, the fix and the advisory in one place with a record.

Two addresses would have been the cleanest separation and they were rejected on
cost rather than on principle. The form already separates the two kinds of
message, so the second address would have bought a distinction that was already
bought, and paid for it in scraping.

`PROSE, NOT ENFORCEMENT`. Nothing in this repository refuses a security report
filed as a public issue, and nothing checks that the form stays enabled. The
command above reads the platform rather than the tree, which is why it is
pasted in `SECURITY.md` as well: a reader can re-run it rather than trust either
document.

## What is not decided here

The domain of the conduct address. It is open, it is the maintainer's, and it is
entry 6 of issue #13.

Until it is answered there is no code of conduct in this repository:

    git ls-files | grep -Ei 'CODE_OF_CONDUCT'
    (no output, exit 1)

`SECURITY.md` covers the security half today and is complete on its own terms.
The conduct half is a file that does not exist, and this record does not create
it.

## Date

2026-08-09

## Reversal condition

Reverse the security half if this repository moves off a platform that offers a
private reporting form, or if the form is observed to turn away a reporter who
had something worth reporting. A reporter who could not reach anybody leaves no
trace, so the signal for the second case will arrive late and by another route
if it arrives at all, and that is a weakness of this choice rather than of the
condition.

Reverse the single conduct address if the volume or the nature of what arrives
at it makes one inbox the wrong place, which is a judgement somebody makes after
reading what arrives rather than a threshold that can be set now.
