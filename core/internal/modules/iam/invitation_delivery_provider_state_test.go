package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/google/uuid"
)

type invitationDeliveryStateFixture struct {
	finalized []invitationdeliveryapp.FinalizeCommand
	released  int
	unknown   int
}

func (*invitationDeliveryStateFixture) Claim(
	context.Context,
	invitationdeliveryapp.ClaimCommand,
) (invitationdeliveryapp.ClaimDecision, string, error) {
	return invitationdeliveryapp.ClaimDecision{}, "",
		invitationdeliveryapp.ErrInvalid
}

func (state *invitationDeliveryStateFixture) Release(
	context.Context,
	invitationdeliveryapp.EffectReference,
) error {
	state.released++
	return nil
}

func (state *invitationDeliveryStateFixture) MarkUnknown(
	context.Context,
	invitationdeliveryapp.EffectReference,
) error {
	state.unknown++
	return nil
}

func (state *invitationDeliveryStateFixture) Finalize(
	_ context.Context,
	command invitationdeliveryapp.FinalizeCommand,
) (invitationdeliveryapp.StoredResult, error) {
	state.finalized = append(state.finalized, command)
	return command.Result, nil
}

func TestInvitationDeliveryAcceptedSMTPFinalizesReceipt(t *testing.T) {
	state := &invitationDeliveryStateFixture{}
	request, now := invitationDeliveryRequestFixture(t)
	provider := &invitationDeliveryProvider{state: state, now: func() time.Time { return now }}
	digest := sha256.Sum256([]byte("SMTP DATA accepted"))
	result := provider.handleSMTPResult(context.Background(), request, smtp.Result{
		Kind: smtp.ResultAccepted, Code: "SMTP_DATA_ACCEPTED",
		ProviderReference: request.messageID, EvidenceSHA256: digest, CompletedAt: now,
	})
	receipt, ok := result.Receipt()
	if !ok || receipt.Status != effectsv1.ReceiptSucceeded ||
		len(state.finalized) != 1 ||
		state.finalized[0].Result.Outcome != invitationdeliveryapp.OutcomeDelivered {
		t.Fatalf("accepted result = %#v, finalized = %#v", receipt, state.finalized)
	}
}

func TestInvitationDeliveryTransientSMTPReleasesClaim(t *testing.T) {
	state := &invitationDeliveryStateFixture{}
	request, now := invitationDeliveryRequestFixture(t)
	provider := &invitationDeliveryProvider{state: state, now: func() time.Time { return now }}
	result := provider.handleSMTPResult(context.Background(), request, smtp.Result{
		Kind: smtp.ResultTransient, Code: "SMTP_UNAVAILABLE",
		ProviderReference: request.messageID,
		EvidenceSHA256:    sha256.Sum256([]byte("dial failed")), CompletedAt: now,
	})
	receipt, ok := result.Receipt()
	if !ok || receipt.Status != effectsv1.ReceiptFailed || receipt.ErrorClass == nil ||
		*receipt.ErrorClass != effectsv1.ErrorTransient || state.released != 1 {
		t.Fatalf("transient result = %#v released=%d", receipt, state.released)
	}
}

func TestInvitationDeliveryAmbiguousSMTPMarksUnknown(t *testing.T) {
	state := &invitationDeliveryStateFixture{}
	request, now := invitationDeliveryRequestFixture(t)
	provider := &invitationDeliveryProvider{state: state, now: func() time.Time { return now }}
	result := provider.handleSMTPResult(context.Background(), request, smtp.Result{
		Kind: smtp.ResultUnknown, Code: "SMTP_OUTCOME_UNKNOWN",
		ProviderReference: request.messageID,
		EvidenceSHA256:    sha256.Sum256([]byte("DATA close ambiguous")), CompletedAt: now,
	})
	if !result.IsUnknown() || state.unknown != 1 || state.released != 0 {
		t.Fatalf("ambiguous result unknown=%t marked=%d released=%d",
			result.IsUnknown(), state.unknown, state.released)
	}
}

func invitationDeliveryRequestFixture(t *testing.T) (invitationRequest, time.Time) {
	t.Helper()
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	artifactID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	effectID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	causationID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	reference, err := deliveryartifactv1.NewReference(tenantID, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	tenant := tenantID.String()
	request, err := effectsv1.NewRequest(
		effectID.String(), InvitationDeliveryJobKind, 1, &tenant,
		effectsv1.Authority{ScopeKind: effectsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "iam_system_resources", GroupGeneration: 1, Domain: "iam",
			CutoverGeneration: 1, WriterOrigin: effectsv1.WriterV2},
		"iam.smtp", "send_invitation", "iam.invitation.delivery:"+reference.String(),
		"job-1", causationID.String(), reference.String(),
		hex.EncodeToString(sha256.New().Sum(nil)), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		t.Fatal(err)
	}
	return invitationRequest{
		request: request, tenantID: tenantID, requestedAt: now,
		messageID: "invite.fixture@example.invalid",
		reference: invitationdeliveryapp.EffectReference{
			TenantID: tenantID, ArtifactRef: reference.String(), EffectID: effectID,
			EffectKey: request.IdempotencyKey, OccurredAt: now,
		},
	}, now
}
