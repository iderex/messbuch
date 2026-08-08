# The group registry

`docs/decisions/0004-record-schema.md` makes `group.id` a required field on
every record and requires it to resolve to an entry in `group/<group-id>.toml`.
This directory is that registry, and this file is the shape of an entry.

A group entry says who made a measurement, at the granularity at which two
measurements share systematics and share a prior. It is not a directory of
institutions and it is not an author list. Nothing here is a measurement and
nothing here holds a number.

## There are no entries yet, and that is a fact rather than an oversight

The registry is empty because the corpus is empty. No record exists to name a
group:

    git ls-files 'record/*' | grep -v '^record/_example/'
    (no output)

`record/_example/1900-example-01.toml` is not a corpus record and says so in
its own first lines. It names no group and it is excluded from the build by the
leading-underscore rule in `docs/decisions/0003-storage-format.md`.

So the obligation this file creates falls on the first record. A record naming a
group that has no entry here is a coded field with no code set, which is the
trap `docs/decisions/0004-record-schema.md` names by name, and the entry lands
in the same pull request as the record that first needs it.

## The shape of an entry

One file per group, at `group/<group-id>.toml`, in the tracked TOML format
fixed by `docs/decisions/0003-storage-format.md`.

    id = "example-collaboration"
    name = "Example Collaboration"
    aliases = [
      "Example Collaboration at the Example Facility",
      "Example Group",
    ]

    note = """
    Invented. This block shows the shape of an entry and nothing else. There is
    no such collaboration, the names above are not the names of any real group,
    and nothing in this block may be cited or copied into a real entry.
    """

`id`, required. A slug under the identifier syntax of
`docs/decisions/0006-quantity-identity.md`: one to sixty characters of
lower-case ASCII letters, digits and hyphens, starting with a letter, not
ending with a hyphen, and with no two consecutive hyphens. It is the file name
without its extension, and the two always agree. Like a quantity identifier it
never changes, for the same reason: every record carrying it would have to be
rewritten, and a better name is added as an alias.

`name`, required. The name the group publishes under most recently. It means
what to print. It does not mean the group's legal identity, its host
institution, or the name under which any particular measurement was published,
which is what `aliases` is for.

`aliases`, required, and may be an empty list. Every other name the group has
published under, including earlier names, the names of predecessor groups whose
apparatus and analysis this group continued, and the spellings a contributor is
likely to search for. It exists for one reason: a contributor who cannot find
an existing entry creates a second one, and two entries for one group is the
defect that makes `group.id` unable to do the job the schema record gives it.

AN EMPTY `aliases` LIST MEANS NO OTHER NAME HAS BEEN RECORDED. It does not mean
no other name exists. Nobody is obliged to have searched, and reading an empty
list as a finished search is the shape of overstatement this project is about.

`note`, optional, and the one free-text field here. What a later curator needs
to know about how this group's boundary was drawn, in particular which
neighbouring efforts were decided to be outside it and why. NO ANALYSIS READS
THIS FIELD, and nothing may be grouped, filtered or counted by it. It is
carried because the boundary judgements below are the expensive part of the
registry and reconstructing one later costs more than writing it down once.

Nothing else. No institution field, no country, no external identifier and no
membership. Each was considered and left out, because a field no analysis reads
and no curator needs is a field somebody will fill in inconsistently, and an
inconsistently filled field is worse than an absent one when the question is
whether two records came from the same place. Adding one later is a change to
this file with its own reason, which is cheaper than removing one that records
already carry.

## Choosing the group: the granularity rule

The test is the one the schema record states, and it is a test about
systematics rather than about organisations. Two measurements belong to one
group when they share the apparatus, the analysis chain or the calibration
inputs closely enough that a defect in one would move both.

Applied to the three cases that decide it:

A named collaboration spanning several institutions is ONE group. It shares
apparatus and analysis, which is the test, and splitting it by institution
would report a single experiment as several independent ones, which is the
error `group.id` exists to prevent.

A laboratory hosting several measurement efforts is NOT one group. Where two
efforts under one roof use different apparatus and different analysis chains,
they are two groups, and the shared roof is recorded, if it matters at all, as
a shared systematic on the uncertainty component rather than as a shared
identity. Treating them as one would go wrong in the other direction: it would
discard real independent evidence.

A group that renamed itself, or that continued as a successor under a new name
with the same apparatus and people, is ONE group. The continuity is in the
apparatus and the analysis, not in the name. The earlier name goes in
`aliases`, and `note` says when the name changed.

Where the sources do not settle which of these a case is, THE CASE IS ARGUED IN
AN ISSUE BEFORE THE RECORD IS FILED. `group.id` is required, so a guessed
identifier is not a gap that a later reader can see; it is a wrong grouping
that looks exactly like a right one, and it propagates into every pooled
interval computed over the corpus.

## What an entry may not carry

A group entry names an organisation. It does not name a person.

No personal names, no author lists, no individual identifiers, no addresses and
no contact details, in any field including `note` and `aliases`. This holds
even where a laboratory is customarily known by the name of a founder: the
entry carries the name the group publishes under, and nothing further about the
individual.

The hard case is a measurement published by one person, where naming the group
is naming them. Where the source names a laboratory, a department or an
institute under which the work was published, that is the group, and the entry
names it rather than the author. WHERE THE SOURCE NAMES NO SUCH BODY, THE
RECORD IS NOT FILED UNTIL THE CASE IS ARGUED IN AN ISSUE. The right answer may
turn out to be an entry that unavoidably corresponds to one individual, and
that is a decision to take deliberately and in the open rather than by a
transcriber's default while working through a series.

This is a rule about what goes into a public corpus and it is not the same
subject as the operator-side personal-data statements, which #58 owes. The two
agree in direction and neither carries the other.

## `group.id` and `correlation_group` are not the same field

A reader meeting both will assume one is a synonym of the other. They are not,
and the difference is worth the paragraph.

`group.id` is a property of the record. It says who made this measurement, and
every record has exactly one.

`correlation_group`, from `docs/decisions/0005-uncertainty.md`, is a property of
one uncertainty component. It names one shared systematic, from the same
apparatus or the same input constant, appearing in more than one record.

Neither implies the other in either direction. Two records from different groups
share a `correlation_group` when both imported the same constant. Two records
from one group carry no `correlation_group` at all when nobody wrote down what
they shared, and that absence means no shared systematic was recorded rather
than that none exists, which `docs/decisions/0005-uncertainty.md` states in its
own words.

An analysis that reads one and reports it as the other is measuring something
it did not name.

## What checks any of this

Nothing.

No check in this repository reads this file, resolves a `group.id` against this
directory, refuses a duplicate entry, refuses an identifier that breaks the
syntax above, or refuses an entry naming a person. The validator that would do
the resolving is owed by #25, and there is no validator, no module and no
source in this repository today:

    git ls-files | grep -E '\.go$|go\.mod'
    (no output, exit 1)

So every rule on this page is held by a person reading it. Saying so is not a
formality: a curator who believes a machine is watching the identifier syntax
stops watching it, and the first duplicate group entry is invisible until an
analysis pools two halves of one experiment as though they were two.
