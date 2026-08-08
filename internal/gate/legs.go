package gate

// Legs returns the gate in the order it runs.
//
// The order is cheapest-first among the things that can invalidate everything
// after them. Checking the toolchain pin costs a file read and decides whether
// any later number means what it says. Verifying the module set costs no
// network and decides whether the code about to be built is the code somebody
// reviewed. Compiling comes before testing because a test result from a tree
// that does not build does not exist.
//
// A leg carrying Owed rather than Run is declared here on purpose. The
// alternative is a gate that silently covers less than a reader assumes, and
// the whole reason this repository has one command instead of a list in a
// document is that the covered set should come out of the run.
func Legs() []Leg {
	return []Leg{
		{
			ID:      "toolchain",
			Subject: "the running release against the pin in go.mod",
			Run:     toolchainLeg,
		},
		{
			ID:      "modules",
			Subject: "the dependency set against go.sum, in locked mode",
			Run:     modulesLeg,
		},
		{
			ID:      "build",
			Subject: "every package compiles",
			Run:     buildLeg,
		},
		{
			ID:      "test",
			Subject: "every package's tests pass",
			Run:     testLeg,
		},
		{
			ID:      "corpus-decodes",
			Subject: "every tracked TOML file parses",
			Run:     corpusDecodesLeg,
		},
		{
			ID:      "format-and-lint",
			Subject: "formatting and lint over the source",
			Owed: "not built. Issue #17 owes it, and it decides which linter\n" +
				"and which settings. Nothing in this run read the formatting of any file.",
		},
		{
			ID:      "validate-corpus",
			Subject: "each record against the schema",
			Owed: "not built. Issue #24 owes the structural leg and #25 the meaning leg,\n" +
				"both against the machine readable schema on #23. The corpus-decodes leg\n" +
				"above answers only whether a file parses, so an unknown field, a missing\n" +
				"required field, a value outside a closed set and a file in the wrong place\n" +
				"all pass this run.",
		},
		{
			ID:      "no-network-imports",
			Subject: "no package outside internal/net reaches a network-capable API",
			Owed: "not built. Issue #65 owes it, under the name that\n" +
				"docs/decisions/0009-offline-by-default.md fixes for the check. Nothing in\n" +
				"this run computed an import graph, so the offline guarantee is an intention\n" +
				"here and not a measured property.",
		},
	}
}
