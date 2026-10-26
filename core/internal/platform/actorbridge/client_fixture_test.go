package actorbridge

import (
	"bytes"
	"crypto/ed25519"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/google/uuid"
)

func runtimeTargets(t *testing.T) map[actorassertionv1.Domain]Target {
	t.Helper()
	targets := make(map[actorassertionv1.Domain]Target, len(exactTargetDomains))
	for _, domain := range exactTargetDomains {
		host := "v1-" + string(domain) + ".fixture.invalid"
		target, err := NewTarget(domain, "https://"+host, host)
		if err != nil {
			t.Fatal(err)
		}
		targets[domain] = target
	}
	return targets
}

func runtimeCredential(t *testing.T, now time.Time) actorassertionv1.Credential {
	t.Helper()
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{22}, ed25519.SeedSize))
	signer, err := actorassertionv1.NewSigner(
		"edge-2026-07", privateKey, bytes.NewReader(bytes.Repeat([]byte{7}, 16)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	target, _ := actorassertionv1.TargetBindingSHA256(
		"POST", "/api/v1/forms/records", "record-17",
	)
	credential, err := signer.Issue(actorassertionv1.DecisionEvidence{
		TenantID:     uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		SubjectID:    uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		PermissionID: uuid.MustParse("30000000-0000-4000-8000-000000000003"),
		TargetDomain: actorassertionv1.DomainForms,
		ResourceKey:  "forms.records", ActionKey: "create",
		TargetBindingSHA256: target, IAMGroupGeneration: 11,
		IAMGeneration: 12, TargetGeneration: 13, TokenVersion: 7,
		AuthorizationVersion: 19, AuthorizationClock: 23,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	return credential
}
