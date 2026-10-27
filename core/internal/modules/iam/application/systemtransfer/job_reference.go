package systemtransfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func JobReference(resource publicresource.Resource, transferID uuid.UUID) (string, error) {
	if !resource.Valid() || transferID == uuid.Nil || transferID.Version() != 7 {
		return "", ErrInvalidReceipt
	}
	return fmt.Sprintf("iam-system-record/%s/%s", resource, transferID), nil
}

func JobReferenceSHA256(reference string) (string, error) {
	if _, _, err := ParseJobReference(reference); err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(reference))
	return hex.EncodeToString(digest[:]), nil
}

func ParseJobReference(value string) (publicresource.Resource, uuid.UUID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 3 || parts[0] != "iam-system-record" {
		return "", uuid.Nil, ErrInvalidReceipt
	}
	resource, ok := publicresource.ParseResource(parts[1])
	id, err := uuid.Parse(parts[2])
	if !ok || err != nil || id.Version() != 7 || id.String() != parts[2] {
		return "", uuid.Nil, ErrInvalidReceipt
	}
	return resource, id, nil
}
