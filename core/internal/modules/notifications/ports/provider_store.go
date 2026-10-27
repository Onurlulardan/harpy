package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type ProviderKind string

const (
	ProviderSMTP     ProviderKind = "smtp"
	ProviderTelegram ProviderKind = "telegram"
)

type ProviderConfig struct {
	ID, VersionID, SecretID domain.ID
	Kind                    ProviderKind
	Name                    string
	Enabled                 bool
	Revision                int64
	Metadata                ProviderMetadata
	MetadataDigest          [32]byte
}

type ProviderMetadata struct {
	SMTPHost        string `json:"smtp_host,omitempty"`
	SMTPUsername    string `json:"smtp_username,omitempty"`
	SMTPSecurity    string `json:"smtp_security,omitempty"`
	FromAddress     string `json:"from_address,omitempty"`
	SMTPPort        int    `json:"smtp_port,omitempty"`
	TelegramAPIBase string `json:"telegram_api_base,omitempty"`
}

type ProviderConfigStore interface {
	SelectProvider(context.Context, domain.Channel, domain.ID) (ProviderConfig, error)
	LoadProviderConfig(context.Context, domain.ID) (ProviderConfig, error)
	CreateProviderConfig(context.Context, ProviderConfig) error
	UpdateProviderConfig(context.Context, ProviderConfig, int64) error
	LoadProviderSecret(context.Context, domain.ID) (EncryptedSecret, error)
	StoreProviderSecret(context.Context, EncryptedSecret) error
	RotateProviderSecret(context.Context, ProviderConfig, EncryptedSecret, int64) error
	RewrapProviderSecret(context.Context, EncryptedSecret, int64) error
}

type EncryptedSecret struct {
	ID                domain.ID
	Ciphertext        []byte
	Nonce, WrappedDEK []byte
	KEKRevision       int64
	Digest            [32]byte
}

type SecretSealer interface {
	Seal(context.Context, domain.ID, []byte, int64) (EncryptedSecret, error)
	Rewrap(context.Context, EncryptedSecret, int64) (EncryptedSecret, error)
}

type SecretOpener interface {
	Open(context.Context, EncryptedSecret) ([]byte, error)
}
