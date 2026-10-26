package tenantsettings

import (
	"net"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	uuidNil            = uuid.Nil
	idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)
	hostPattern        = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*$`)
)

func normalizeUpdate(command UpdateCommand) (UpdateCommand, bool) {
	if command.Name != nil {
		value := strings.TrimSpace(*command.Name)
		command.Name = &value
	}
	return command, command.ExpectedRevision > 0 &&
		validIdempotencyKey(command.IdempotencyKey) &&
		(command.Name != nil || command.SMTP != nil) &&
		(command.Name == nil || *command.Name != "" && len(*command.Name) <= 255) &&
		(command.SMTP == nil || validSMTP(*command.SMTP))
}

func validSMTP(value SMTPMutation) bool {
	return hostPattern.MatchString(value.Host) && net.ParseIP(value.Host) == nil &&
		len(value.Host) <= 253 && value.Port > 0 && value.Port <= 65535 &&
		len(value.Username) <= 320 && !strings.ContainsAny(value.Username, "\r\n\x00") &&
		len(value.Password) <= 4096 && validEmail(value.From)
}

func normalizeEmail(value string) string { return strings.TrimSpace(value) }

func validEmail(value string) bool {
	if value == "" || len(value) > 512 || strings.ContainsAny(value, "\r\n\x00") {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address != ""
}

func validIdempotencyKey(value string) bool { return idempotencyPattern.MatchString(value) }

func validSnapshot(value Snapshot, tenantID uuid.UUID) bool {
	if value.ID != tenantID || value.Name == "" || value.Slug == "" ||
		(value.Status != "ACTIVE" && value.Status != "TRIAL" && value.Status != "SUSPENDED") ||
		value.Revision < 1 || value.CreatedAt.IsZero() || value.UpdatedAt.IsZero() ||
		value.UserCount < 0 || value.FormCount < 0 || value.PendingInvitationCount < 0 {
		return false
	}
	if value.SMTP != nil && !validSMTP(SMTPMutation{
		Host: value.SMTP.Host, Port: value.SMTP.Port, Username: value.SMTP.Username,
		From: value.SMTP.From}) {
		return false
	}
	return value.Subscription == nil || value.Subscription.MaxUsers >= 0 &&
		value.Subscription.MaxForms >= 0 && value.Subscription.MaxStorageMB >= 0 &&
		value.Subscription.SubmissionsPerMonth >= 0
}
