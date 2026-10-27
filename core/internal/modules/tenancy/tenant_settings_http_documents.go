package tenancy

import (
	"encoding/json"
	"fmt"
	"strconv"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

type tenantSettingsSecret struct{ value []byte }

func (secret *tenantSettingsSecret) UnmarshalJSON(document []byte) error {
	if secret == nil {
		return fmt.Errorf("SMTP secret target is nil")
	}
	var value string
	if err := json.Unmarshal(document, &value); err != nil || len(value) > 4096 {
		return fmt.Errorf("SMTP secret is invalid")
	}
	clear(secret.value)
	secret.value = append(secret.value[:0], value...)
	return nil
}

type tenantSMTPDocument struct {
	Host     *string               `json:"smtp_host"`
	Port     *string               `json:"smtp_port"`
	Username *string               `json:"smtp_username"`
	Password *tenantSettingsSecret `json:"smtp_password"`
	From     *string               `json:"email_from"`
}

type tenantSettingsUpdateDocument struct {
	Name *string             `json:"name"`
	SMTP *tenantSMTPDocument `json:"smtp"`
}

type tenantSettingsTestEmailDocument struct {
	To string `json:"to"`
}

func (document *tenantSMTPDocument) mutation() (*settings.SMTPMutation, bool) {
	if document == nil {
		return nil, true
	}
	if document.Host == nil || document.Port == nil || document.Username == nil ||
		document.Password == nil || document.From == nil {
		return nil, false
	}
	port, err := strconv.Atoi(*document.Port)
	if err != nil || port < 1 || strconv.Itoa(port) != *document.Port {
		return nil, false
	}
	return &settings.SMTPMutation{Host: *document.Host, Port: port,
		Username: *document.Username, Password: append([]byte(nil), document.Password.value...),
		From: *document.From}, true
}

func (document *tenantSMTPDocument) clear() {
	if document != nil && document.Password != nil {
		clear(document.Password.value)
	}
}
