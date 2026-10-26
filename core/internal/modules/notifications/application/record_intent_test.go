package application

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

func TestRecordIntentPersistsCanonicalStateAndReplaysExactEffect(t *testing.T) {
	application, stores, scope, request := newIntentFixture(t)
	first, err := application.RecordIntent(t.Context(), scope, request)
	if err != nil || first.Replay || first.InboxRevision != 1 {
		t.Fatalf("first RecordIntent() = %+v, %v", first, err)
	}
	if len(stores.messages) != 1 || len(stores.deliveries) != 1 ||
		len(stores.jobs) != 1 || len(stores.hints) != 1 || len(stores.audits) != 1 {
		t.Fatalf("persisted counts = messages:%d deliveries:%d jobs:%d hints:%d audits:%d",
			len(stores.messages), len(stores.deliveries), len(stores.jobs),
			len(stores.hints), len(stores.audits))
	}
	second, err := application.RecordIntent(t.Context(), scope, request)
	if err != nil || !second.Replay || second.MessageID != first.MessageID ||
		len(stores.jobs) != 1 {
		t.Fatalf("replay RecordIntent() = %+v, %v", second, err)
	}
}

func TestRecordIntentRejectsEffectKeyReuseWithDifferentPayload(t *testing.T) {
	application, _, scope, request := newIntentFixture(t)
	if _, err := application.RecordIntent(t.Context(), scope, request); err != nil {
		t.Fatal(err)
	}
	request.Data[0].String = "Farklı"
	request.DataDigest, _ = api.CanonicalDataDigest(request.Data)
	_, err := application.RecordIntent(t.Context(), scope, request)
	if !errors.Is(err, api.ErrEffectKeyReused) {
		t.Fatalf("reuse error = %v", err)
	}
}

func TestRecordIntentKeepsMandatorySecurityMessageWhenEmailUnreachable(t *testing.T) {
	application, stores, scope, request := newIntentFixture(t)
	request.Category = api.CategorySecurity
	request.Channels = []api.Channel{api.ChannelInApp}
	recipient := application.recipients.(testRecipients).value
	recipient.EmailVerified = false
	application.recipients = testRecipients{recipient}
	if _, err := application.RecordIntent(t.Context(), scope, request); err != nil {
		t.Fatal(err)
	}
	if len(stores.messages) != 1 || len(stores.deliveries) != 1 || len(stores.alerts) != 1 {
		t.Fatalf("mandatory failure counts = %d/%d/%d",
			len(stores.messages), len(stores.deliveries), len(stores.alerts))
	}
}
