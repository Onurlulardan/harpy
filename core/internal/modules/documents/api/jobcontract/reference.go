package jobcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const prefix = "documents/v1/"

type Reference struct {
	TenantID, ResourceID ID
}

type PurgeReference struct {
	Reference
	AfterDeletionID ID
}

func Archive(tenantID, archiveID ID) (string, string, error) {
	return build("archive", tenantID, archiveID, "archive_id")
}

func Purge(tenantID, itemID ID) (string, string, error) {
	return PurgePage(tenantID, itemID, ID{})
}

func Deletion(tenantID, deletionID ID) (string, string, error) {
	return build("deletion", tenantID, deletionID, "deletion_id")
}

func ParseArchive(value string) (Reference, error)  { return parse("archive", value) }
func ParseDeletion(value string) (Reference, error) { return parse("deletion", value) }

func PurgePage(tenantID, itemID, afterDeletionID ID) (string, string, error) {
	if !tenantID.Valid() || !itemID.UUIDv7() ||
		afterDeletionID.Valid() && !afterDeletionID.UUIDv7() {
		return "", "", ErrInvalidID
	}
	reference := prefix + "purge/" + tenantID.String() + "/" + itemID.String()
	payload := map[string]string{"item_id": itemID.String()}
	if afterDeletionID.Valid() {
		reference += "/after/" + afterDeletionID.String()
		payload["after_deletion_id"] = afterDeletionID.String()
	}
	document, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256(document)
	return reference, hex.EncodeToString(digest[:]), nil
}

func ParsePurge(value string) (PurgeReference, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 5 && len(parts) != 7 || len(parts) >= 3 &&
		(parts[0] != "documents" || parts[1] != "v1" || parts[2] != "purge") {
		return PurgeReference{}, ErrInvalidID
	}
	base, err := parse("purge", strings.Join(parts[:5], "/"))
	if err != nil {
		return PurgeReference{}, err
	}
	result := PurgeReference{Reference: base}
	if len(parts) == 7 {
		result.AfterDeletionID, err = ParseID(parts[6])
		if parts[5] != "after" || err != nil || !result.AfterDeletionID.UUIDv7() {
			return PurgeReference{}, ErrInvalidID
		}
	}
	return result, nil
}

func build(kind string, tenantID, resourceID ID, field string) (string, string, error) {
	if !tenantID.Valid() || !resourceID.UUIDv7() {
		return "", "", ErrInvalidID
	}
	reference := prefix + kind + "/" + tenantID.String() + "/" + resourceID.String()
	document, err := json.Marshal(map[string]string{field: resourceID.String()})
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256(document)
	return reference, hex.EncodeToString(digest[:]), nil
}

func parse(kind, value string) (Reference, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 5 || parts[0] != "documents" || parts[1] != "v1" ||
		parts[2] != kind {
		return Reference{}, ErrInvalidID
	}
	tenantID, tenantErr := ParseID(parts[3])
	resourceID, resourceErr := ParseID(parts[4])
	if tenantErr != nil || resourceErr != nil || !resourceID.UUIDv7() {
		return Reference{}, ErrInvalidID
	}
	return Reference{TenantID: tenantID, ResourceID: resourceID}, nil
}
