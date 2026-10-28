package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRandomBase64NeverStartsWithFlagDash(t *testing.T) {
	t.Parallel()
	for _, size := range []int{1, 12, 16, 32} {
		for attempt := 0; attempt < 256; attempt++ {
			value, err := randomBase64(size)
			if err != nil {
				t.Fatalf("randomBase64(%d): %v", size, err)
			}
			if strings.HasPrefix(value, "-") {
				t.Fatalf("randomBase64(%d) produced CLI-unsafe value %q", size, value)
			}
		}
	}
}

func TestRandomBase64KeepsRawURLEncodingShape(t *testing.T) {
	t.Parallel()
	for _, size := range []int{1, 12, 16, 32} {
		value, err := randomBase64(size)
		if err != nil {
			t.Fatalf("randomBase64(%d): %v", size, err)
		}
		if want := base64.RawURLEncoding.EncodedLen(size); len(value) != want {
			t.Fatalf("randomBase64(%d) length = %d, want %d", size, len(value), want)
		}
		decoded, err := base64.RawURLEncoding.DecodeString(value)
		if err != nil {
			t.Fatalf("randomBase64(%d) output is not raw URL base64: %v", size, err)
		}
		if len(decoded) != size {
			t.Fatalf("randomBase64(%d) decoded to %d bytes", size, len(decoded))
		}
	}
}
