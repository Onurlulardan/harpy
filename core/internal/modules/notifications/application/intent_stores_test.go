package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type testStores struct {
	effects                map[string]domain.EffectReceipt
	messages               map[domain.ID]domain.Message
	deliveries             map[domain.ID]domain.Delivery
	preferences            map[string]domain.Preference
	template               domain.PublishedTemplate
	provider               ports.ProviderConfig
	providerSecrets        map[domain.ID]ports.EncryptedSecret
	smtpReceipts           map[string]ports.SMTPSettingsReceipt
	smtpTestIntents        map[string]ports.SMTPTestIntent
	latestSMTPRevision     int64
	jobs                   []ports.DeliveryJob
	hints                  []ports.RealtimeHint
	audits                 []string
	alerts                 []string
	revision               int64
	createdTemplate        ports.TemplateDraft
	createdTemplateVersion ports.TemplateVersion
}

func newTestStores(template domain.PublishedTemplate) *testStores {
	return &testStores{effects: make(map[string]domain.EffectReceipt),
		messages: make(map[domain.ID]domain.Message), deliveries: make(map[domain.ID]domain.Delivery),
		preferences: make(map[string]domain.Preference), template: template,
		providerSecrets: make(map[domain.ID]ports.EncryptedSecret),
		smtpReceipts:    make(map[string]ports.SMTPSettingsReceipt),
		smtpTestIntents: make(map[string]ports.SMTPTestIntent)}
}

func (value *testStores) Effects() ports.EffectStore            { return value }
func (value *testStores) Messages() ports.MessageStore          { return value }
func (value *testStores) Preferences() ports.PreferenceStore    { return value }
func (value *testStores) Templates() ports.TemplateStore        { return value }
func (value *testStores) Providers() ports.ProviderConfigStore  { return value }
func (value *testStores) SMTPSettings() ports.SMTPSettingsStore { return value }
func (value *testStores) Deliveries() ports.DeliveryStore       { return value }
func (value *testStores) Realtime() ports.RealtimeStore         { return value }
func (value *testStores) Evidence() ports.EvidenceStore         { return value }

func (*testStores) LockEffect(context.Context, string, [32]byte, domain.ID) error { return nil }

func (value *testStores) FindEffect(_ context.Context, producer string, hash [32]byte,
	recipient domain.ID,
) (domain.EffectReceipt, bool, error) {
	receipt, found := value.effects[testKey(producer, hash, recipient)]
	return receipt, found, nil
}

func (value *testStores) InsertEffect(_ context.Context, receipt domain.EffectReceipt) error {
	value.effects[testKey(receipt.Producer, receipt.EffectKeyHash, receipt.RecipientID)] = receipt
	return nil
}

func (value *testStores) NextInboxRevision(context.Context, domain.ID) (int64, error) {
	value.revision++
	return value.revision, nil
}

func (value *testStores) RecordMessage(_ context.Context, message domain.Message) error {
	value.messages[message.ID] = message
	return nil
}

func (value *testStores) GetMessage(_ context.Context, id, recipient domain.ID) (domain.Message, error) {
	message, found := value.messages[id]
	if !found || message.RecipientID != recipient {
		return domain.Message{}, ErrNotFound
	}
	return message, nil
}

func (value *testStores) ListMessages(_ context.Context, recipient domain.ID,
	filter ports.MessageFilter,
) (ports.MessagePage, error) {
	page := ports.MessagePage{}
	for _, message := range value.messages {
		if message.RecipientID != recipient || message.DismissedAt != nil ||
			filter.UnreadOnly && message.ReadAt != nil {
			continue
		}
		page.Items = append(page.Items, message)
		if message.ReadAt == nil {
			page.Unread++
		}
	}
	return page, nil
}

func (value *testStores) PersistMessageRevision(_ context.Context, message domain.Message, _ int64) error {
	value.messages[message.ID] = message
	return nil
}

func (value *testStores) CompleteInboxRead(_ context.Context, recipient domain.ID,
	boundary int64, _ domain.ID,
) (int64, error) {
	var count int64
	for id, message := range value.messages {
		if message.RecipientID != recipient || message.ReadAt != nil ||
			message.InboxRevision > boundary {
			continue
		}
		updated, _ := message.MarkRead(message.Revision, time.Now())
		value.messages[id] = updated
		count++
	}
	return count, nil
}

func (value *testStores) CurrentInboxRevision(context.Context, domain.ID) (int64, error) {
	return value.revision, nil
}

func testPreferenceKey(recipient domain.ID, category domain.Category, channel domain.Channel) string {
	return recipient.String() + ":" + string(category) + ":" + string(channel)
}

func (value *testStores) LoadPreference(_ context.Context, recipient domain.ID,
	category domain.Category, channel domain.Channel,
) (domain.Preference, bool, error) {
	preference, found := value.preferences[testPreferenceKey(recipient, category, channel)]
	return preference, found, nil
}

func (*testStores) ReadPreferences(context.Context, domain.ID) ([]domain.Preference, error) {
	return nil, nil
}

func (value *testStores) SetPreference(_ context.Context, preference domain.Preference, _ int64) error {
	value.preferences[testPreferenceKey(preference.RecipientID, preference.Category, preference.Channel)] = preference
	return nil
}

func (value *testStores) LoadPublishedTemplate(context.Context, domain.ID,
	[]string,
) (domain.PublishedTemplate, error) {
	return value.template, nil
}

func (value *testStores) AddTemplate(_ context.Context, draft ports.TemplateDraft) error {
	value.createdTemplate = draft
	return nil
}
func (value *testStores) StoreTemplateVersion(_ context.Context, version ports.TemplateVersion) error {
	value.createdTemplateVersion = version
	return nil
}
func (*testStores) PublishTemplate(context.Context, domain.ID, domain.ID, int64) error { return nil }
