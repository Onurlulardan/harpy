package notifications

import (
	"crypto/rand"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/secretcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type PlatformSMTPSettingsDependencies struct {
	Catalogs           platformasync.WriteCatalogs
	KEKKeys            map[int64][32]byte
	CurrentKEKRevision int64
}

func NewPlatformSMTPSettingsPort(
	dependencies PlatformSMTPSettingsDependencies,
) (api.SMTPSettingsPort, error) {
	if dependencies.CurrentKEKRevision < 1 {
		return nil, application.ErrConfiguration
	}
	if _, found := dependencies.KEKKeys[dependencies.CurrentKEKRevision]; !found {
		return nil, application.ErrConfiguration
	}
	binder, err := newPlatformNotificationBinder(dependencies.Catalogs)
	if err != nil {
		return nil, err
	}
	base, err := application.NewApplication(binder, recipientResolver{binder: binder},
		newScopeAuthorizer(), productionClock{}, productionIDs{})
	if err != nil {
		return nil, fmt.Errorf("build notification application: %w", err)
	}
	wrapper, err := secretcrypto.NewAESKeyWrapper(dependencies.KEKKeys, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("build notification key wrapper: %w", err)
	}
	sealer, err := secretcrypto.New(rand.Reader, wrapper)
	if err != nil {
		return nil, fmt.Errorf("build notification secret sealer: %w", err)
	}
	settings, err := application.NewSMTPSettingsApplication(base, sealer,
		dependencies.CurrentKEKRevision)
	if err != nil {
		return nil, fmt.Errorf("build notification SMTP settings application: %w", err)
	}
	return settings, nil
}
