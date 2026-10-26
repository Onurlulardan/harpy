package domain

import "testing"

func TestColorUsesCanonicalLowercaseHex(t *testing.T) {
	if !ValidColor("#64748b") || ValidColor("#64748B") || ValidColor("blue") {
		t.Fatal("color contract drifted")
	}
}
