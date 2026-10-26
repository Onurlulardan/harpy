package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

type testScope struct{ tenant, actor uuid.UUID }

func (value testScope) TenantID() uuid.UUID { return value.tenant }
func (value testScope) ActorID() uuid.UUID  { return value.actor }
func (testScope) OperationID() string       { return "notifications.intent.create" }

type testClock struct{ at time.Time }

func (value testClock) Now() time.Time { return value.at }

type testIDs struct{ next byte }

func (value *testIDs) New(at time.Time) (domain.ID, error) {
	value.next++
	return domain.NewV7(at.Add(time.Duration(value.next)*time.Nanosecond),
		bytes.NewReader(bytes.Repeat([]byte{value.next}, 10)))
}

type testAuthorizer struct{ denied bool }

func (value testAuthorizer) AuthorizeIntent(context.Context, api.Scope, string, string) error {
	if value.denied {
		return ErrForbidden
	}
	return nil
}

func (value testAuthorizer) AuthorizeOperation(context.Context, api.Scope, string) error {
	if value.denied {
		return ErrForbidden
	}
	return nil
}

type testRecipients struct{ value ports.Recipient }

func (value testRecipients) ResolveRecipient(context.Context, api.Scope,
	domain.ID,
) (ports.Recipient, error) {
	return value.value, nil
}

type testBinder struct{ stores *testStores }

func (value testBinder) Bind(context.Context, api.Scope) (ports.Stores, error) {
	return value.stores, nil
}

func newIntentFixture(t *testing.T) (*Application, *testStores, testScope, api.IntentRequest) {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	tenant, recipient := uuid.New(), uuid.New()
	actor := recipient
	tenantID, _ := domainID(tenant)
	recipientID, _ := domainID(recipient)
	templateID, _ := domainID(uuid.New())
	templateRoot, _ := domainID(uuid.New())
	providerID, _ := domainID(uuid.New())
	stores := newTestStores(domain.PublishedTemplate{ID: templateRoot,
		VersionID: templateID, NotificationType: "workflow.task.due",
		Kind: domain.TemplateSystemHTML, Locale: "tr", SubjectPattern: "Görev {{name}}",
		BodyPattern: "<p>{{name}}</p>", Variables: []string{"name"}, Revision: 1})
	stores.provider = ports.ProviderConfig{ID: providerID, VersionID: providerID,
		Kind: ports.ProviderSMTP, Enabled: true, Revision: 1}
	application, err := NewApplication(testBinder{stores}, testRecipients{ports.Recipient{
		TenantID: tenantID, PrincipalID: recipientID, SupportedLocales: []string{"tr", "en"},
		EmailVerified: true, AuthorizationVersion: 7}}, testAuthorizer{},
		testClock{now}, &testIDs{})
	if err != nil {
		t.Fatal(err)
	}
	data := []api.TemplateValue{{Key: "name", Kind: api.ValueString, String: "Bakım"}}
	digest, _ := api.CanonicalDataDigest(data)
	request := api.IntentRequest{Producer: "workflow", NotificationType: "workflow.task.due",
		RecipientPrincipalID: api.ID(recipient), TemplateVersionID: wireID(templateID),
		LocaleInput: "tr", EffectKey: "instance/node/1", Category: api.CategoryWorkflowAction,
		Channels: []api.Channel{api.ChannelInApp, api.ChannelEmail}, Data: data,
		DataDigest: digest, CorrelationID: api.ID(uuid.New()), CausationID: api.ID(uuid.New())}
	return application, stores, testScope{tenant, actor}, request
}

func testKey(producer string, hash [32]byte, recipient domain.ID) string {
	return fmt.Sprintf("%s:%x:%s", producer, hash, recipient.String())
}

func differentDigest(value [32]byte) [32]byte {
	return sha256.Sum256(append(value[:], 1))
}
