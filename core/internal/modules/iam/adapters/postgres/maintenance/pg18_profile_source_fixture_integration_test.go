//go:build integration

package maintenance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const (
	maintenanceTargetFieldID = "46000000-0000-4000-8000-000000000001"
	maintenanceTargetSchemaA = "45000000-0000-4000-8000-000000000001"
	maintenanceTargetSchemaB = "45000000-0000-4000-8000-000000000002"
	maintenancePlanA         = "47000000-0000-4000-8000-000000000001"
	maintenancePlanB         = "47000000-0000-4000-8000-000000000002"
)

func seedMaintenancePG18ProfileSources(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
UPDATE iam_users SET profile_values=$2::jsonb,profile_sha256=$3::bytea,
 revision=revision+1,updated_at='2026-07-16T01:01:00Z'
WHERE id=$1::uuid AND profile_schema_version_id=$4::uuid AND revision=2`
	for tenantIndex, tenant := range maintenancePG18Tenants {
		for userIndex, userID := range maintenanceProfileUsers {
			value := maintenanceMixedProfileName(tenantIndex, userIndex)
			document, digest := maintenanceProfileDocument(maintenanceProfileField, value)
			execMaintenancePG18Tenant(t, ctx, database, tenant, statement,
				userID, document, digest[:], maintenanceUserSchemaID)
		}
	}
}

func newMaintenanceTargetSchema(
	t *testing.T, schemaID string, tenantIndex int,
) profileapp.Schema {
	t.Helper()
	retired := maintenanceField(t, maintenanceProfileField, "display_name",
		profileapp.FieldRetired, 64+tenantIndex*64)
	optional := maintenanceField(t, maintenanceOptionalFieldID, "department_code",
		profileapp.FieldActive, 64)
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID:       uuid.MustParse(maintenanceTargetFieldID),
		Resource: profileapp.ResourceUser, Key: "preferred_name",
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1,
		State:          profileapp.FieldActive,
		Validation:     profileapp.ValidationInput{MaximumBytes: 128},
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(schemaID), Resource: profileapp.ResourceUser,
		Version: 3, State: profileapp.SchemaActive,
		Fields: []profileapp.Field{retired, optional, field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func maintenanceProfileDocument(fieldID, value string) (string, [32]byte) {
	document := fmt.Sprintf(`{"%s":%q}`, fieldID, value)
	return document, sha256.Sum256([]byte(document))
}

func maintenancePageDigest(rows ...[][]byte) [32]byte {
	digest := sha256.New()
	for _, row := range rows {
		writeMaintenanceDigest(digest, row...)
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeMaintenanceDigest(digest hash.Hash, values ...[]byte) {
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = digest.Write(size[:])
		_, _ = digest.Write(value)
	}
}
