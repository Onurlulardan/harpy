package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func openInvitationSMTP(ctx context.Context,
	settings config.IAMInvitationSMTP) (*smtp.Client, error) {
	host, port, username, password, timeout, err := settings.Connection()
	if err != nil {
		return nil, fmt.Errorf("read IAM invitation SMTP connection: %w", err)
	}
	from, fromName, baseURL, template, err := settings.Message()
	if err != nil {
		return nil, fmt.Errorf("read IAM invitation SMTP message config: %w", err)
	}
	subject, textBody, htmlBody, err := template.Values()
	if err != nil {
		return nil, fmt.Errorf("read IAM invitation template: %w", err)
	}
	mailSettings, err := smtp.NewSettings(host, port, username, password.Value(),
		from, fromName, baseURL, subject, textBody, htmlBody, timeout)
	if err != nil {
		return nil, fmt.Errorf("build IAM invitation SMTP settings: %w", err)
	}
	client, err := smtp.Open(mailSettings)
	if err != nil {
		return nil, err
	}
	if err := client.Probe(ctx); err != nil {
		return nil, err
	}
	return client, nil
}
