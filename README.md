# messbuch

Historical measurement series show bandwagon effects, and the deviations between experiments follow an exponential rather than a normal distribution out to ten standard deviations, which means the plus-minus in the literature does not mean what everyone assumes. No open, machine-readable, cross-disciplinary dataset of published measurements with date, uncertainty and method exists; the PDG holds one internally for particle physics and has run meta-analyses since 1957 with essentially no exchange with meta-analysis outside the field. Two deliverables in order: the corpus, then a meta-analysis tool importing the medical tradition, random-effects models and funnel plots and publication-bias diagnostics, rather than home-built scale factors.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is written down there with its reasons before the code
that depends on it exists.

See [NOTICE.md](NOTICE.md) for the intended-use notice.

## License

The code is AGPL-3.0, decided by the maintainer on 2026-08-08 and recorded in
[0016](docs/decisions/0016-repository-license.md).

The full text is in [LICENSE](LICENSE). Read that file rather than this line,
and if you want the platform's own reading of it, run:

    gh api repos/iderex/messbuch --jq '.license.spdx_id'

The corpus under `record/` and `vocabulary/` is a separate decision and is
CC BY 4.0, taken on 2026-08-09 and recorded in
[0017](docs/decisions/0017-corpus-license.md). What is licensed there is the
collection rather than any individual measured number, and it covers the
vocabulary and the transcriptions alike.

No file in this repository carries the CC BY 4.0 text. `LICENSE` is the AGPL-3.0
text and there is no second license file:

    git ls-files | grep -Ei 'LICENSE'
    LICENSE
