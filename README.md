# messbuch

Historical measurement series show bandwagon effects, and the deviations between experiments follow an exponential rather than a normal distribution out to ten standard deviations, which means the plus-minus in the literature does not mean what everyone assumes. No open, machine-readable, cross-disciplinary dataset of published measurements with date, uncertainty and method exists; the PDG holds one internally for particle physics and has run meta-analyses since 1957 with essentially no exchange with meta-analysis outside the field. Two deliverables in order: the corpus, then a meta-analysis tool importing the medical tradition, random-effects models and funnel plots and publication-bias diagnostics, rather than home-built scale factors.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is written down there with its reasons before the code
that depends on it exists.

See [NOTICE.md](NOTICE.md) for the intended-use notice.

## License

AGPL-3.0, decided by the maintainer on 2026-08-08. It answers entry 1 of
issue #13 and no other entry in that issue.

The full text is in [LICENSE](LICENSE). Read that file rather than this line,
and if you want the platform's own reading of it, run:

    gh api repos/iderex/messbuch --jq '.license.spdx_id'

Whether the corpus under `record/` and `vocabulary/` is to carry terms of its
own is entry 2 of issue #13, and that entry is not answered here.
