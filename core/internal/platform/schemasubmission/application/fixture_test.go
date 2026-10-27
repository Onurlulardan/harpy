package application

import (
	"encoding/base32"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/google/uuid"
)

const (
	fixtureTenantID   = "018f4f9e-7c21-7a3b-8c4d-123456789abc"
	fixtureResourceID = "018f4f9e-7c21-7a3b-8c4d-123456789abd"
	fixtureJobID      = "018f4f9e-7c21-7a3b-8c4d-123456789abe"
)

func fixturePlan() schemav1.Plan {
	return schemav1.Plan{
		ContractVersion: schemav1.PlanContractVersion,
		TenantID:        fixtureTenantID, Owner: "forms",
		Resource: schemav1.Resource{
			Kind: schemav1.ResourceFormStorage, ID: fixtureResourceID, Revision: 1,
		},
		Compiler: schemav1.Compiler{
			Version:         schemav1.InitialCompilerVersion,
			PostgreSQLMajor: schemav1.RequiredPostgreSQLMajor,
		},
		AllowedOperations: []schemav1.OperationID{schemav1.CreateFormTable},
		Steps: []schemav1.Step{{
			Ordinal: 0, OperationID: schemav1.CreateFormTable,
			TransactionClass: schemav1.Transactional,
			ExecutorKind:     schemav1.MigratorExecutor,
			IdentifierMapping: schemav1.IdentifierMapping{
				Form: schemav1.Identifier{
					LogicalID:    fixtureResourceID,
					PhysicalName: fixturePhysicalName("frm_", fixtureResourceID),
				},
				RelatedForms: []schemav1.Identifier{}, Fields: []schemav1.FieldIdentifier{},
				Relations: []schemav1.Identifier{}, Indexes: []schemav1.Identifier{},
				Constraints: []schemav1.Identifier{},
			},
			Arguments: schemav1.OperationArguments{},
		}},
	}
}

func fixturePhysicalName(prefix, identifier string) string {
	value := uuid.MustParse(identifier)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(value[:])
	return prefix + strings.ToLower(encoded)
}
