package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type outboxCatalogArrays struct {
	types         []string
	versions      []int32
	domains       []string
	schemaSHA256s []string
}

func (repository *workerRepository) AssertOutboxCatalog(
	ctx context.Context,
	catalog eventsv1.ProducerCatalog,
) error {
	if ctx == nil {
		return ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return ErrWorkerStoreUnavailable
	}
	policies, err := validatedOutboxPolicies(catalog)
	if err != nil {
		return err
	}
	rows, err := repository.database.Query(
		ctx, assertLiveOutboxCatalogSQL, policies.types, policies.versions,
		policies.domains, policies.schemaSHA256s,
	)
	if err != nil {
		return fmt.Errorf("inspect live outbox catalog: %w", err)
	}
	defer rows.Close()
	var compatible bool
	if !rows.Next() {
		return fmt.Errorf(
			"inspect live outbox catalog: %w",
			errors.Join(rows.Err(), ErrOutboxCatalogIncompatible),
		)
	}
	if err := rows.Scan(&compatible); err != nil {
		return fmt.Errorf("scan live outbox catalog: %w", err)
	}
	if rows.Next() || rows.Err() != nil {
		return fmt.Errorf("inspect live outbox catalog cardinality: %w", ErrOutboxCatalogIncompatible)
	}
	if !compatible {
		detail := incompatibleOutboxDetail(ctx, repository, policies)
		if detail == "" {
			return ErrOutboxCatalogIncompatible
		}
		return fmt.Errorf("%w: %s", ErrOutboxCatalogIncompatible, detail)
	}
	return nil
}

func incompatibleOutboxDetail(
	ctx context.Context,
	repository *workerRepository,
	policies outboxCatalogArrays,
) string {
	rows, err := repository.database.Query(
		ctx, listIncompatibleOutboxCatalogSQL, policies.types, policies.versions,
		policies.domains, policies.schemaSHA256s,
	)
	if err != nil {
		return ""
	}
	defer rows.Close()
	var mismatches []string
	for rows.Next() {
		var eventType, domain, schema string
		var version int64
		if rows.Scan(&eventType, &version, &domain, &schema) != nil {
			return ""
		}
		mismatches = append(mismatches, fmt.Sprintf(
			"%s version %d domain %s schema %s", eventType, version, domain, schema,
		))
	}
	if rows.Err() != nil {
		return ""
	}
	return strings.Join(mismatches, "; ")
}

func validatedOutboxPolicies(catalog eventsv1.ProducerCatalog) (outboxCatalogArrays, error) {
	policies := catalog.Policies()
	validated, err := eventsv1.NewProducerCatalog(policies)
	if err != nil {
		return outboxCatalogArrays{}, fmt.Errorf("validate outbox producer catalog: %w", err)
	}
	result := outboxCatalogArrays{
		types: make([]string, 0, len(policies)), versions: make([]int32, 0, len(policies)),
		domains: make([]string, 0, len(policies)), schemaSHA256s: make([]string, 0, len(policies)),
	}
	for _, policy := range validated.Policies() {
		result.types = append(result.types, policy.EventType)
		result.versions = append(result.versions, int32(policy.EventVersion))
		result.domains = append(result.domains, policy.OwnerDomain)
		result.schemaSHA256s = append(result.schemaSHA256s, policy.PayloadSchemaSHA256)
	}
	return result, nil
}
