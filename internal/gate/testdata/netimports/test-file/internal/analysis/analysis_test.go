package analysis

import (
	"net/http"
	"testing"
)

func TestTheImportIsUsed(t *testing.T) {
	if http.MethodGet == "" {
		t.Error("the fixture's network-capable import is unused")
	}
}
