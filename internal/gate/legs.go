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
			Limits:  limitsOfTheTestLeg,
		},
		{
			ID:      "headless-tests",
			Subject: "no test reaches a display, elevation, the network or a path outside this tree",
			Run:     headlessLeg,
			Limits:  limitsOfHeadless,
		},
		{
			ID:      "corpus-decodes",
			Subject: "every tracked TOML file parses",
			Run:     corpusDecodesLeg,
		},
		{
			ID:      "format-and-lint",
			Subject: "formatting and lint over the source and the tracked prose",
			Run:     formatAndLintLeg,
			Limits:  limitsOfFormatAndLint,
		},
		{
			ID:      "validate-corpus",
			Subject: "each record against the schema",
			Run:     validateCorpusLeg,
			Limits:  limitsOfValidateCorpus,
		},
		{
			ID:      "refusal-sites",
			Subject: "every place the validator can refuse is executed by a fixture",
			Run:     refusalSitesLeg,
			Limits:  limitsOfRefusalSites,
		},
		{
			ID:      "coverage-floor",
			Subject: "how much of the surface that decides refusals is executed at all",
			Run:     coverageFloorLeg,
			Limits:  limitsOfCoverageFloor,
		},
		{
			ID:      "no-network-imports",
			Subject: "no package outside internal/net reaches a network-capable API",
			Run:     noNetworkImportsLeg,
			Limits:  limitsOfNoNetworkImports,
		},
	}
}
