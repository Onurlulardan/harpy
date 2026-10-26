package notifications

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/providerregistry"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/secretcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/smtp"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/telegram"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type productionClock struct{}

func (productionClock) Now() time.Time { return time.Now() }

type productionIDs struct{}

func (productionIDs) New(at time.Time) (ports.ID, error) {
	return application.NewID(at, rand.Reader)
}

func NewProductionDeliveryJobProcessors(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore, coordinator *platformtx.Coordinator,
	keys map[int64][32]byte,
) ([]jobexecution.Processor, error) {
	runtime, err := postgres.NewDeliveryRuntime(coordinator)
	if err != nil {
		return nil, fmt.Errorf("build notification delivery store: %w", err)
	}
	wrapper, err := secretcrypto.NewAESKeyWrapper(keys, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("build notification key wrapper: %w", err)
	}
	sealer, err := secretcrypto.New(rand.Reader, wrapper)
	if err != nil {
		return nil, fmt.Errorf("build notification secret opener: %w", err)
	}
	smtpClient, err := smtp.NewProductionClient()
	if err != nil {
		return nil, fmt.Errorf("build notification SMTP client: %w", err)
	}
	telegramClient, err := telegram.NewProductionClient()
	if err != nil {
		return nil, fmt.Errorf("build notification Telegram client: %w", err)
	}
	registry, err := providerregistry.New(smtpClient, telegramClient)
	if err != nil {
		return nil, fmt.Errorf("build notification provider registry: %w", err)
	}
	worker, err := application.NewDeliveryWorker(runtime,
		deliveryTargetResolver{runtime: runtime}, sealer, registry,
		productionClock{}, productionIDs{})
	if err != nil {
		return nil, fmt.Errorf("build notification delivery worker: %w", err)
	}
	return NewDeliveryJobProcessors(release, store, worker)
}
