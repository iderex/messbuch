# 0005 How uncertainty is represented

## Question

How does a record store what a publication said about the uncertainty of its
measurement, given that the whole reason this corpus exists is a claim about
what those quoted intervals turn out to mean?

The constraint that decides this record: a representation that flattens
uncertainty to one number makes the finding the corpus was assembled to test
unmeasurable on the corpus. Getting this wrong is not a formatting mistake, it
is the loss of the dataset's reason to exist.

## Options considered

A single number per measurement, being whatever the paper printed after the
plus-minus sign.

A single number plus a free-text note holding everything else.

A structured list of named components, each with its own interval and its own
coverage statement, with limits and absence carried as first-class states.

A structured list as above, but with the components combined on the way in and
the parts kept only for reference.

## What each option costs

A single number. Trivial to write, trivial to parse, and every existing tool
already expects it. It costs the project its subject. A statistical and a
systematic component added in quadrature on the way in destroys the paper's own
combination choice, which is itself a thing worth studying; an asymmetric
interval becomes symmetric and the asymmetry is unrecoverable; a 90 per cent
confidence limit becomes indistinguishable from a one-standard-deviation
interval, which is a factor of about one and two thirds applied silently in an
unknown direction; and an upper limit with no central value cannot be stored at
all, so those measurements leave the corpus without anybody deciding that they
should.

A single number plus a note. Cheap, and nothing is lost from the human record.
It costs everything at analysis time, because a note is not readable by the
estimators. The observable outcome is a corpus that looks complete and produces
wrong pooled numbers, since the analysis silently treats intervals of different
coverage as the same kind of thing while the note explaining otherwise sits one
field away and is never read.

A structured list of components. It carries every case the literature actually
produces without a special case for any of them. It costs the contributor real
effort per record, it costs the schema and the validator complexity, and it
makes the corpus harder to consume than a two-column table, so some consumers
will flatten it themselves and get the flattening wrong in ways this project
cannot control.

A structured list combined on the way in. Slightly easier to consume and it
looks like it keeps everything. It costs the same thing the first option costs,
one step further back: the combined number becomes the number people use, the
combination rule becomes an assumption baked into the corpus rather than a
choice made by an analysis that states it, and the paper's own combination
stops being distinguishable from ours.

## Choice

Uncertainty is a list of named components with their own intervals and their
own coverage statements. Nothing is combined on the way in. Absence, zero and
a limit are three different states and are stored as three different things.

The representation, written in the tracked TOML format fixed by
`docs/decisions/0003-storage-format.md`. Field names below are the ones this
record proposes; the schema record is what fixes them, and where the two
disagree the schema record wins and this one is amended.

```toml
[measurement]
value = 6.6743e-11          # absent when the record is a limit
unit = "m^3 kg^-1 s^-2"
uncertainty_status = "reported"

[[measurement.uncertainty]]
component = "statistical"
plus = 1.5e-15
minus = 1.5e-15
coverage = "k=1"
```

`uncertainty_status` is required on every record and takes exactly one of three
values.

- `reported`. The source printed an uncertainty. At least one component block
  is present.
- `none-in-source`. The source printed a value and no uncertainty at all. No
  component block is present. This is a fact about the publication and it
  happens throughout the older literature.
- `not-transcribed`. Nobody has read the uncertainty out of the source yet. No
  component block is present. The record is not analysis-ready and any analysis
  that reads it says so.

`none-in-source` and `not-transcribed` are the same on disk except for this
field, and that is the point. Without it, a record nobody has finished and a
record whose source genuinely quoted nothing are the same file, and the corpus
cannot tell a gap in the literature from a gap in our work.

Each component block carries:

- `component`, one of `statistical`, `systematic`, `total`, or `other`. `other`
  requires a `label` field holding the name the source used. `total` means the
  source quoted one combined number and did not break it down, which is a
  different statement from the source quoting only a statistical component.
- `plus` and `minus`, both non-negative, both in the unit of `value`, both
  always written. A symmetric interval is written with `plus` equal to `minus`
  rather than with a single field, so an asymmetric interval is the ordinary
  case and not a special one, and so a parser never has to guess which shape it
  is reading.
