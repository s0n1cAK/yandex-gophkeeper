package respond

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON_DisallowUnknownFields(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"a":1}`))

	var dst struct {
		B int `json:"b"`
	}

	if err := DecodeJSON(r, &dst); err == nil {
		t.Fatalf("expected error (unknown field)")
	}
}
