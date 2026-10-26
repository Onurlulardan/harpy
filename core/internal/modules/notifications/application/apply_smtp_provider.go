package application

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *SMTPSettingsApplication) applySMTPProvider(ctx context.Context,
	stores ports.Stores, request api.SMTPSettingsReplaceRequest,
) (ports.ProviderConfig, error) {
	metadata := ports.ProviderMetadata{SMTPHost: request.Host,
		SMTPUsername: request.Username, SMTPPort: request.Port,
		SMTPSecurity: smtpSecurity(request.Port), FromAddress: request.From}
	current, err := stores.Providers().SelectProvider(ctx,
		domain.ChannelEmail, domain.ID{})
	if errors.Is(err, ports.ErrProviderNotFound) {
		return application.createSMTPProvider(ctx, stores, metadata, request.Password)
	}
	if err != nil || current.Kind != ports.ProviderSMTP || !current.Enabled ||
		current.Revision < 1 {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	if len(request.Password) == 0 && current.Metadata.SMTPUsername == "" {
		return ports.ProviderConfig{}, ErrInvalid
	}
	versionID, err := application.base.nextID(application.base.clock.Now().UTC())
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	expected := current.Revision
	current.VersionID, current.Metadata = versionID, metadata
	current.MetadataDigest = providerMetadataDigest(current.Kind, metadata)
	current.Revision++
	if len(request.Password) == 0 {
		if err = stores.Providers().UpdateProviderConfig(ctx, current, expected); err != nil {
			return ports.ProviderConfig{}, ErrConflict
		}
		return current, nil
	}
	secretID, err := application.base.nextID(application.base.clock.Now().UTC())
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	sealed, err := application.sealer.Seal(ctx, secretID,
		request.Password, application.kekRevision)
	if err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	current.SecretID = secretID
	if err = stores.Providers().RotateProviderSecret(ctx, current, sealed, expected); err != nil {
		return ports.ProviderConfig{}, ErrConflict
	}
	return current, nil
}

func (application *SMTPSettingsApplication) createSMTPProvider(ctx context.Context,
	stores ports.Stores, metadata ports.ProviderMetadata, password []byte,
) (ports.ProviderConfig, error) {
	if len(password) == 0 {
		return ports.ProviderConfig{}, ErrInvalid
	}
	now := application.base.clock.Now().UTC()
	configID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	versionID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	secretID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	sealed, err := application.sealer.Seal(ctx, secretID, password,
		application.kekRevision)
	if err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	config := ports.ProviderConfig{ID: configID, VersionID: versionID,
		SecretID: secretID, Kind: ports.ProviderSMTP, Name: "Tenant SMTP",
		Enabled: true, Revision: 1, Metadata: metadata,
		MetadataDigest: providerMetadataDigest(ports.ProviderSMTP, metadata)}
	if err = stores.Providers().StoreProviderSecret(ctx, sealed); err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	if err = stores.Providers().CreateProviderConfig(ctx, config); err != nil {
		return ports.ProviderConfig{}, ErrConflict
	}
	return config, nil
}

func smtpSecurity(port int) string {
	if port == 465 {
		return "implicit_tls"
	}
	return "starttls"
}