- `coverage`, taking one of `k=<n>` for a stated multiple of a standard
  deviation, `cl=<percent>` for a stated confidence or credible level,
  `unstated` when the source gave an interval and did not say how to read it,
  or `other` with a `coverage_note` for the conventions that are neither. Also
  `interval_kind`, one of `frequentist`, `bayesian` or `unstated`, because a 90
  per cent credible interval and a 90 per cent confidence interval are not the
  same object and the difference is invisible in the number.

Three further shapes, each on the component or on the measurement.

A limit with no central value. `value` is absent, `uncertainty_status` is
`reported`, and the measurement carries a `[measurement.limit]` block with
`direction` of `upper` or `lower`, `bound`, and its own `coverage` and
`interval_kind`. A limit is not an interval around a value and is not stored as
one.

A systematic shared between measurements. A component may carry
`correlation_group`, a string. Two components in two different records naming
the same group are the same systematic, from the same apparatus or the same
input constant, appearing twice. The group is a name and not a correlation
matrix, because what the literature actually tells you is that two experiments
shared an apparatus, and a matrix would be an invented number dressed as a
measured one.

An uncertainty the authors rescaled. `plus` and `minus` hold the value as
published, which for a rescaled uncertainty is the rescaled one, and the
component carries a `[.. .rescaling]` block with `original_plus`,
`original_minus`, `factor` where the source stated one, `by` taking
`authors` or `third-party`, and `note`. Both states are on the record. Which
one an analysis uses is the analysis's decision and it prints which.

## The seven cases, written out

The blocks below show each required case in the representation. Their numbers
are invented for the purpose of showing the shape and are not transcribed from
any publication. See the debt named at the end of this section.

1. Separate statistical and systematic components, kept separate.

```toml
uncertainty_status = "reported"

[[measurement.uncertainty]]
component = "statistical"
plus = 4.1e-10
minus = 4.1e-10
coverage = "k=1"
interval_kind = "frequentist"

[[measurement.uncertainty]]
component = "systematic"
plus = 2.9e-10
minus = 2.9e-10
coverage = "k=1"
interval_kind = "frequentist"
```

2. An asymmetric interval.

```toml
[[measurement.uncertainty]]
component = "total"
plus = 0.31
minus = 0.24
coverage = "k=1"
interval_kind = "frequentist"
```

3. A coverage convention the source did not state.

```toml
[[measurement.uncertainty]]
component = "total"
plus = 0.05
minus = 0.05
coverage = "unstated"
interval_kind = "unstated"
```

4. An upper limit with no central value.

```toml
[measurement]
unit = "eV"
uncertainty_status = "reported"

[measurement.limit]
direction = "upper"
bound = 0.8
coverage = "cl=90"
interval_kind = "frequentist"
```

5. A measurement published with no uncertainty at all.

```toml
[measurement]
value = 4.774e-10
unit = "esu"
uncertainty_status = "none-in-source"
```

6. A systematic shared between two measurements from one apparatus. Both
   records carry a component naming the same group.

```toml
[[measurement.uncertainty]]
component = "systematic"
plus = 0.7
minus = 0.7
coverage = "k=1"
interval_kind = "frequentist"
correlation_group = "torsion-balance-fibre-anelasticity-lab-x"
```

7. An uncertainty the authors themselves rescaled.

```toml
[[measurement.uncertainty]]
component = "total"
plus = 0.30
minus = 0.30
coverage = "k=1"
interval_kind = "frequentist"

[measurement.uncertainty.rescaling]
original_plus = 0.10
original_minus = 0.10
factor = 3.0
by = "authors"
note = "authors inflated the quoted interval in the erratum"
```

THE REAL PUBLISHED EXAMPLE PER CASE IS OWED AND IS NOT IN THIS RECORD. The
issue that asks for this record asks for a real published example of each of
the seven cases, and every block above is invented. Quoting a published number
here without having read it out of the source would be exactly the defect this
corpus is about, and the sources cannot be read on the route that produced this
record. The examples are attached when the seed series are transcribed from
their primary sources under the seed-corpus milestone, and the issue is not
closed until they are. This paragraph is the whole disclosure and nothing later
in this record softens it.

## What an analysis may assume when a field is absent

Stated per field, because the defaults are where a representation quietly
becomes an assumption.

