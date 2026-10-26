package iam

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func profileBackfillPublicationInput(
	runtime *ProfileBackfillPublicationRuntime,
	scope platformtx.TenantScope,
	source profileapp.Schema,
	target profileapp.Schema,
	evidence profileapp.MutationEvidence,
) bool {
	return runtime != nil && runtime.reader != nil && runtime.plans != nil &&
		runtime.clock != nil && evidence.Validate() == nil && scope.ActorID() == evidence.ActorID &&
		source.ID() != uuid.Nil && target.ID() != uuid.Nil && source.ID() != target.ID() &&
		source.Resource() == target.Resource() && source.State() == profileapp.SchemaActive &&
		target.State() == profileapp.SchemaActive && target.Version() == source.Version()+1
}

func profileBackfillEvidence(
	evidence profileapp.MutationEvidence,
) profilebackfillapp.Evidence {
	return profilebackfillapp.Evidence{
		OperationID: evidence.OperationID, RequestSHA256: evidence.RequestSHA256,
		ActorID: evidence.ActorID, ManagePermissionID: evidence.ManagePermissionID,
		ExpectedTokenVersion:         evidence.ExpectedTokenVersion,
		ExpectedAuthorizationVersion: evidence.ExpectedAuthorizationVersion,
	}
}

func (runtime *ProfileBackfillPublicationRuntime) buildPlan(
	ctx context.Context,
	scope platformtx.TenantScope,
	source profileapp.Schema,
	target profileapp.Schema,
	evidence profileapp.MutationEvidence,
	migrations []ProfileMigrationMappingInput,
	first []profilebackfillapp.SourceResource,
) (profilebackfillapp.Plan, error) {
	mappings, err := profileBackfillMappings(source, target, migrations)
	if err != nil {
		wipeProfileSources(first)
		return profilebackfillapp.Plan{}, err
	}
	sharedMappings, err := profileBackfillMigrationMappings(mappings)
	if err != nil {
		wipeProfileSources(first)
		return profilebackfillapp.Plan{}, err
	}
	schemaReader, err := bindProfileHistoricalSchemaReader(scope)
	if err != nil {
		wipeProfileSources(first)
		return profilebackfillapp.Plan{}, err
	}
	planID := uuid.NewSHA1(evidence.OperationID, []byte("iam.profile.backfill.plan.v1"))
	runAt := runtime.clock().UTC()
	if runAt.IsZero() {
		wipeProfileSources(first)
		return profilebackfillapp.Plan{}, profilebackfillapp.ErrUnavailable
	}
	pages := make([]profilebackfillapp.Page, 0, 1)
	rows, after := first, uuid.Nil
	var sourceCount int64
	for pageNumber := 1; len(rows) > 0; pageNumber++ {
		if pageNumber > 65536 {
			wipeProfileSources(rows)
			return profilebackfillapp.Plan{}, profilebackfillapp.ErrInvalid
		}
		schemas, schemaErr := profileBackfillActualSchemas(
			ctx, schemaReader, source, rows,
		)
		if schemaErr != nil {
			wipeProfileSources(rows)
			return profilebackfillapp.Plan{}, schemaErr
		}
		page, pageErr := profileBackfillPage(
			planID, pageNumber, runAt, after, source, target,
			sharedMappings, schemas, rows,
		)
		wipeProfileSources(rows)
		if pageErr != nil {
			return profilebackfillapp.Plan{}, pageErr
		}
		pages = append(pages, page)
		sourceCount += int64(page.ExpectedCount)
		after = page.ThroughResourceID
		if page.ExpectedCount < profilebackfillapp.MaximumPageSize {
			break
		}
		rows, err = runtime.readPage(ctx, scope, source, evidence, after)
		if err != nil {
			return profilebackfillapp.Plan{}, err
		}
	}
	definition := target.DefinitionSHA256()
	return profilebackfillapp.Plan{
		ID: planID, SourceSchemaID: source.ID(),
		Target: profilebackfillapp.TargetSchema{
			ID: target.ID(), ResourceKind: string(target.Resource()), Version: target.Version(),
			DefinitionJSON: target.DefinitionJSON(), DefinitionSHA256: definition,
			Fields: profileBackfillTargetFields(target),
		},
		ConverterVersion: profilebackfillapp.ConverterTypedMap,
		SourceCount:      sourceCount, PageSize: profilebackfillapp.MaximumPageSize,
		InitialRollingSHA256: sha256.Sum256(nil), Mappings: mappings, Pages: pages,
	}, nil
}

func profileBackfillMappings(
	source profileapp.Schema,
	target profileapp.Schema,
	input []ProfileMigrationMappingInput,
) ([]profilebackfillapp.Mapping, error) {
	declared, err := profileDeclaredMappings(source, target, input)
	if err != nil {
		return nil, err
	}
	declaredTargets := make(map[uuid.UUID]bool, len(declared))
	for _, mapping := range declared {
		declaredTargets[mapping.TargetFieldID] = true
	}
	mappings := make([]profilebackfillapp.Mapping, 0, len(target.Fields())+len(declared))
	for _, targetField := range target.Fields() {
		if targetField.State() != profileapp.FieldActive || declaredTargets[targetField.ID()] {
			continue
		}
		sourceField, found := source.Field(targetField.ID())
		if !found {
			continue
		}
		if sourceField.SemanticSHA256() != targetField.SemanticSHA256() {
			return nil, ErrProfileBackfillUnmappable
		}
		mappings = append(mappings, profilebackfillapp.Mapping{
			SourceFieldID: sourceField.ID(), TargetFieldID: targetField.ID(),
			ConverterKind: profilebackfillapp.MappingCopy,
		})
	}
	mappings = append(mappings, declared...)
	if len(mappings) == 0 {
		return nil, ErrProfileBackfillUnmappable
	}
	sort.Slice(mappings, func(left, right int) bool {
		if mappings[left].TargetFieldID != mappings[right].TargetFieldID {
			return mappings[left].TargetFieldID.String() < mappings[right].TargetFieldID.String()
		}
		return mappings[left].SourceFieldID.String() < mappings[right].SourceFieldID.String()
	})
	sources, targets := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, mapping := range mappings {
		if sources[mapping.SourceFieldID] || targets[mapping.TargetFieldID] {
			return nil, ErrProfileBackfillUnmappable
		}
		sources[mapping.SourceFieldID], targets[mapping.TargetFieldID] = true, true
	}
	return mappings, nil
}

func profileBackfillTargetFields(
	target profileapp.Schema,
) []profilebackfillapp.TargetField {
	fields := make([]profilebackfillapp.TargetField, 0, len(target.Fields()))
	for _, field := range target.Fields() {
		fields = append(fields, profilebackfillapp.TargetField{
			ID: field.ID(), Required: field.Required(),
			Active:         field.State() == profileapp.FieldActive,
			SemanticSHA256: field.SemanticSHA256(),
		})
	}
	return fields
}

func profileBackfillJobID(planID uuid.UUID, pageNumber int) uuid.UUID {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", planID, pageNumber)))
	var result uuid.UUID
	copy(result[:], digest[:len(result)])
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}

func profileBackfillRunAt(base time.Time, pageNumber int) time.Time {
	return base.Add(time.Duration(pageNumber-1) * time.Nanosecond).UTC()
}
