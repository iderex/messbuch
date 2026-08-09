// A fixture tree for the refusal-site accounting. It is not part of any build
// and nothing here validates anything.
package validate

func newRefusal(site, detail string) string { return site + detail }

func one() string { return newRefusal("first-site", "reached") }

func two() string { return newRefusal("second-site", "unreached") }

func notASite() string { return newRefusal(other(), "") }

func other() string { return "computed" }
