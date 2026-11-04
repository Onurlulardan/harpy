package domain

import "encoding/hex"

const TenantSchemaNameLength = 39

type SchemaName struct {
	value string
}

func SchemaNameForTenant(tenantID TenantID) (SchemaName, error) {
	if tenantID == (TenantID{}) {
		return SchemaName{}, ErrInvalidTenantID
	}
	return SchemaName{value: "tenant_" + hex.EncodeToString(tenantID[:])}, nil
}

func SchemaNameFromUUID(value string) (SchemaName, error) {
	tenantID, err := ParseTenantID(value)
	if err != nil {
		return SchemaName{}, err
	}
	return SchemaNameForTenant(tenantID)
}

func (name SchemaName) String() string {
	return name.value
}
