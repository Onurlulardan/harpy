package iam

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func profileRecoveryParent(
	t *testing.T,
	source profileapp.Schema,
	target profileapp.Schema,
) profilebackfillapp.RecoveryParent {
	t.Helper()
	mappings, err := profileBackfillMappings(source, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	definitionSHA := target.DefinitionSHA256()
	return profilebackfillapp.RecoveryParent{
		ID:         uuid.MustParse("15000000-0000-4000-8000-000000000001"),
		State:      profilebackfillapp.RecoveryParentFenced,
		PlanSHA256: sha256.Sum256([]byte("parent")), SourceSchemaID: source.ID(),
		Target: profilebackfillapp.TargetSchema{
			ID: target.ID(), ResourceKind: string(target.Resource()),
			Version: target.Version(), DefinitionJSON: target.DefinitionJSON(),
			DefinitionSHA256: definitionSHA, Fields: profileBackfillTargetFields(target),
		},
		ConverterVersion: profilebackfillapp.ConverterTypedMap, Mappings: mappings,
	}
}

func profileRecoveryResource(
	id string,
	schemaID uuid.UUID,
	document []byte,
) profilebackfillapp.SourceResource {
	return profilebackfillapp.SourceResource{
		ID: uuid.MustParse(id), SchemaVersionID: schemaID, Revision: 1,
		Document: document, SHA256: sha256.Sum256(document),
	}
}

func classifyProfileRecoveryFixture(
	t *testing.T,
	parent profilebackfillapp.RecoveryParent,
	source profileapp.Schema,
	target profileapp.Schema,
	schemas map[uuid.UUID]profileapp.Schema,
	resources []profilebackfillapp.SourceResource,
) (profilebackfillapp.RecoveryInventory, error) {
	t.Helper()
	classifier, err := NewProfileRecoveryInventoryClassifier(
		ProfileRecoveryInventoryInput{
			Parent: parent, Source: source, Target: target,
			Normalizer: profileHTTPNormalizer{},
			Clock:      func() time.Time { return time.Unix(1, 0).UTC() },
		},
	)
	if err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	if err := classifier.AppendBatch(context.Background(), schemas, resources); err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	return classifier.Finish(context.Background())
}

func profileRecoveryRetiredSchema(
	t *testing.T,
	id string,
	version int64,
	fields ...profileapp.Field,
) profileapp.Schema {
	t.Helper()
	if fields == nil {
		fields = []profileapp.Field{}
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(id), Resource: profileapp.ResourceUser,
		Version: version, State: profileapp.SchemaRetired, Fields: fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
