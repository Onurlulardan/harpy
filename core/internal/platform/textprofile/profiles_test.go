package textprofile

import "testing"

func TestProfilesHaveDeterministicUnicodeSemantics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		normalize func(string) string
		input     string
		want      string
	}{
		{"exact NFC", ExactNFC, "Cafe\u0301", "Caf\u00e9"},
		{"default fold sharp s", UnicodeDefaultCasefoldV1, "Stra\u00dfe", "strasse"},
		{"default fold Turkish I", UnicodeDefaultCasefoldV1, "I\u0130\u0131i", "ii\u0307\u0131i"},
		{"accent profile", AccentInsensitiveV1, "CAF\u00c9", "cafe"},
		{"all mark classes", AccentInsensitiveV1, "a\u0301a\u0903a\u20dd", "aaa"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.normalize(test.input); got != test.want {
				t.Fatalf("normalized value = %q, want %q", got, test.want)
			}
		})
	}
}

func TestProfilesAreIdempotent(t *testing.T) {
	t.Parallel()
	profiles := []func(string) string{
		ExactNFC, UnicodeDefaultCasefoldV1, AccentInsensitiveV1,
	}
	values := []string{"", "\x00", "Cafe\u0301", "Stra\u00dfe", "I\u0130\u0131i"}
	for _, profile := range profiles {
		for _, value := range values {
			first := profile(value)
			if second := profile(first); second != first {
				t.Fatalf("profile is not idempotent for %q: %q != %q", value, second, first)
			}
		}
	}
}

func TestProfilesLeaveNULAndEmptyPolicyToConsumers(t *testing.T) {
	t.Parallel()
	for _, profile := range []func(string) string{
		ExactNFC, UnicodeDefaultCasefoldV1, AccentInsensitiveV1,
	} {
		if profile("") != "" || profile("\x00") != "\x00" {
			t.Fatal("technical text profile applied business acceptance policy")
		}
	}
}
