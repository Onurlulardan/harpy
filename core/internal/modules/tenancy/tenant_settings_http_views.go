package tenancy

import (
	"strconv"
	"time"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

type tenantSettingsSMTPView struct {
	Host               string `json:"smtp_host"`
	Port               string `json:"smtp_port"`
	Username           string `json:"smtp_username"`
	Password           string `json:"smtp_password"`
	From               string `json:"email_from"`
	PasswordConfigured bool   `json:"smtp_password_configured"`
}

type tenantSettingsSubscriptionView struct {
	PlanName            string `json:"plan_name"`
	PlanSlug            string `json:"plan_slug"`
	Status              string `json:"status"`
	MaxUsers            int    `json:"max_users"`
	MaxForms            int    `json:"max_forms"`
	MaxStorageMB        int64  `json:"max_storage_mb"`
	SubmissionsPerMonth int64  `json:"submissions_per_month"`
}

type tenantSettingsResponseView struct {
	ID                     string                          `json:"id"`
	Name                   string                          `json:"name"`
	Slug                   string                          `json:"slug"`
	Status                 string                          `json:"status"`
	Revision               int64                           `json:"revision"`
	CreatedAt              string                          `json:"created_at"`
	UpdatedAt              string                          `json:"updated_at"`
	UserCount              int64                           `json:"user_count"`
	FormCount              int64                           `json:"form_count"`
	PendingInvitationCount int64                           `json:"pending_invitation_count"`
	SMTP                   *tenantSettingsSMTPView         `json:"smtp,omitempty"`
	Subscription           *tenantSettingsSubscriptionView `json:"subscription,omitempty"`
}

func tenantSettingsView(value settings.Snapshot) tenantSettingsResponseView {
	result := tenantSettingsResponseView{ID: value.ID.String(), Name: value.Name,
		Slug: value.Slug, Status: value.Status, Revision: value.Revision,
		CreatedAt: value.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: value.UpdatedAt.UTC().Format(time.RFC3339Nano),
		UserCount: value.UserCount, FormCount: value.FormCount,
		PendingInvitationCount: value.PendingInvitationCount}
	if value.SMTP != nil {
		result.SMTP = &tenantSettingsSMTPView{Host: value.SMTP.Host,
			Port: strconv.Itoa(value.SMTP.Port), Username: value.SMTP.Username,
			From: value.SMTP.From, PasswordConfigured: value.SMTP.PasswordConfigured}
	}
	if value.Subscription != nil {
		result.Subscription = &tenantSettingsSubscriptionView{
			PlanName: value.Subscription.PlanName, PlanSlug: value.Subscription.PlanSlug,
			Status: value.Subscription.Status, MaxUsers: value.Subscription.MaxUsers,
			MaxForms: value.Subscription.MaxForms, MaxStorageMB: value.Subscription.MaxStorageMB,
			SubmissionsPerMonth: value.Subscription.SubmissionsPerMonth}
	}
	return result
}

func tenantSettingsMutationView(value settings.UpdateResult) any {
	return struct {
		Revision int64 `json:"revision"`
		Replay   bool  `json:"replay"`
	}{Revision: value.Revision, Replay: value.Replay}
}

func tenantSettingsTestEmailView(value settings.TestEmailResult) any {
	return struct {
		OperationID string `json:"operation_id"`
		Replay      bool   `json:"replay"`
	}{OperationID: value.OperationID.String(), Replay: value.Replay}
}
