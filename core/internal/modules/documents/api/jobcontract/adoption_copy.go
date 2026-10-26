package jobcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

type AdoptionCopyReference struct {
	Reference
	SourceVersionID   ID
	CutoverGeneration uint64
	ManifestSHA256    [32]byte
}

func AdoptionCopy(tenantID, runID, sourceVersionID ID,
	cutoverGeneration uint64, manifestSHA256 [32]byte,
) (string, string, error) {
	if !tenantID.Valid() || !runID.UUIDv7() || !sourceVersionID.UUIDv7() ||
		cutoverGeneration == 0 || cutoverGeneration > math.MaxInt64 ||
		manifestSHA256 == [32]byte{} {
		return "", "", ErrInvalidID
	}
	document, err := json.Marshal(struct {
		RunID             string `json:"run_id"`
		SourceVersionID   string `json:"source_version_id"`
		CutoverGeneration uint64 `json:"cutover_generation"`
		ManifestSHA256    string `json:"manifest_sha256"`
	}{RunID: runID.String(), SourceVersionID: sourceVersionID.String(),
		CutoverGeneration: cutoverGeneration,
		ManifestSHA256:    hex.EncodeToString(manifestSHA256[:])})
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256(document)
	reference := prefix + "adoption-copy/" + tenantID.String() + "/" +
		runID.String() + "/" + sourceVersionID.String() + "/" +
		strconv.FormatUint(cutoverGeneration, 10) + "/" +
		hex.EncodeToString(manifestSHA256[:])
	return reference, hex.EncodeToString(digest[:]), nil
}

func ParseAdoptionCopy(value string) (AdoptionCopyReference, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 8 || parts[0] != "documents" || parts[1] != "v1" ||
		parts[2] != "adoption-copy" {
		return AdoptionCopyReference{}, ErrInvalidID
	}
	tenantID, tenantErr := ParseID(parts[3])
	runID, runErr := ParseID(parts[4])
	sourceVersionID, sourceErr := ParseID(parts[5])
	generation, generationErr := strconv.ParseUint(parts[6], 10, 64)
	manifest, manifestErr := hex.DecodeString(parts[7])
	if tenantErr != nil || runErr != nil || sourceErr != nil || generationErr != nil ||
		manifestErr != nil || !runID.UUIDv7() || !sourceVersionID.UUIDv7() ||
		generation == 0 || generation > math.MaxInt64 || len(manifest) != 32 {
		return AdoptionCopyReference{}, ErrInvalidID
	}
	result := AdoptionCopyReference{Reference: Reference{TenantID: tenantID,
		ResourceID: runID}, SourceVersionID: sourceVersionID,
		CutoverGeneration: generation}
	copy(result.ManifestSHA256[:], manifest)
	if result.ManifestSHA256 == [32]byte{} {
		return AdoptionCopyReference{}, ErrInvalidID
	}
	return result, nil
}
