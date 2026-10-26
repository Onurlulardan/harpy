package effectsv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRequestRoundTripIsStrictAndDeterministic(t *testing.T) {
	value := requestFixture(t)
	first, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("request marshal is not deterministic")
	}
	decoded, err := DecodeRequest(first)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.EffectID != value.EffectID || !equalOptional(decoded.TenantID, value.TenantID) {
		t.Fatal("request round trip changed identity")
	}

	unknown := append([]byte(`{"unknown":true,`), first[1:]...)
	if _, err := DecodeRequest(unknown); err == nil {
		t.Fatal("unknown request field was accepted")
	}
	duplicate := append([]byte(`{"contract_version":"effect-request.v1",`), first[1:]...)
	if _, err := DecodeRequest(duplicate); err == nil {
		t.Fatal("duplicate request field was accepted")
	}
	oversized := append(append([]byte(nil), first...), bytes.Repeat([]byte(" "), MaximumDocumentBytes)...)
	if _, err := DecodeRequest(oversized); err == nil {
		t.Fatal("oversized effect request document was accepted")
	}
	if _, err := DecodeRequest(append(first, []byte(` {}`)...)); err == nil {
		t.Fatal("trailing request document was accepted")
	}
	wrongVersion := bytes.Replace(first, []byte(RequestContractVersion), []byte("effect-request.v2"), 1)
	if _, err := DecodeRequest(wrongVersion); err == nil {
		t.Fatal("unknown request version was accepted")
	}
	var invalidUTF8 map[string]string
	if err := decodeStrict([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'}, &invalidUTF8); err == nil {
		t.Fatal("invalid UTF-8 document was accepted")
	}
}

func TestNewRequestDefensivelyCopiesTenant(t *testing.T) {
	tenantID := testTenantID
	value := requestFixture(t)
	value, err := NewRequest(
		value.EffectID, value.EffectKind, value.EffectRevision, &tenantID,
		value.Authority, value.Provider, value.Operation, value.IdempotencyKey,
		value.CorrelationID, value.CausationID,
		value.PayloadReference, value.PayloadSHA256, value.RequestedAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	tenantID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	if value.TenantID == nil || *value.TenantID != testTenantID {
		t.Fatal("constructor retained caller tenant pointer")
	}
	clone := value.TenantIDClone()
	*clone = tenantID
	if *value.TenantID != testTenantID {
		t.Fatal("tenant clone exposed request state")
	}
}

func TestRequestRejectsAuthorityAndEvidenceMismatch(t *testing.T) {
	tests := map[string]func(Request) Request{
		"tenant mismatch": func(value Request) Request {
			other := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
			value.TenantID = &other
			return value
		},
		"tenant scope without tenant": func(value Request) Request {
			value.TenantID = nil
			return value
		},
		"global scope with tenant": func(value Request) Request {
			value.Authority.ScopeKind = ScopeGlobal
			value.Authority.ScopeID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
			return value
		},
		"invalid group": func(value Request) Request {
			value.Authority.GroupID = "Notifications"
			return value
		},
		"invalid group generation": func(value Request) Request {
			value.Authority.GroupGeneration = 0
			return value
		},
		"invalid domain": func(value Request) Request {
			value.Authority.Domain = "Notifications"
			return value
		},
		"short effect kind": func(value Request) Request {
			value.EffectKind = "ab"
			return value
		},
		"invalid provider": func(value Request) Request {
			value.Provider = "Email Provider"
			return value
		},
		"invalid operation": func(value Request) Request {
			value.Operation = "send action"
			return value
		},
		"unknown writer": func(value Request) Request {
			value.Authority.WriterOrigin = "bridge"
			return value
		},
		"bad correlation": func(value Request) Request {
			value.CorrelationID = " request"
			return value
		},
		"bad causation": func(value Request) Request {
			value.CausationID = "00000000-0000-0000-0000-000000000000"
			return value
		},
		"bad digest": func(value Request) Request {
			value.PayloadSHA256 = strings.Repeat("A", 64)
			return value
		},
		"non canonical time": func(value Request) Request {
			value.RequestedAt = "2026-07-15T10:00:00+00:00"
			return value
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			if err := mutate(requestFixture(t)).Validate(); err == nil {
				t.Fatal("invalid request was accepted")
			}
		})
	}
}
