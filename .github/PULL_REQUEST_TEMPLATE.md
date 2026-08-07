<!--
Keep the two headings. Replace the text under each one. A pull request that
leaves either answer as it stands has not answered it.
-->

## What this changes, and what failure it prevents

REPLACE THIS. Say what changed and what goes wrong without it. Where a number
is asserted, paste the command that produced it, run at the commit being pushed
and against the reference the reviewer will have.

## Which documents does this change make wrong?

REPLACE THIS WITH A SENTENCE, NOT A TICK. `docs/downstream-documents.md` maps
which documents are downstream of which parts of the tree; read the rows that
cover what you touched and answer from them.

Three answers are acceptable and one is not.

- Naming the documents you fixed in this pull request.
- Naming the documents that are now wrong and the issue you opened on the
  current milestone before this merges, with its number.
- `None, and here is why`, followed by the why. Which rows of the map you
  checked is a good why. `None` on its own is not.

Leaving this heading as it is, is the answer that is not acceptable. It is a
question rather than a checkbox for exactly that reason: a box goes unticked by
default and reads the same whether it was considered or skipped.

Nothing machine checks this answer. No check in this repository reads this
template, compares it against the map, or refuses a pull request that left the
heading untouched, and none is claimed. The review is where a skipped answer is
caught. A rule that pretends to be enforced is worse than one that admits it is
not, and documentation debt is invisible until the person who finds it is a user
rather than a maintainer.
