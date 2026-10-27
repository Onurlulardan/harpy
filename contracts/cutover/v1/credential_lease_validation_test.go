package cutoverv1

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateCredentialLeaseAcceptsExactRegistrySemantics(t *testing.T) {
	lease := validTypedCredentialLease()
	if err := ValidateCredentialLease(lease); err != nil {
		t.Fatalf("valid tenant lease rejected: %v", err)
	}
	billingForms := validTypedCredentialLease()
	billingForms.GroupID = "billing"
	billingForms.Domains = []string{"billing", "forms"}
	if err := ValidateCredentialLease(billingForms); err != nil {
		t.Fatalf("valid billing lease rejected: %v", err)
	}
	globalBilling := validTypedCredentialLease()
	globalBilling.ScopeKind = "global"
	globalBilling.ScopeID = globalBilling.EnvironmentID
	globalBilling.GroupID = "billing"
	globalBilling.Domains = []string{"billing"}
	if err := ValidateCredentialLease(globalBilling); err != nil {
		t.Fatalf("valid global billing lease rejected: %v", err)
	}
}

func TestValidateCredentialLeaseRejectsSemanticDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CredentialLease)
	}{
		{"window exceeds fifteen minutes", func(v *CredentialLease) { v.ExpiresAt = "2026-07-13T12:15:00.000000001Z" }},
		{"lease issued too early", func(v *CredentialLease) { v.IssuedAt = "2026-07-13T11:59:59Z" }},
		{"issue after not-before", func(v *CredentialLease) { v.IssuedAt = "2026-07-13T12:00:01Z" }},
		{"duplicate authority ID", func(v *CredentialLease) { v.OperationID = v.RunID }},
		{"database identity drift", func(v *CredentialLease) { v.Target.DatabaseOID++ }},
		{"decoder limit drift", func(v *CredentialLease) { v.RoleLeases[0].ConnectionLimit = 3 }},
		{"adopter replication", func(v *CredentialLease) { v.RoleLeases[1].Replication = true }},
		{"partial billing forms", func(v *CredentialLease) {
			v.GroupID, v.Domains = "billing", []string{"billing"}
		}},
		{"reordered billing forms", func(v *CredentialLease) {
			v.GroupID, v.Domains = "billing", []string{"forms", "billing"}
		}},
		{"tenant billing singleton", func(v *CredentialLease) {
			v.GroupID, v.Domains = "billing", []string{"billing"}
		}},
		{"legacy iam command group", func(v *CredentialLease) {
			v.GroupID, v.Domains = "iam", []string{"iam"}
		}},
		{"missing primary key", func(v *CredentialLease) { v.Relations[0].Columns[0].PrimaryKey = false }},
		{"missing checkpoint plan", func(v *CredentialLease) {
			v.Relations[0].ExpectedCheckpointKeys = nil
		}},
		{"duplicate checkpoint key", func(v *CredentialLease) {
			v.Relations[0].ExpectedCheckpointKeys = []string{"main", "main"}
		}},
		{"unordered checkpoint keys", func(v *CredentialLease) {
			v.Relations[0].ExpectedCheckpointKeys = []string{"z", "a"}
		}},
		{"checkpoint key control byte", func(v *CredentialLease) {
			v.Relations[0].ExpectedCheckpointKeys = []string{"main\n"}
		}},
		{"duplicate source column", func(v *CredentialLease) {
			v.Relations[0].Columns = append(v.Relations[0].Columns, ColumnMapping{
				Ordinal: 2, SourceAttnum: 1, TargetAttnum: 2,
				TransformID: "identity.uuid", PrimaryKey: false,
			})
		}},
		{"duplicate source relation", func(v *CredentialLease) {
			second := v.Relations[0]
			second.Ordinal = 2
			second.Target.RelationOID++
			v.Relations = append(v.Relations, second)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lease := validTypedCredentialLease()
			test.mutate(&lease)
			if err := ValidateCredentialLease(lease); err == nil {
				t.Fatal("unsafe credential lease accepted")
			}
		})
	}
}

func TestValidateCredentialLeaseRejectsMoreThanTenThousandCheckpointKeys(t *testing.T) {
	lease := validTypedCredentialLease()
	lease.Relations = make([]RelationMapping, 10)
	for relationIndex := range lease.Relations {
		relation := validTypedRelation()
		relation.Ordinal = relationIndex + 1
		relation.Source.RelationOID += uint32(relationIndex)
		relation.Target.RelationOID += uint32(relationIndex)
		relation.ExpectedCheckpointKeys = make([]string, 1001)
		for keyIndex := range relation.ExpectedCheckpointKeys {
			relation.ExpectedCheckpointKeys[keyIndex] = fmt.Sprintf("%04d", keyIndex)
		}
		lease.Relations[relationIndex] = relation
	}
	if err := ValidateCredentialLease(lease); err == nil {
		t.Fatal("unbounded checkpoint plan accepted")
	}
}

func TestValidateCredentialLeaseSignatureIsExact(t *testing.T) {
	value := CredentialLeaseSignature{
		ContractVersion: "cutover-credential-lease-signature.v1",
		ManifestSHA256:  strings.Repeat("a", 64), Signature: validTypedSignature(),
	}
	if err := ValidateCredentialLeaseSignature(value); err != nil {
		t.Fatalf("valid lease signature rejected: %v", err)
	}
	value.Signature.Value += "="
	if err := ValidateCredentialLeaseSignature(value); err == nil {
		t.Fatal("non-canonical lease signature accepted")
	}
}
