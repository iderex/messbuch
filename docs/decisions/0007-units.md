# 0007 Units, normalization and redefinitions

## Question

A measurement has to be stored as it was published, because that is the fact,
and it has to be comparable to every other measurement of the same quantity,
because that is the point of the corpus. Those are two different numbers. How
are both stored, what converts one into the other, and what happens when the
conversion is not a constant factor?

## Options considered

Store the published value only and convert at analysis time.

Store a normalized value only and keep the published form as free text.

Store both, with the conversion recorded on the record.

Store both, with the conversion recorded centrally in a conversion table the
records refer to by name.

## What each option costs

Published only, converted at analysis time. Nothing is lost from the source and
there is one number per record. It costs reproducibility in the way this
project cares about most: the conversion factor becomes part of the tool rather
than part of the data, so upgrading the tool changes every number computed from
the corpus, silently, and a result published last year cannot be reproduced
from this year's tool even against the same corpus revision. It also puts a
per-record judgement, such as which definitional regime a 1970 value was
published under, inside code that only sees a number and a unit string.

Normalized only, published as text. Every record is directly comparable and the
analysis is simple. It costs the fact. The published number is what a reader
checks against the paper, and once it is text nothing validates it, nothing
sums it, and a transcription error in it is invisible. It also makes the
corpus's own conversions unfalsifiable, since there is no stored input to
recompute them from.

Both, conversion on the record. The published fact and the comparable number
are both first-class, the factor that relates them is visible in the diff of
the contribution that added it, and a later change to a recommended constant
does not move a number that has already been cited. The cost is duplication:
two numbers that can disagree, a per-record factor that a contributor has to
find and enter, and a validator obligation to check that the two are consistent
with the stored factor.

Both, conversion in a central table. Same benefits, less duplication, and one
place to fix a wrong factor. The cost is the same one the first option has, one
step smaller: the stored normalized values now depend on a file that can
change, so correcting the table silently changes normalized values across the
corpus without any record's own history showing it. That is the failure the
correction path is supposed to make visible, and a central table routes around
it.

## Choice

Both values are stored on the record, and the conversion that relates them is
stored on the record with them. Nothing is derived on the fly at analysis time.

### The canonical unit

The canonical unit for a dimension is the coherent SI unit for that dimension,
written without a prefix. That is a rule rather than a list, so it answers for
a dimension the corpus has not met yet, and a list of units in a document
drifts against the vocabulary that actually assigns them. The vocabulary entry
for a quantity is the authority for that quantity's dimension and canonical
unit; this record is the authority for how that unit is chosen.

Applying the rule to the dimensions the seed corpus needs:

| Dimension | Canonical unit |
| --- | --- |
| dimensionless | `1` |
| length | `m` |
| mass | `kg` |
| time | `s` |
| electric current | `A` |
| thermodynamic temperature | `K` |
| amount of substance | `mol` |
| luminous intensity | `cd` |
| energy | `J` |
| frequency | `Hz` |
| electric charge | `C` |
| magnetic flux density | `T` |

No prefixes and no field-preferred units in the canonical form. The electronvolt
is the unit most of the particle-physics literature reports energies in, and it
is not the canonical unit here. It is what the published value is stored in,
which is where it belongs. Choosing a canonical unit per discipline would put
the corpus's comparability at the mercy of which discipline a record came from,
which is the one thing a cross-disciplinary corpus cannot afford.

### The published value and the normalized value

Every record carries a published block and a normalized block.

The published block holds the value and its uncertainty components exactly as
the source printed them, in the unit the source used. It is the fact. Nothing
in the corpus may rewrite it.

The normalized block holds the same measurement expressed in the canonical
unit, with its own uncertainty components. It is computed once, at
transcription time, and committed. The relationship is a stored multiplication:

    normalized.value = published.value * conversion.factor

The factor is stored on the record together with what it came from. The
normalized value is stored with enough significant digits that dividing it by
the stored factor recovers the published value to the precision the source
printed. A validator can therefore check the arithmetic of every record without
knowing anything about units, and it can do it from the record alone.

Where the published unit is already the canonical unit, the factor is exactly
`1` and it is still written. A field that is present on some records and absent
on others is a field every consumer has to branch on.

### Conversion uncertainty

Some conversions run through a measured constant, and a converted number that
carries only the original uncertainty claims a precision nobody measured.

A conversion is exact when the factor is fixed by definition. Since the 2019
revision of the SI, the conversions that run through the elementary charge, the
Planck constant, the Boltzmann constant, the Avogadro constant and the speed of
light are in this class, because those constants have exact defined values.
Conversions between an SI unit and a unit defined as an exact multiple of it,
such as the inch, are also exact. An exact conversion imports no uncertainty
and the record says `conversion.exact = true`.

A conversion is inexact when the factor is a measured quantity with a published
uncertainty of its own. For those, the record stores the factor's relative
uncertainty alongside the factor, and the normalized block carries an
additional uncertainty component with `component = "other"` and
`label = "unit-conversion"`. That component appears in the normalized block
only. It never appears in the published block, because the source did not quote
it and the published block is the fact.

