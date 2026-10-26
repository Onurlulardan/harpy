package cryptography

import (
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
)

func TestWebAuthnDiscoverableUsesCredentialProperties(t *testing.T) {
	for name, fixture := range map[string]struct {
		outputs protocol.AuthenticationExtensionsClientOutputs
		want    bool
	}{
		"absent": {outputs: nil, want: false},
		"false": {outputs: protocol.AuthenticationExtensionsClientOutputs{
			"credProps": map[string]any{"rk": false},
		}, want: false},
		"true": {outputs: protocol.AuthenticationExtensionsClientOutputs{
			"credProps": map[string]any{"rk": true},
		}, want: true},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := webAuthnDiscoverable(fixture.outputs)
			if err != nil || got != fixture.want {
				t.Fatalf("discoverable=%t want=%t err=%v", got, fixture.want, err)
			}
		})
	}
}

func TestWebAuthnDiscoverableRejectsMalformedCredentialProperties(t *testing.T) {
	for _, output := range []any{"true", map[string]any{"rk": "true"}} {
		_, err := webAuthnDiscoverable(
			protocol.AuthenticationExtensionsClientOutputs{"credProps": output},
		)
		if !errors.Is(err, ErrRegistrationDenied) {
			t.Fatalf("malformed credProps result=%v", err)
		}
	}
}
