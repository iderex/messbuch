// Package net reaches the network through a helper rather than directly, which
// is the shape docs/decisions/0009-offline-by-default.md permits when it says
// the permitted set is the package and its own dependencies.
package net

import "messbuch.example/fixture/internal/transport"

// Name is what the helper reaches.
var Name = transport.Name
