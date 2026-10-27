package jobcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
)

const referencePrefix = "work/v1/"

var ErrInvalid = errors.New("Work job reference is invalid")

type Reference struct {
	TenantID string
	TargetID string
}

func Build(kind, tenantID, targetID string) (string, string, error) {
	slug, valid := slugForKind(kind)
	if !valid || !validUUID(tenantID) || !validUUID(targetID) {
		return "", "", ErrInvalid
	}
	reference := referencePrefix + slug + "/" + tenantID + "/" + targetID
	document, err := json.Marshal(map[string]string{"target_id": targetID})
	if err != nil {
		return "", "", err
	}
	digest := sha256.Sum256(document)
	return reference, hex.EncodeToString(digest[:]), nil
}

func Parse(kind, value string) (Reference, error) {
	slug, valid := slugForKind(kind)
	parts := strings.Split(value, "/")
	if !valid || len(parts) != 5 || parts[0] != "work" || parts[1] != "v1" ||
		parts[2] != slug {
		return Reference{}, ErrInvalid
	}
	if !validUUID(parts[3]) || !validUUID(parts[4]) {
		return Reference{}, ErrInvalid
	}
	return Reference{TenantID: parts[3], TargetID: parts[4]}, nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := value[:8] + value[9:13] + value[14:18] + value[19:23] + value[24:]
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != 16 {
		return false
	}
	for _, value := range decoded {
		if value != 0 {
			return true
		}
	}
	return false
}

func slugForKind(kind string) (string, bool) {
	switch kind {
	case asyncmanifestv1.WorkProjectionRepairJobKind:
		return "projection-repair", true
	case asyncmanifestv1.WorkPurgeExecuteJobKind:
		return "purge-execute", true
	case asyncmanifestv1.WorkRecurrenceMaterializeJobKind:
		return "recurrence-materialize", true
	case asyncmanifestv1.WorkReminderDispatchJobKind:
		return "reminder-dispatch", true
	default:
		return "", false
	}
}
