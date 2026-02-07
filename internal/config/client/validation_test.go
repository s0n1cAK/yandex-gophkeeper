package client

import (
	"testing"
	"time"

	customtype "yandex-gophkeeper/internal/customType/duration"
)

func TestValidate_InsecureSkipVerify_OnlyLocalhostHTTPS(t *testing.T) {
	ok := Config{
		Address:            "https://localhost:8080",
		Timeout:            customtype.Duration(5 * time.Second),
		TokenFile:          "x",
		RetryMax:           0,
		RetryWaitMin:       customtype.Duration(0),
		RetryWaitMax:       customtype.Duration(0),
		InsecureSkipVerify: true,
	}
	if err := Validate(ok); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}

	bad1 := ok
	bad1.Address = "http://localhost:8080"
	if err := Validate(bad1); err == nil {
		t.Fatalf("expected error for http + insecure")
	}

	bad2 := ok
	bad2.Address = "https://example.com"
	if err := Validate(bad2); err == nil {
		t.Fatalf("expected error for non-localhost + insecure")
	}
}
