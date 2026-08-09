// Package transport is reached only from the permitted package.
package transport

import "net/http"

// Name is what net reads.
var Name = http.MethodGet
