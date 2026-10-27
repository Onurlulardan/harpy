package v1bridge

import (
	"context"
	"sort"

	bridgeapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1bridge"
	"github.com/google/uuid"
)

type Store struct {
	database Database
	enqueuer Enqueuer
	tenantID uuid.UUID
}

func (store *Store) LockAuthority(ctx context.Context,
	producer string,
) (bridgeapp.Authority, error) {
	if !store.usable(ctx) || !validProducer(producer) {
		return bridgeapp.Authority{}, bridgeapp.ErrAuthority
	}
	domains := []string{producer, "notifications"}
	if producer == "notifications" {
		return bridgeapp.Authority{}, bridgeapp.ErrAuthority
	}
	sort.Strings(domains)
	values := make(map[string]domainAuthority, 2)
	for _, domain := range domains {
		var value domainAuthority
		err := store.database.QueryRow(ctx, lockDomainAuthoritySQL,
			store.tenantID, domain).Scan(&value.state, &value.writer, &value.generation)
		if err != nil {
			return bridgeapp.Authority{}, bridgeapp.ErrAuthority
		}
		values[domain] = value
	}
	var localWriter, lifecycle string
	var localGeneration int64
	err := store.database.QueryRow(ctx, lockWriterAuthoritySQL).
		Scan(&localWriter, &lifecycle, &localGeneration)
	producerValue, notification := values[producer], values["notifications"]
	if err != nil || localGeneration < 1 || lifecycle != "writable" ||
		localWriter != notification.writer {
		return bridgeapp.Authority{}, bridgeapp.ErrAuthority
	}
	return bridgeapp.Authority{ProducerWriter: bridgeapp.Writer(producerValue.writer),
		ProducerState: producerValue.state, ProducerGeneration: producerValue.generation,
		NotificationWriter:     bridgeapp.Writer(notification.writer),
		NotificationState:      notification.state,
		NotificationGeneration: notification.generation}, nil
}

type domainAuthority struct {
	state, writer string
	generation    int64
}

func validProducer(value string) bool {
	switch value {
	case "analytics", "billing", "catalog", "documents", "forms", "iam", "work", "workflow":
		return true
	default:
		return false
	}
}

func (store *Store) usable(ctx context.Context) bool {
	return store != nil && store.database != nil && store.enqueuer != nil &&
		store.tenantID != uuid.Nil && ctx != nil && ctx.Err() == nil
}

var _ bridgeapp.Store = (*Store)(nil)
