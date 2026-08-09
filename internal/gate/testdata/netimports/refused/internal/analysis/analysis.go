// Package analysis stands for any ordinary package of the tool. In this tree
// it carries the network-capable import, which is what the leg refuses.
package analysis

import "net/http"

// Client is what makes the import used. A fixture that does not compile is a
// fixture go list cannot read, and it would prove the failure path rather than
// the one it was written for.
var Client *http.Client
