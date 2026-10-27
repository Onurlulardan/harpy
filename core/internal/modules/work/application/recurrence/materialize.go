package recurrence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) MaterializeRecurrence(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, ruleID domain.ID,
	from, to time.Time, tenantRemaining int,
) (ports.OperationReceipt, error) {
	if !ruleID.Valid() || meta.ExpectedVersion < 1 || !to.After(from) ||
		tenantRemaining < 1 || tenantRemaining > domain.MaxFuturePerTenant {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(ruleID.String(), from.UTC().Format(time.RFC3339Nano),
		to.UTC().Format(time.RFC3339Nano), ports.VersionString(int64(tenantRemaining)),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.recurrence.materialize", ruleID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			record, err := stores.Calendar().LoadRecurrenceRevision(ctx, ruleID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if record.Revision != meta.ExpectedVersion {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			location, err := service.tzdb.Load(ctx, record.TimeZone, record.TZDBVersion)
			if err != nil {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			definition, err := domain.ParseRecurrence(domain.RecurrenceInput{DTStart: record.DTSTART,
				TimeZone: record.TimeZone, TZDBVersion: record.TZDBVersion, RRULE: record.RRULE,
				EXDATE: append([]string{}, record.EXDATE...), Location: location})
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			limit := domain.MaxFuturePerRule
			if tenantRemaining < limit {
				limit = tenantRemaining
			}
			occurrences, err := domain.Materialize(definition, from.UTC(), to.UTC(), limit)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			generated := make([]ports.GeneratedOccurrence, 0, len(occurrences))
			for _, occurrence := range occurrences {
				id, idErr := service.ids.New(service.clock.Now().UTC())
				if idErr != nil {
					return ports.OperationReceipt{}, nil, ports.ErrUnavailable
				}
				generated = append(generated, ports.GeneratedOccurrence{ID: id, Occurrence: occurrence})
			}
			inserted, err := stores.Calendar().InsertGeneratedEvents(ctx, ruleID, generated)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			payload, _ := json.Marshal(struct {
				RuleID   string
				Revision int64
				Inserted int
			}{
				ruleID.String(), record.Revision, inserted})
			return ports.OperationReceipt{AggregateID: ruleID, RowVersion: record.Revision}, payload, nil
		})
}
