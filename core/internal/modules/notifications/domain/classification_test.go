package domain

import "testing"

func TestClassificationIsClosedAndLocaleFallbackIsDeterministic(t *testing.T) {
	if Category("other").Valid() || Channel("websocket").Valid() {
		t.Fatal("unknown category or WebSocket channel was accepted")
	}
	locale, ok := ResolveLocale("de", []string{"en", "tr"})
	if !ok || locale != "tr" {
		t.Fatalf("fallback = %q, %t", locale, ok)
	}
	if _, ok = ResolveLocale("de", []string{"fr"}); ok {
		t.Fatal("unsupported locale was accepted")
	}
}
