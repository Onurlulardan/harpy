package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

type lifecycleStateTenantFixture struct {
	artifact lifecycledeliveryapp.Artifact
}

func (fixture lifecycleStateTenantFixture) Claim(
	context.Context,
	lifecycledeliveryapp.ClaimCommand,
) (lifecycledeliveryapp.ClaimDecision, string, error) {
	return lifecycledeliveryapp.ClaimDecision{Artifact: &fixture.artifact},
		"harpy", nil
}

func (lifecycleStateTenantFixture) Release(
	context.Context, lifecycledeliveryapp.EffectReference,
) error {
	return nil
}

func (lifecycleStateTenantFixture) MarkUnknown(
	context.Context, lifecycledeliveryapp.EffectReference,
) error {
	return nil
}

func (lifecycleStateTenantFixture) Finalize(
	context.Context, lifecycledeliveryapp.FinalizeCommand,
) (lifecycledeliveryapp.StoredResult, error) {
	return lifecycledeliveryapp.StoredResult{}, lifecycledeliveryapp.ErrInvalid
}

type lifecycleCredentialFixture struct {
	token []byte
}

func (fixture lifecycleCredentialFixture) RevealOnce(reveal func([]byte) error) error {
	return reveal(fixture.token)
}

func (lifecycleCredentialFixture) Destroy() {}

type lifecycleOpenerFixture struct {
	credential lifecycleCredentialFixture
}

func (fixture lifecycleOpenerFixture) Open(
	context.Context, uuid.UUID, lifecycledeliveryv1.Envelope,
) (lifecycleCredential, error) {
	return fixture.credential, nil
}

type lifecycleSenderTenantFixture struct {
	message smtp.Message
}

func (*lifecycleSenderTenantFixture) StableMessageID(string) (string, error) {
	return "lifecycle.fixture@example.invalid", nil
}

func (fixture *lifecycleSenderTenantFixture) Send(
	_ context.Context,
	_ lifecycledeliveryv1.Purpose,
	message smtp.Message,
) smtp.Result {
	fixture.message = message
	return smtp.Result{Kind: smtp.ResultUnknown, Code: "SMTP_OUTCOME_UNKNOWN"}
}

func TestLifecycleDeliveryCarriesResolvedTenantSlugToSMTP(t *testing.T) {
	request, artifact, now := lifecycleDeliveryTenantFixture(t)
	sender := &lifecycleSenderTenantFixture{}
	provider := &lifecycleDeliveryProvider{
		state: lifecycleStateTenantFixture{artifact: artifact},
		opener: lifecycleOpenerFixture{credential: lifecycleCredentialFixture{
			token: []byte("abcdefghijklmnopqrstuv.ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopq"),
		}},
		smtp: sender,
		now:  func() time.Time { return now },
	}
	result := provider.Execute(context.Background(), request)
	if !result.IsUnknown() || sender.message.TenantSlug != "harpy" {
		t.Fatalf("result unknown=%t tenant slug=%q",
			result.IsUnknown(), sender.message.TenantSlug)
	}
}

func lifecycleDeliveryTenantFixture(
	t *testing.T,
) (effectsv1.Request, lifecycledeliveryapp.Artifact, time.Time) {
	t.Helper()
	tenantID, artifactID := uuid.New(), uuid.New()
	effectID, causationID := uuid.New(), uuid.New()
	reference, err := lifecycledeliveryv1.NewReference(tenantID, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	payload := reference.String()
	digest := sha256.Sum256([]byte("iam.lifecycle.delivery.v1\x00" + payload))
	tenant, now := tenantID.String(), time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	request, err := effectsv1.NewRequest(
		effectID.String(), LifecycleDeliveryJobKind, 1, &tenant,
		effectsv1.Authority{ScopeKind: effectsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "iam_system_resources", GroupGeneration: 1, Domain: "iam",
			CutoverGeneration: 1, WriterOrigin: effectsv1.WriterV2},
		"iam.smtp", "send_lifecycle_credential",
		"iam.lifecycle.delivery:"+payload, "job-1", causationID.String(),
		payload, hex.EncodeToString(digest[:]), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		t.Fatal(err)
	}
	return request, lifecycledeliveryapp.Artifact{
		Reference: payload, Recipient: "owner@example.invalid",
		Envelope: lifecycledeliveryv1.Envelope{
			ArtifactID: artifactID,
			Purpose:    lifecycledeliveryv1.PurposePasswordReset,
		},
		ExpiresAt: now.Add(time.Hour),
	}, now
}

var _ effectexecution.Provider = (*lifecycleDeliveryProvider)(nil)
