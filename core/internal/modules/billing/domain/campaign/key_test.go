package campaign

import "testing"

func TestCampaignKeyVocabulary(t *testing.T) {
	for _, value := range []string{"summer-sale", "launch.2026", "team_2"} {
		key, err := NewKey(value)
		if err != nil || !key.Valid() {
			t.Fatalf("valid key %q: %v", value, err)
		}
	}
	for _, value := range []string{"a", "Summer", "-sale", "sale.", "sale/now"} {
		if _, err := NewKey(value); err != ErrKey {
			t.Fatalf("invalid key %q accepted: %v", value, err)
		}
	}
}
