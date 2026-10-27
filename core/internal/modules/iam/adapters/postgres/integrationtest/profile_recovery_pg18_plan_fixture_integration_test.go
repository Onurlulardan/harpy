//go:build integration

package integrationtest

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"strconv"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func recoveryRootPlan(
	t *testing.T,
	planID uuid.UUID,
	source profileapp.Schema,
	target profileapp.Schema,
	mappings []profilebackfillapp.Mapping,
	rows []recoveryPlanRow,
	wrongPage int,
) profilebackfillapp.Plan {
	t.Helper()
	shared := recoveryMigrationMappings(mappings)
	pages := make([]profilebackfillapp.Page, 0, len(rows))
	var after uuid.UUID
	for index, row := range rows {
		selected, err := profileapp.SelectMigrationMappingsForSource(
			row.schema, target, shared,
		)
		if err != nil {
			t.Fatalf("select recovery root mappings for %s: %v", row.id, err)
		}
		sourceWire := row.document.CanonicalJSON()
		sourceSHA := row.document.SHA256()
		converted, err := profileapp.MigrateDocument(
			row.schema, target, selected, sourceWire, sourceSHA[:],
			recoveryTextNormalizer{},
		)
		if err != nil {
			t.Fatalf("convert recovery root row %s: %v", row.id, err)
		}
		targetSHA := converted.SHA256()
		sourcePage := recoveryDigest(
			[]byte(row.id.String()), []byte(row.schema.ID().String()),
			[]byte(strconv.FormatInt(row.revision, 10)), sourceSHA[:],
		)
		targetPage := recoveryDigest([]byte(row.id.String()), targetSHA[:])
		if index+1 == wrongPage {
			targetPage[sha256.Size-1] ^= 0xff
		}
		pages = append(pages, profilebackfillapp.Page{
			JobID: recoveryPlanJobID(planID, index+1),
			ScheduledRunAt: time.Date(
				2026, 7, 16, 16, 0, index, 0, time.UTC,
			),
			AfterResourceID: after, ThroughResourceID: row.id,
			ExpectedCount: 1, SourceSHA256: sourcePage,
			ExpectedTargetSHA256: targetPage,
		})
		after = row.id
	}
	definition := target.DefinitionSHA256()
	return profilebackfillapp.Plan{
		ID: planID, SourceSchemaID: source.ID(),
		Target: profilebackfillapp.TargetSchema{
			ID: target.ID(), ResourceKind: string(target.Resource()),
			Version: target.Version(), DefinitionJSON: target.DefinitionJSON(),
			DefinitionSHA256: definition, Fields: recoveryTargetFields(target),
		},
		ConverterVersion: profilebackfillapp.ConverterTypedMap,
		SourceCount:      int64(len(rows)), PageSize: 1,
		InitialRollingSHA256: sha256.Sum256(nil),
		Mappings:             mappings, Pages: pages,
	}
}

func recoveryMigrationMappings(
	mappings []profilebackfillapp.Mapping,
) []profileapp.MigrationMapping {
	result := make([]profileapp.MigrationMapping, 0, len(mappings))
	for _, mapping := range mappings {
		result = append(result, profileapp.MigrationMapping{
			SourceFieldID: mapping.SourceFieldID,
			TargetFieldID: mapping.TargetFieldID,
			Converter:     profileapp.MigrationConverter(mapping.ConverterKind),
		})
	}
	return result
}

func recoveryTargetFields(schema profileapp.Schema) []profilebackfillapp.TargetField {
	fields := schema.Fields()
	result := make([]profilebackfillapp.TargetField, 0, len(fields))
	for _, field := range fields {
		result = append(result, profilebackfillapp.TargetField{
			ID: field.ID(), Required: field.Required(),
			Active:         field.State() == profileapp.FieldActive,
			SemanticSHA256: field.SemanticSHA256(),
		})
	}
	return result
}

func recoveryPlanJobID(planID uuid.UUID, page int) uuid.UUID {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", planID, page)))
	var result uuid.UUID
	copy(result[:], digest[:len(result)])
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}

func recoveryDigest(values ...[]byte) [sha256.Size]byte {
	digest := sha256.New()
	writeRecoveryDigestFixture(digest, values...)
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeRecoveryDigestFixture(digest hash.Hash, values ...[]byte) {
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = digest.Write(size[:])
		_, _ = digest.Write(value)
	}
}
