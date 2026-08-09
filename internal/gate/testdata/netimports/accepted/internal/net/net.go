// Package net is the one package permitted to reach a network-capable API, and
// in this tree it is where the import was moved to.
package net

import "net/http"

// Client is what makes the import used.
var Client *http.Client
