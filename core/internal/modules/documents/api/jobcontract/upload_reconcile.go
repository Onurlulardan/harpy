package jobcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

type UploadReference struct {
	Reference
	Generation uint64
}

func UploadReconcile(tenantID, uploadID ID,
	generation uint64,
) (string, string, error) {
	if !tenantID.Valid() || !uploadID.UUIDv7() || generation == 0 {
		return "", "", ErrInvalidID
	}
	document, err := json.Marshal(struct {
		UploadID   string `json:"upload_id"`
		Generation uint64 `json:"generation"`
	}{UploadID: uploadID.String(), Generation: generation})
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256(document)
	reference := prefix + "upload-reconcile/" + tenantID.String() + "/" +
		uploadID.String() + "/" + strconv.FormatUint(generation, 10)
	return reference, hex.EncodeToString(digest[:]), nil
}

func ParseUploadReconcile(value string) (UploadReference, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 6 || parts[0] != "documents" || parts[1] != "v1" ||
		parts[2] != "upload-reconcile" {
		return UploadReference{}, ErrInvalidID
	}
	tenantID, tenantErr := ParseID(parts[3])
	uploadID, uploadErr := ParseID(parts[4])
	generation, generationErr := strconv.ParseUint(parts[5], 10, 64)
	if tenantErr != nil || uploadErr != nil || generationErr != nil ||
		!uploadID.UUIDv7() || generation == 0 {
		return UploadReference{}, ErrInvalidID
	}
	return UploadReference{Reference: Reference{TenantID: tenantID,
		ResourceID: uploadID}, Generation: generation}, nil
}
