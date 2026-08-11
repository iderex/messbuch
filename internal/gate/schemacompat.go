package gate

import (
	"fmt"
	"strings"

	"github.com/iderex/messbuch/internal/compat"
)

// schemaCompatibilityLeg refuses a schema change that makes an existing corpus
// unreadable, or readable and differently interpreted.
//
// The second half is the one worth having. A corpus that stops validating is
// loud, and whoever made the change sees it immediately. A corpus that still
// validates and now means something else is silent, and the person who finds it
// is holding a number that moved without anything saying so.
//
// What it compares is a reading rather than a verdict, and internal/compat is
// where that choice is argued. It fails closed: no frozen corpus at all, a
// frozen reading that cannot be parsed, and a frozen file that has left the
// corpus are refusals rather than a pass over what happens to be left.
func schemaCompatibilityLeg(root string) (string, error) {
	report, err := compat.Check(root)
	if err != nil {
		return "", err
	}

	if len(report.Differences) > 0 {
		return "", fmt.Errorf("%d frozen file(s) read differently now:\n  %s\n\nA break made on purpose is a new schema version, a frozen corpus at the old one and a line in the changelog. This refuses the break nobody meant.",
			len(report.Differences), strings.Join(report.Differences, "\n  "))
	}

	var parts []string
	total := 0
	for _, version := range report.Versions {
		parts = append(parts, fmt.Sprintf("%d file(s) at schema version %d", len(version.Files), version.Schema))
		total += len(version.Files)
	}
	return fmt.Sprintf("%d frozen file(s) read as they did when frozen: %s", total, strings.Join(parts, ", ")), nil
}

// limitsOfSchemaCompatibility says where the leg stops, printed beside its
// result.
const limitsOfSchemaCompatibility = `this compares what this tree makes of a frozen file against what it made of the
same bytes when the file was frozen, so it refuses a change in interpretation
and never says the interpretation was right. The frozen readings were produced
by this code, deliberately: what they pin is that nothing moved. It reaches only
the schema versions with a directory under internal/compat/testdata, and no
version of this schema has been released, so what is frozen is what the tree
carries rather than what anybody has shipped. A refusal's wording is not frozen,
only its site and its field, so rewording a message is not a compatibility
break.`
