package strutil

import "testing"

func TestTruncateUTF8(t *testing.T) {
	if got := TruncateUTF8("hello", 10); got != "hello" {
		t.Fatalf("expected unchanged ascii string, got %q", got)
	}

	if got := TruncateUTF8("привет", 3); got != "при" {
		t.Fatalf("expected unicode-safe truncation, got %q", got)
	}
}
