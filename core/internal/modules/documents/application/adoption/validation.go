package adoption

import (
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func validAuthority(value Authority) bool {
	return value.TenantID.Valid() && value.RunID.UUIDv7() &&
		value.CutoverGeneration > 0 && value.ManifestSHA256 != [32]byte{}
}

func validCheckpoint(value Checkpoint, authority Authority) bool {
	return value.RunID == authority.RunID && value.SourceItemID.UUIDv7() &&
		value.SourceVersionID.UUIDv7() && value.TargetItemID.UUIDv7() &&
		value.TargetVersionID.UUIDv7() && value.SourceObjectKey != "" &&
		len(value.SourceObjectKey) <= 1024 && value.SourceRevision != "" &&
		len(value.SourceRevision) <= 1024 && value.SourceBytes > 0 &&
		value.SourceSHA256 != [32]byte{} && value.SourceMetadataSHA256 != [32]byte{} &&
		validLSN(value.SourceLSN)
}

func validRun(value Run, authority Authority) bool {
	return value.Authority == authority && foundation.ValidInstant(value.CreatedAt) &&
		foundation.ValidInstant(value.UpdatedAt) && value.State != ""
}

func validLSN(value string) bool {
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	for _, part := range parts {
		if len(part) > 8 {
			return false
		}
		if _, err := strconv.ParseUint(part, 16, 32); err != nil {
			return false
		}
	}
	return strings.ToUpper(value) == value
}

func mapError(err error) error {
	if err == nil || err == ErrInvalid || err == ErrConflict || err == ErrBlocked {
		return err
	}
	return ErrUnavailable
}
