package domain

import "encoding/base32"

const TenantSchemaNameLength = 28

var tenantSchemaEncoding = base32.NewEncoding(
	"abcdefghijklmnopqrstuvwxyz234567",
).WithPadding(base32.NoPadding)

type SchemaName struct {
	value string
}

func SchemaNameForTenant(tenantID TenantID) (SchemaName, error) {
	if tenantID == (TenantID{}) {
		return SchemaName{}, ErrInvalidTenantID
	}
	return SchemaName{
		value: "t_" + tenantSchemaEncoding.EncodeToString(tenantID[:]),
	}, nil
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
