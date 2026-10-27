package identifier

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeEmailV1CanonicalizesExactProfile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input      string
		display    string
		normalized string
	}{
		{" User@Ex\u00e4mple.COM ", "User@Ex\u00e4mple.COM", "user@xn--exmple-cua.com"},
		{"Stra\u00dfe@Example.com", "Stra\u00dfe@Example.com", "strasse@example.com"},
		{"Cafe\u0301@Example.com", "Cafe\u0301@Example.com", "caf\u00e9@example.com"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			value, err := NormalizeEmailV1(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if !value.Valid() || value.Kind() != EmailV1 ||
				value.Display() != test.display || value.Normalized() != test.normalized {
				t.Fatalf("email value = %#v", value)
			}
		})
	}
}

func TestNormalizeEmailV1RejectsInvalidShapes(t *testing.T) {
	t.Parallel()
	longLocal := strings.Repeat("a", maximumEmailLocalBytes+1) + "@example.com"
	longLabel := "user@" + strings.Repeat("a", 64) + ".com"
	tests := []string{
		"", "user", "a@b@example.com", "@example.com", "user@",
		".user@example.com", "user.@example.com", "us..er@example.com",
		`"user"@example.com`, "user(comment)@example.com", "user@[127.0.0.1]",
		"us er@example.com", "user@exam ple.com", "user\x7f@example.com",
		"user@example.com.", string([]byte{0xff}) + "@example.com",
		longLocal, longLabel,
	}
	for _, input := range tests {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if value, err := NormalizeEmailV1(input); !errors.Is(err, ErrInvalid) ||
				value.Valid() {
				t.Fatalf("invalid email accepted: %q %#v %v", input, value, err)
			}
		})
	}
}

func TestNormalizeEmailV1IsIdempotent(t *testing.T) {
	t.Parallel()
	first, err := NormalizeEmailV1("Stra\u00dfe@B\u00dcCHER.example")
	if err != nil {
		t.Fatal(err)
	}
	second, err := NormalizeEmailV1(first.Normalized())
	if err != nil || second.Normalized() != first.Normalized() {
		t.Fatalf("second normalization = %#v, %v", second, err)
	}
}
