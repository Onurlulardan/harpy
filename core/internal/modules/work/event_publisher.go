package work

import (
	"context"
	"strings"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type eventPublisher struct{ catalogs platformasync.WriteCatalogs }

func newEventPublisher(catalogs platformasync.WriteCatalogs) (*eventPublisher, error) {
	_, producers, err := catalogs.Snapshot()
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	for _, eventType := range workMutationTypes() {
		policy, resolveErr := producers.Resolve(eventType, 1)
		if resolveErr != nil || policy.OwnerDomain != "work" ||
			policy.PayloadSchemaSHA256 != asyncmanifestv1.WorkMutationSchemaSHA256 {
			return nil, ports.ErrUnavailable
		}
	}
	return &eventPublisher{catalogs: catalogs}, nil
}

func (publisher *eventPublisher) Record(ctx context.Context, local ports.Scope,
	intent ports.EventIntent,
) error {
	scope, ok := platformScope(local)
	if publisher == nil || !ok || !eventIntentValid(intent) || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireOwner(scope, "work") != nil || scope.Mode() != platformtx.Command {
		return ports.ErrInvalid
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "work")
	if err != nil {
		return ports.ErrUnavailable
	}
	writer, err := asyncpostgres.Bind(scope, authority, publisher.catalogs)
	if err != nil {
		return ports.ErrUnavailable
	}
	envelope, err := publisher.envelope(authority, local, intent)
	if err != nil {
		return ports.ErrInvalid
	}
	_, err = writer.EnqueueEvent(ctx, envelope)
	return err
}

func eventIntentValid(intent ports.EventIntent) bool {
	return intent.EventID.Valid() && intent.AggregateID.Valid() && intent.Type != "" &&
		strings.HasPrefix(intent.Type, "work.") && intent.AggregateKind != "" &&
		intent.AggregateVersion > 0 && intent.EventSequence > 0 &&
		!intent.OccurredAt.IsZero() && intent.IdempotencyKey != ""
}

var _ ports.EventPublisher = (*eventPublisher)(nil)
