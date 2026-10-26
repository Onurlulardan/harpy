package iam

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) relationProjections(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	resource publicresource.Resource,
	items []json.RawMessage,
) ([]publicresource.RelationProjection, error) {
	references := make([][]identityRelationReference, len(items))
	targets := make(map[publicresource.Resource]map[uuid.UUID]struct{})
	total := 0
	for index, item := range items {
		var err error
		references[index], err = identityRelationReferences(resource, item)
		if err != nil {
			return nil, err
		}
		for _, reference := range references[index] {
			total += len(reference.ids)
			if total > maximumIdentityRelationReferences {
				return nil, publicresource.ErrUnavailable
			}
			set := targets[reference.definition.target]
			if set == nil {
				set = make(map[uuid.UUID]struct{})
				targets[reference.definition.target] = set
			}
			for _, id := range reference.ids {
				set[id] = struct{}{}
			}
		}
	}
	resolved := make(map[publicresource.Resource]map[uuid.UUID]publicresource.RelationSummary)
	visible := make(map[publicresource.Resource]bool)
	for target, set := range targets {
		ids := identitySortedRelationIDs(set)
		summaries, allowed, err := port.resolveRelationSummaries(
			ctx, scope, evidence, target, ids,
		)
		if err != nil {
			return nil, err
		}
		resolved[target], visible[target] = summaries, allowed
	}
	result := make([]publicresource.RelationProjection, len(items))
	for index := range result {
		values, err := identityRelationValues(
			references[index], resolved, visible,
		)
		if err != nil {
			return nil, err
		}
		result[index], err = publicresource.NewRelationProjection(resource, values)
		if err != nil {
			return nil, publicresource.ErrUnavailable
		}
	}
	return result, nil
}

func identitySortedRelationIDs(
	set map[uuid.UUID]struct{},
) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].String() < result[right].String()
	})
	return result
}

func identityRelationValues(
	references []identityRelationReference,
	resolved map[publicresource.Resource]map[uuid.UUID]publicresource.RelationSummary,
	visible map[publicresource.Resource]bool,
) ([]publicresource.RelationValue, error) {
	values := make([]publicresource.RelationValue, 0, len(references))
	for _, reference := range references {
		if !visible[reference.definition.target] {
			continue
		}
		summaries := make([]publicresource.RelationSummary, len(reference.ids))
		for index, id := range reference.ids {
			summary, found := resolved[reference.definition.target][id]
			if !found {
				return nil, publicresource.ErrUnavailable
			}
			summaries[index] = summary
		}
		var (
			value publicresource.RelationValue
			err   error
		)
		if reference.definition.multiple {
			value, err = publicresource.NewToManyRelation(
				reference.definition.field, summaries,
			)
		} else {
			var summary *publicresource.RelationSummary
			if len(summaries) == 1 {
				summary = &summaries[0]
			}
			value, err = publicresource.NewToOneRelation(
				reference.definition.field, summary,
			)
		}
		if err != nil {
			return nil, publicresource.ErrUnavailable
		}
		values = append(values, value)
	}
	return values, nil
}

func relationReadForbidden(err error) bool {
	return errors.Is(err, publicresource.ErrForbidden)
}
