// Package analysis imports no network-capable path itself. It reaches one two
// hops away, which is the case a check reading only direct imports would pass.
package analysis

import "messbuch.example/fixture/internal/report"

// Name names what this package reaches.
var Name = report.Name
