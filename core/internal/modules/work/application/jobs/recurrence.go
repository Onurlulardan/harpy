package jobs

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runner *Runner) materialize(ctx context.Context, claim Claim) (Result, error) {
	revision, err := runner.store.LoadRecurrence(ctx, claim.TargetID)
	if err != nil || revision.Revision < 1 || revision.Contract != domain.RecurrenceContract {
		return Result{}, ports.ErrInvalid
	}
	location, err := runner.tzdb.Load(ctx, revision.TimeZone, revision.TZDBVersion)
	if err != nil {
		return Result{}, ports.ErrUnavailable
	}
	definition, err := domain.ParseRecurrence(domain.RecurrenceInput{
		DTStart: revision.DTSTART, TimeZone: revision.TimeZone,
		TZDBVersion: revision.TZDBVersion, RRULE: revision.RRULE,
		EXDATE: append([]string(nil), revision.EXDATE...), Location: location,
	})
	if err != nil {
		return Result{}, err
	}
	future, err := runner.store.FutureOccurrenceCount(ctx)
	if err != nil || future < 0 || future >= domain.MaxFuturePerTenant {
		return Result{}, domain.ErrRecurrenceLimit
	}
	limit := domain.MaxFuturePerTenant - future
	if limit > domain.MaxFuturePerRule {
		limit = domain.MaxFuturePerRule
	}
	occurrences, err := domain.Materialize(definition,
		claim.Now.Add(-domain.RecurrencePastWindow),
		claim.Now.Add(domain.RecurrenceFutureWindow), limit)
	if err != nil {
		return Result{}, err
	}
	generated := make([]ports.GeneratedOccurrence, len(occurrences))
	for index, occurrence := range occurrences {
		id, idErr := runner.ids.New(claim.Now)
		if idErr != nil || !id.UUIDv7() {
			return Result{}, ports.ErrUnavailable
		}
		generated[index] = ports.GeneratedOccurrence{ID: id, Occurrence: occurrence}
	}
	inserted, err := runner.store.InsertOccurrences(ctx, claim.TargetID, generated)
	if err != nil || inserted < 0 || inserted > len(generated) {
		return Result{}, ports.ErrUnavailable
	}
	return Result{Occurrences: inserted}, nil
}
