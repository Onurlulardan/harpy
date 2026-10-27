package smtp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"
	"time"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	gomail "github.com/wneessen/go-mail"
)

type LifecycleSender struct {
	settings LifecycleSettings
	sender   mailSender
	probe    func(context.Context) error
	now      func() time.Time
}

func OpenLifecycleSender(settings LifecycleSettings) (*LifecycleSender, error) {
	common := Settings{host: settings.host, port: settings.port,
		username: settings.username, password: settings.password,
		timeout: settings.timeout}
	client, err := gomail.NewClient(settings.host, common.clientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("build IAM lifecycle SMTP client: %w", ErrContract)
	}
	probe := func(ctx context.Context) error {
		connection, dialErr := client.DialToSMTPClientWithContext(ctx)
		if dialErr != nil {
			return dialErr
		}
		return client.CloseWithSMTPClient(connection)
	}
	return &LifecycleSender{settings: settings, sender: client,
		probe: probe, now: time.Now}, nil
}

func (sender *LifecycleSender) Probe(ctx context.Context) error {
	if sender == nil || sender.probe == nil || ctx == nil || ctx.Err() != nil {
		return ErrContract
	}
	if err := sender.probe(ctx); err != nil {
		return fmt.Errorf("IAM lifecycle SMTP dependency preflight: %w", err)
	}
	return nil
}

func (sender *LifecycleSender) StableMessageID(effectKey string) (string, error) {
	if sender == nil || len(effectKey) < 1 || len(effectKey) > 512 {
		return "", ErrContract
	}
	separator := strings.LastIndexByte(sender.settings.from, '@')
	if separator < 1 || separator == len(sender.settings.from)-1 {
		return "", ErrContract
	}
	digest := sha256.Sum256([]byte("iam.lifecycle.smtp-message.v1\x00" + effectKey))
	return "lifecycle." + hex.EncodeToString(digest[:]) +
		sender.settings.from[separator:], nil
}

func (sender *LifecycleSender) Send(ctx context.Context,
	purpose lifecycledeliveryv1.Purpose, value Message) Result {
	if sender == nil || sender.sender == nil || ctx == nil || ctx.Err() != nil {
		return Result{Kind: ResultUnknown, Code: "SMTP_OUTCOME_UNKNOWN"}
	}
	message, err := sender.buildMessage(purpose, value)
	resultClient := &Client{now: sender.now}
	if err != nil {
		return resultClient.result(ResultPermanent, "SMTP_MESSAGE_INVALID",
			value.MessageID, 0, "")
	}
	defer message.UnsetAllParts()
	bounded, cancel := context.WithTimeout(ctx, sender.settings.timeout)
	defer cancel()
	err = sender.sender.DialAndSendWithContext(bounded, message)
	if message.IsDelivered() {
		return resultClient.result(ResultAccepted, "SMTP_DATA_ACCEPTED",
			value.MessageID, 250, message.ServerResponse())
	}
	kind, code, status := classifySendError(err)
	return resultClient.result(kind, code, value.MessageID, status, "")
}

func (sender *LifecycleSender) buildMessage(purpose lifecycledeliveryv1.Purpose,
	value Message) (*gomail.Msg, error) {
	template, exists := sender.settings.templates[purpose]
	parsed, err := mail.ParseAddress(value.Recipient)
	if !exists || err != nil || parsed.Name != "" || parsed.Address != value.Recipient ||
		!messageIDPattern.MatchString(value.MessageID) ||
		!validTenantSlug(value.TenantSlug) ||
		!credentialPattern.Match(value.Token) || value.Date.IsZero() {
		return nil, ErrContract
	}
	message := gomail.NewMsg()
	if err := message.FromFormat(sender.settings.fromName, sender.settings.from); err != nil {
		return nil, ErrContract
	}
	if err := message.To(value.Recipient); err != nil {
		return nil, ErrContract
	}
	message.Subject(template.subject)
	message.SetMessageIDWithValue(value.MessageID)
	message.SetDateWithValue(value.Date.UTC())
	message.SetBulk()
	message.SetBodyWriter(gomail.TypeTextPlain,
		template.text.writerRoute(sender.settings.baseURL, template.route,
			value.Token, value.TenantSlug))
	message.AddAlternativeWriter(gomail.TypeTextHTML,
		template.html.writerRoute(sender.settings.baseURL, template.route,
			value.Token, value.TenantSlug))
	return message, nil
}