`uncertainty_status = "none-in-source"`. An analysis may not treat the
uncertainty as zero. Zero uncertainty gives the measurement infinite weight in
any inverse-variance scheme, which turns a missing datum into the dominant one.
The default is to exclude such a record from any weighted estimator, count it
in the excluded set, and print the count with the result. Including them under
an imputed uncertainty is permitted, requires an explicit option, and that
option is named in the output stamp.

`uncertainty_status = "not-transcribed"`. Excluded from every analysis, counted
separately from `none-in-source`, and printed separately. Merging the two
counts would hide our own incompleteness inside a statement about the
literature.

`coverage = "unstated"`. An analysis may not assume one standard deviation.
That assumption is the single most common way a meta-analysis of historical
data goes wrong, and this corpus is not entitled to make it on the reader's
behalf. The default is exclusion with a printed count; assuming a coverage
requires an explicit option and the assumed value appears in the stamp.

`interval_kind = "unstated"`. Usable, and the analysis records that the mixture
contains intervals of unknown kind. This is weaker than the coverage case
because the numerical effect is usually small, and the honest statement is that
it is usually small rather than that it is absent.

A component absent entirely, for example no `systematic` block on a record
whose status is `reported`. This means the source did not report that component
separately. It does not mean the systematic was zero. An analysis that needs a
systematic and finds none uses the `total` if one is present, and where it is
not, treats the record as having no separable systematic and says so in the
split by whether a systematic was quoted, which is one of the splits the
distribution analysis is required to produce.

`correlation_group` absent. This means no shared systematic is recorded. It
does not mean the measurements are independent. The literature rarely says, and
a corpus that reports absence of evidence as evidence of independence would be
overstating precision in a known direction, which is the exact error this
project is about. An analysis assuming independence states that it assumed it.

`rescaling` absent. The quoted interval is as published and was not rescaled as
far as the transcription found. It does not mean nobody rescaled it.

A zero written into `plus` or `minus` is a fact about the source. It means the
source printed a zero. Any analysis meeting a zero-width component treats it as
a refusal case rather than as a weight, because there is no arithmetic in which
that number is usable, and reports it as a data defect to be checked against
the source.

## Reasons

The component list wins because every rejected option loses the same thing in a
different place, and what it loses is the corpus's subject. The costs of the
chosen option are real and are paid deliberately. The contributor effort is
paid because the alternative is a corpus that is cheaper to build and cannot
answer the question. The consumer difficulty is paid, and it is mitigated by
the built artifact rather than by simplifying the tracked source, which is the
separation the storage record already argues for.

The three-state `uncertainty_status` field is the part of this record most
likely to be seen as overhead, and it is the part that would be missed most.
Without it there is no way, from the corpus alone, to distinguish a gap in the
literature from unfinished transcription work, and every count of how many
historical measurements quoted no uncertainty would silently include our own
backlog.

`unstated` is a stored value rather than an assumed default because the
assumption is not neutral. Reading an unstated interval as one standard
deviation and reading it as a 90 per cent confidence interval differ by roughly
a factor of one and two thirds in the width, which moves every deviation
statistic computed from it, in the direction that makes the corpus look better
behaved than it is. A corpus assembled to test whether quoted intervals are
trustworthy may not fill in the missing ones with a convenient convention.

Nothing in this repository refuses a record that violates any of this.
`PROSE, NOT ENFORCEMENT`. There is no schema file and no validator here today;
the structural and meaning legs of the validator are open on the corpus
milestone. Every rule above is checked by a reviewer until they land, and the
analysis-time defaults are descriptions of code that does not exist yet rather
than of behaviour anybody has observed.

## Date

2026-08-07

## Reversal condition

Reverse the closed set of `component` values if `other` with a free label
becomes the common case rather than the exception, because at that point the
closed set is a fiction and the labels are the real vocabulary. Measure it
rather than guess: the trigger is the share of components carrying `other`
across the corpus, and it is a count the validator can print.

Reverse `correlation_group` toward an explicit covariance representation if the
sources start giving correlations as numbers often enough that recording them
as a name discards real information. Today they mostly do not, and inventing a
matrix from an apparatus description would be manufacturing a measurement.

Revisit the whole record if the schema decision or a later analysis finds a
case in the literature that none of the seven shapes above carries. That is the
expected way this record changes, and a new case is a reason to extend the
representation rather than a reason to force the case into an existing shape.
