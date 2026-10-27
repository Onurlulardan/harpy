package iam

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type ProfileRecoveryInventoryInput struct {
	Parent     profilebackfillapp.RecoveryParent
	Source     profileapp.Schema
	Target     profileapp.Schema
	Normalizer profileapp.TextNormalizer
	Clock      profilebackfillapp.RecoveryInventoryClock
}

type ProfileRecoveryInventoryClassifier struct {
	source      profileapp.Schema
	target      profileapp.Schema
	mappings    []profileapp.MigrationMapping
	normalizer  profileapp.TextNormalizer
	accumulator *profilebackfillapp.RecoveryInventoryAccumulator
	budget      *profilebackfillapp.RecoveryInventoryBudget
}

func NewProfileRecoveryInventoryClassifier(
	input ProfileRecoveryInventoryInput,
) (*ProfileRecoveryInventoryClassifier, error) {
	if input.Parent.Validate() != nil ||
		!profileRecoverySchemasMatch(input.Parent, input.Source, input.Target) ||
		nilProfileHTTPDependency(input.Normalizer) {
		return nil, profilebackfillapp.ErrInvalid
	}
	mappings, err := profileBackfillMigrationMappings(input.Parent.Mappings)
	if err != nil {
		return nil, err
	}
	accumulator, err := profilebackfillapp.NewRecoveryInventoryAccumulator(
		input.Parent.ID, input.Source.ID(), input.Target.ID(),
		profilebackfillapp.MaximumPageSize,
	)
	if err != nil {
		return nil, err
	}
	budget, err := profilebackfillapp.NewRecoveryInventoryBudget(input.Clock)
	if err != nil {
		return nil, err
	}
	return &ProfileRecoveryInventoryClassifier{
		source: input.Source, target: input.Target, mappings: mappings,
		normalizer: input.Normalizer, accumulator: accumulator, budget: budget,
	}, nil
}

func (classifier *ProfileRecoveryInventoryClassifier) Check(
	ctx context.Context,
) error {
	if classifier == nil || classifier.budget == nil {
		return profilebackfillapp.ErrInvalid
	}
	return classifier.budget.Check(ctx)
}

func (classifier *ProfileRecoveryInventoryClassifier) AppendBatch(
	ctx context.Context,
	actualSchemas map[uuid.UUID]profileapp.Schema,
	resources []profilebackfillapp.SourceResource,
) error {
	if classifier == nil || classifier.accumulator == nil || classifier.budget == nil ||
		len(resources) == 0 ||
		len(resources) > profilebackfillapp.RecoveryInventoryReadBatchSize {
		wipeProfileSources(resources)
		return profilebackfillapp.ErrInvalid
	}
	defer wipeProfileSources(resources)
	selected := make(map[uuid.UUID][]profileapp.MigrationMapping, len(actualSchemas))
	for index := range resources {
		resource := resources[index]
		err := classifier.append(ctx, resource, actualSchemas, selected)
		clear(resources[index].Document)
		resources[index].Document = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (classifier *ProfileRecoveryInventoryClassifier) Finish(
	ctx context.Context,
) (profilebackfillapp.RecoveryInventory, error) {
	if err := classifier.Check(ctx); err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	return classifier.accumulator.Seal()
}

func (classifier *ProfileRecoveryInventoryClassifier) append(
	ctx context.Context,
	resource profilebackfillapp.SourceResource,
	actualSchemas map[uuid.UUID]profileapp.Schema,
	selected map[uuid.UUID][]profileapp.MigrationMapping,
) error {
	if resource.Validate() != nil {
		return profilebackfillapp.ErrInvalid
	}
	if err := classifier.budget.Admit(ctx, len(resource.Document)); err != nil {
		return err
	}
	if resource.SchemaVersionID == classifier.target.ID() {
		return classifier.appendTarget(resource)
	}
	actual, found := actualSchemas[resource.SchemaVersionID]
	if !found || !validProfileRecoverySource(
		classifier.source, actual, resource.SchemaVersionID,
	) {
		return profilebackfillapp.ErrInvalid
	}
	mappings, found := selected[resource.SchemaVersionID]
	if !found {
		var err error
		mappings, err = profileapp.SelectMigrationMappingsForSource(
			actual, classifier.target, classifier.mappings,
		)
		if err != nil {
			return profilebackfillapp.ErrInvalid
		}
		selected[resource.SchemaVersionID] = mappings
	}
	converted, err := profileapp.MigrateDocument(
		actual, classifier.target, mappings, resource.Document,
		resource.SHA256[:], classifier.normalizer,
	)
	if err != nil {
		return profilebackfillapp.ErrInvalid
	}
	return classifier.accumulator.AppendRemaining(profilebackfillapp.RecoveryResource{
		ID: resource.ID, SchemaVersionID: resource.SchemaVersionID,
		Revision: resource.Revision, SourceSHA256: resource.SHA256,
		ExpectedTargetSHA256: converted.SHA256(),
	})
}

func (classifier *ProfileRecoveryInventoryClassifier) appendTarget(
	resource profilebackfillapp.SourceResource,
) error {
	if _, err := profileapp.DecodeDocument(
		classifier.target, resource.Document, resource.SHA256[:], classifier.normalizer,
	); err != nil {
		return profilebackfillapp.ErrInvalid
	}
	return classifier.accumulator.AppendTarget(resource)
}
