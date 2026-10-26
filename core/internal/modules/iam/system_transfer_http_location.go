package iam

import (
	"fmt"

	"github.com/google/uuid"
)

func systemExportLocation(resource SystemTransferResource, id string) string {
	return fmt.Sprintf("/api/iam/system-resources/%s/record-exports/%s", resource, id)
}

func systemImportSourceLocation(resource SystemTransferResource, id string) string {
	return fmt.Sprintf("/api/iam/system-resources/%s/record-import-sources/%s", resource, id)
}

func systemImportLocation(resource SystemTransferResource, id string) string {
	return fmt.Sprintf("/api/iam/system-resources/%s/record-imports/%s", resource, id)
}

func uuidV7(wire string) (uuid.UUID, error) {
	value, err := uuid.Parse(wire)
	if err != nil || value.Version() != 7 || value.Variant() != uuid.RFC4122 ||
		value.String() != wire {
		return uuid.Nil, ErrSystemTransferInvalid
	}
	return value, nil
}
