package cutoverv1

import "unicode/utf8"

type ColumnMapping struct {
	Ordinal      int    `json:"ordinal"`
	SourceAttnum int    `json:"source_attnum"`
	TargetAttnum int    `json:"target_attnum"`
	TransformID  string `json:"transform_id"`
	PrimaryKey   bool   `json:"primary_key"`
}

type RelationMapping struct {
	Ordinal                 int              `json:"ordinal"`
	Source                  RelationIdentity `json:"source"`
	Target                  RelationIdentity `json:"target"`
	ReplicaIdentity         string           `json:"replica_identity"`
	Columns                 []ColumnMapping  `json:"columns"`
	ProcedureContractID     string           `json:"procedure_contract_id"`
	ApplyProcedureOID       uint32           `json:"apply_procedure_oid"`
	ApplyProcedureSignature string           `json:"apply_procedure_signature"`
	ExpectedCheckpointKeys  []string         `json:"expected_checkpoint_keys"`
}

func validateRelationMappings(
	relations []RelationMapping,
	sourceDatabase DatabaseIdentity,
	targetDatabase DatabaseIdentity,
) error {
	if len(relations) == 0 || len(relations) > 4096 {
		return contractError("relation allowlist")
	}
	sources := make(map[string]struct{}, len(relations))
	targets := make(map[string]struct{}, len(relations))
	totalCheckpointKeys := 0
	for index, relation := range relations {
		if relation.Ordinal != index+1 || validateRelationIdentity(relation.Source) != nil ||
			validateRelationIdentity(relation.Target) != nil {
			return contractError("relation ordinal or identity")
		}
		sourceKey := relationIdentityKey(relation.Source)
		targetKey := relationIdentityKey(relation.Target)
		if _, duplicate := sources[sourceKey]; duplicate {
			return contractError("duplicate source relation")
		}
		if _, duplicate := targets[targetKey]; duplicate {
			return contractError("duplicate target relation")
		}
		if sameDatabaseIdentity(sourceDatabase, targetDatabase) && sourceKey == targetKey {
			return contractError("active and inactive relation identity")
		}
		sources[sourceKey], targets[targetKey] = struct{}{}, struct{}{}
		if err := validateRelationMapping(relation); err != nil {
			return err
		}
		totalCheckpointKeys += len(relation.ExpectedCheckpointKeys)
		if totalCheckpointKeys > 10000 {
			return contractError("checkpoint key set")
		}
	}
	return nil
}

func validateRelationMapping(relation RelationMapping) error {
	if relation.ReplicaIdentity != "primary_key" && relation.ReplicaIdentity != "index" {
		return contractError("replica identity")
	}
	if !transformIDPattern.MatchString(relation.ProcedureContractID) ||
		relation.ApplyProcedureOID == 0 ||
		!applyProcedurePattern.MatchString(relation.ApplyProcedureSignature) {
		return contractError("apply procedure")
	}
	if len(relation.Columns) == 0 || len(relation.Columns) > 2048 {
		return contractError("column allowlist")
	}
	sourceColumns := make(map[int]struct{}, len(relation.Columns))
	targetColumns := make(map[int]struct{}, len(relation.Columns))
	primaryKey := false
	for index, column := range relation.Columns {
		if column.Ordinal != index+1 || column.SourceAttnum < 1 || column.SourceAttnum > 32767 ||
			column.TargetAttnum < 1 || column.TargetAttnum > 32767 ||
			!transformIDPattern.MatchString(column.TransformID) {
			return contractError("column mapping")
		}
		if _, duplicate := sourceColumns[column.SourceAttnum]; duplicate {
			return contractError("duplicate source column")
		}
		if _, duplicate := targetColumns[column.TargetAttnum]; duplicate {
			return contractError("duplicate target column")
		}
		sourceColumns[column.SourceAttnum], targetColumns[column.TargetAttnum] = struct{}{}, struct{}{}
		primaryKey = primaryKey || column.PrimaryKey
	}
	if !primaryKey {
		return contractError("relation primary key")
	}
	if err := validateExpectedCheckpointKeys(relation.ExpectedCheckpointKeys); err != nil {
		return err
	}
	return nil
}

func validateExpectedCheckpointKeys(keys []string) error {
	if len(keys) == 0 || len(keys) > 1024 {
		return contractError("checkpoint key set")
	}
	for index, key := range keys {
		if len(key) == 0 || len(key) > 512 || !utf8.ValidString(key) || containsControl(key) ||
			(index > 0 && keys[index-1] >= key) {
			return contractError("checkpoint key order")
		}
	}
	return nil
}

func containsControl(value string) bool {
	for _, character := range []byte(value) {
		if character < 0x20 || character == 0x7f {
			return true
		}
	}
	return false
}
