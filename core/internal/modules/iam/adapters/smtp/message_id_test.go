package smtp

import "testing"

func TestStableMessageIDIsDeterministicAndDomainBound(t *testing.T) {
	client := &Client{settings: smtpSettingsFixture(t)}
	first, err := client.StableMessageID("iam.invitation.delivery:artifact-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.StableMessageID("iam.invitation.delivery:artifact-1")
	if err != nil {
		t.Fatal(err)
	}
	if first != second || !messageIDPattern.MatchString(first) ||
		first[len(first)-len("@example.invalid"):] != "@example.invalid" {
		t.Fatalf("stable message ID = %q/%q", first, second)
	}
}
