package iam

import (
	"context"
	"fmt"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func openLifecycleSMTP(ctx context.Context,
	settings config.IAMInvitationSMTP) (*smtp.LifecycleSender, error) {
	host, port, username, password, timeout, err := settings.Connection()
	if err != nil {
		return nil, fmt.Errorf("read IAM lifecycle SMTP connection: %w", err)
	}
	from, fromName, baseURL, _, err := settings.Message()
	if err != nil {
		return nil, fmt.Errorf("read IAM lifecycle SMTP message config: %w", err)
	}
	mailSettings, err := smtp.NewLifecycleSettings(host, port, username,
		password.Value(), from, fromName, baseURL, timeout, lifecycleTemplates())
	if err != nil {
		return nil, fmt.Errorf("build IAM lifecycle SMTP settings: %w", err)
	}
	sender, err := smtp.OpenLifecycleSender(mailSettings)
	if err != nil {
		return nil, err
	}
	if err := sender.Probe(ctx); err != nil {
		return nil, err
	}
	return sender, nil
}

func lifecycleTemplates() []smtp.LifecycleTemplateInput {
	placeholder := "{{LIFECYCLE_URL}}"
	return []smtp.LifecycleTemplateInput{
		{Purpose: lifecycledeliveryv1.PurposeEmailVerification,
			Route: "/verify-email/", Subject: "Verify your email address",
			Text: "Complete email verification: {{LIFECYCLE_URL}}",
			HTML: smtp.GraphiteActionHTML("Verify your email",
				"Confirm this address to continue using Harpy.",
				"Verify email", placeholder)},
		{Purpose: lifecycledeliveryv1.PurposePasswordReset,
			Route: "/reset-password/", Subject: "Reset your password",
			Text: "Complete your password reset: {{LIFECYCLE_URL}}",
			HTML: smtp.GraphiteActionHTML("Reset your password",
				"Choose a new password to continue.",
				"Open reset page", placeholder)},
		{Purpose: lifecycledeliveryv1.PurposePortalSetup,
			Route: "/portal/setup/", Subject: "Complete your portal setup",
			Text: "Complete your portal setup: {{LIFECYCLE_URL}}",
			HTML: smtp.GraphiteActionHTML("Complete portal setup",
				"Finish setup to access your Harpy portal.",
				"Set up portal access", placeholder)},
		{Purpose: lifecycledeliveryv1.PurposeInitialAdminSetup,
			Route: "/initial-admin/setup/", Subject: "Set up your administrator account",
			Text: "Set up your administrator account: {{LIFECYCLE_URL}}",
			HTML: smtp.GraphiteActionHTML("Set up your administrator account",
				"Create your password to finish installation.",
				"Finish account setup", placeholder)},
	}
}
