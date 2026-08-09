# 0019 Whether contributions from outside are accepted before the first release

## Question

Does this project take contributions from people outside it now, or only from
the first release onwards?

## Options considered

Accept them now. Decline them until the first release, and say so before
anybody starts. Decline them and say nothing, leaving a contributor to find out
when their pull request is turned away.

## What each option costs

Accepting them now. More data sooner, and more eyes on the schema while it can
still change, which is when a schema criticism is cheapest to act on. The costs
are that every early contribution is written against a schema that is still
moving, that the review load arrives before the validator that is supposed to
carry it exists, and that the license question had to be settled first because
the sign-off certifies to it.

Declining until the first release, stated up front. The schema stops moving
under people and the validator arrives before the review load does. The cost is
the contributions that do not happen, including the schema criticism that would
have been cheapest now, and the impression a closed door leaves on somebody who
came to help.

Declining and saying nothing. No cost that shows up in this repository, which is
why it is listed. The cost lands entirely on the contributor: a refusal
discovered after the work is the refusal plus the work.

## Choice

Contributions from outside are not accepted until the first release, and the
refusal is written at the top of the contributing document, before anything a
contributor would act on.

    git show origin/main:CONTRIBUTING.md | grep -n 'contributions from outside are not accepted yet'
    3:## Read this first: contributions from outside are not accepted yet

Two routes stay open to anybody, and the same document names both: reporting a
number in the corpus that is wrong, and arguing with a decision record on the
tracker.

The position that outside contributions are welcome under a sign-off is dated
rather than cancelled by this. It applies from the first release onwards, and
the sign-off apparatus for it is already running:

    gh api repos/iderex/messbuch/commits/68eb9076e8e07b788d48ee2e11a771c043f56f47/check-runs \
      --jq '.check_runs[] | select(.name=="DCO sign-off") | "\(.name) \(.conclusion)"'
    DCO sign-off success

That is the head of the last pull request to land rather than the tip of the
branch. The check runs on a pull request, so a merge commit carries no verdict
from it, and asking the tip for one returns nothing.

## Reasons

The two costs that decided it are not the same size. A contribution written
against a moving schema has to be redone, and redoing somebody else's careful
transcription is worse than not having had it, because the person who did it
learns that their work was disposable.

The second half of the choice carries more weight than the first. Declining is
defensible on its own; declining silently is not, and it is the option that
costs this project nothing while costing a contributor everything they spent.
Putting the sentence at the top of the document is what makes the decision cost
the person who took it rather than the person who reads it.

`PROSE, NOT ENFORCEMENT`. Nothing refuses a pull request from outside this
repository. The hygiene check deliberately does not apply its refusing tier to a
head branch on a fork, on the argument that somebody arriving from outside
cannot know an issue number before the issue exists, so the one check that
distinguishes inside from outside is the one that is gentler to outsiders rather
than closed to them. Whoever handles such a pull request is what carries this
decision.

## Date

2026-08-09

## Reversal condition

This reverses on a date rather than on a judgement: the first release. What has
to become true is a release existing, which is

    gh release list --repo iderex/messbuch --limit 5
    (no output)

today. `docs/release-checklist.md` is the list a release is cut against, and
the contributing document is what has to change on the same day, since it is
what a contributor reads.

Reverse it earlier if the schema stops moving before the release does, because
the stated reason is a moving schema and not a preference for working alone. The
signal is `docs/decisions/0004-record-schema.md` and the machine readable schema
on #23 agreeing and staying agreed through a series being transcribed.