The negligible case is recorded rather than dropped. The imported term is
treated as negligible when its relative size is at most one hundredth of the
smallest relative uncertainty the measurement itself quotes. The reason is
arithmetic and can be checked: two terms combined in quadrature, where one is
one hundredth of the other, give a total larger by a factor of
`sqrt(1 + 0.01^2)`, which is `1.00005`, a change of five parts in a hundred
thousand. That is below the precision at which any of this literature prints
its uncertainties, so carrying the term would be recording a difference nobody
can observe.

    python -c "import math; print(math.sqrt(1 + 0.01**2))"
    1.0000499987500624

When the term is negligible, the record still stores the factor's relative
uncertainty and sets `conversion.imported_uncertainty = "negligible"`, so the
judgement carries the two numbers that produced it and a later reader can
recheck it rather than trust it. Silently omitting a negligible term and
silently omitting a large one look identical on disk, and that is the state
this rule exists to prevent.

### Values published under a superseded definition

No value is converted across a redefinition.

Some units were redefined in ways that change what the number means. The metre
was fixed to the speed of light in 1983, which turned a measured quantity into
a defined one. The kilogram was redefined in 2019. A value published before a
redefinition, in the units of its day, is not the same number as the same
measurement expressed today, and a corpus that quietly converts across a
redefinition manufactures a trend or erases one. Either outcome is fatal here,
because a manufactured trend is exactly the signal this corpus was assembled to
look for.

So every record carries `definition_epoch`, a string naming the definitional
regime the source published under, taken from a closed set the vocabulary
maintains per dimension. Normalization uses the conversion factors in force at
the publication date of the source, not today's.

Where no defensible conversion across the regimes exists, the normalized block
is absent and the record carries
`normalization_status = "not-convertible-across-redefinition"` with a note
saying which regimes are involved. Such a record is a valid record. It is
excluded by default from any analysis that would compare it against records
from another regime, it is counted in the excluded set, and the count is
printed with the result. An analysis may include it, and doing so requires an
explicit option that appears in the output stamp.

Where the redefinition turned the quantity from measured into defined, the
vocabulary entry says so with the date, and measurements published after that
date are not measurements of the same thing. That is a vocabulary question
about quantity identity and this record does not settle it; it only requires
that the vocabulary answer it.

### Dimensionless quantities and ratios

For a dimensionless quantity the canonical unit is `1` and the question is not
conversion but what the denominator was. The record carries
`denominator_quantity`, naming the vocabulary entry of the quantity the value
is expressed relative to, or the literal `definitional` where the
dimensionlessness comes from the definition of the quantity itself rather than
from a division by another measured quantity. A dimensionless value with
neither is not comparable to anything and is refused.

Where the denominator is itself a measured quantity, the conversion rules above
apply to it: expressing the ratio against a different denominator is an inexact
conversion and imports that denominator's uncertainty.

### The source of the conversion factors

Exact factors and unit definitions come from the SI Brochure published by the
BIPM, and the edition is named on the record.

Inexact factors come from a CODATA recommended-values adjustment, and the
adjustment is named on the record by year, for example `codata-2018`.

The factor's numeric value is stored on the record. Naming the source is not
enough on its own, because a later CODATA adjustment would otherwise change
every normalized value in the corpus without any record's history showing it,
and a number somebody cited last year would quietly become a different number.
Storing the value pins it. Updating to a newer adjustment is then a visible
change to specific records, which is what the correction path is for.

## Reasons

Both numbers are stored because they answer different questions and neither can
be reconstructed from the other without carrying something extra. The published
value cannot be recovered from the normalized one without the factor, and the
normalized one cannot be recovered from the published one without a judgement
about which definitional regime applies. Once the factor and the regime have to
be on the record anyway, storing the second number costs one field and buys a
validator that can check the arithmetic from the record alone.

The conversion is on the record rather than in a central table for the reason
the central table looks best: one place to fix a wrong factor. That is the same
mechanism as one place to silently change a thousand cited numbers. This
project's correction path exists so that a changed number is visible as a
changed number, and a shared table would route the most likely large correction
around it. The duplication cost is paid, and it is what the consistency check
on `normalized = published * factor` is for.

The refusal to convert across a redefinition is the strictest rule here and it
is the one with the least room for judgement. Every alternative amounts to the
corpus inventing a number, and the corpus exists because inventing numbers is
what it is trying to detect.

Nothing in this repository refuses any of this. `PROSE, NOT ENFORCEMENT`. There
is no schema, no vocabulary file and no validator here today, so a record whose
normalized value does not equal its published value times its stored factor
passes every route in this tree. The arithmetic check described above is a
check that is owed, on the meaning leg of the validator, and no line of it
exists yet.

## Date

2026-08-07

## Reversal condition

Reverse the negligible-term threshold if a series appears whose own quoted
uncertainties are precise enough that one part in a hundred thousand is inside
what the source prints. The threshold is a claim about the precision of this
literature and it stops holding the moment the literature gets more precise
than it.

Reverse the stored-factor rule in favour of a central table if the corpus ever
gains a mechanism that makes a table change visible per record, meaning that
regenerating normalized values from a new adjustment produces a per-record
correction entry with its own history. At that point the objection above is
answered and the duplication is no longer buying anything.

Revisit the canonical-unit rule if a dimension appears for which no coherent SI
unit exists or for which the coherent SI unit is so far from the literature's
practice that every stored normalized value is a number no source ever printed
and no reader recognises. The rule survives inconvenience; it does not survive
being useless.
