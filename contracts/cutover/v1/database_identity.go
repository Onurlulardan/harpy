package cutoverv1

import "strconv"

type DatabaseIdentity struct {
	DatabaseName     string `json:"database_name"`
	DatabaseOrigin   string `json:"database_origin"`
	SystemIdentifier string `json:"system_identifier"`
	DatabaseOID      uint32 `json:"database_oid"`
}

func ValidateDatabaseIdentity(identity DatabaseIdentity) error {
	if !sqlNamePattern.MatchString(identity.DatabaseName) ||
		(identity.DatabaseOrigin != "fresh" && identity.DatabaseOrigin != "v1_adopted") ||
		!validSystemIdentifier(identity.SystemIdentifier) || identity.DatabaseOID == 0 {
		return contractError("database identity")
	}
	return nil
}

func sameDatabaseIdentity(left, right DatabaseIdentity) bool {
	return left == right
}

type RelationIdentity struct {
	SchemaOID    uint32 `json:"schema_oid"`
	RelationOID  uint32 `json:"relation_oid"`
	SchemaName   string `json:"schema_name"`
	RelationName string `json:"relation_name"`
}

func validateRelationIdentity(identity RelationIdentity) error {
	if identity.SchemaOID == 0 || identity.RelationOID == 0 ||
		!sqlNamePattern.MatchString(identity.SchemaName) ||
		!sqlNamePattern.MatchString(identity.RelationName) {
		return contractError("relation identity")
	}
	return nil
}

func relationIdentityKey(identity RelationIdentity) string {
	return identity.SchemaName + "." + identity.RelationName + ":" +
		strconv.FormatUint(uint64(identity.SchemaOID), 10) + ":" +
		strconv.FormatUint(uint64(identity.RelationOID), 10)
}
