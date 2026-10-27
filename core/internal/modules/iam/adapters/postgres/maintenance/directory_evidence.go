package maintenance

import (
	"encoding/hex"
	"encoding/json"
)

type directoryRepairEvidence struct {
	KeyID string `json:"key_id"`
	HMAC  string `json:"hmac"`
}

type directoryCountEvidence struct {
	KeyID string `json:"key_id"`
	Count int64  `json:"count"`
}

func marshalDirectoryRepairEvidence(
	keys []DirectoryKey,
	normalized []byte,
	active bool,
) (string, error) {
	entries := make([]directoryRepairEvidence, 0, len(keys))
	if active {
		for _, key := range keys {
			value := directoryLookupHMAC(key, normalized)
			entries = append(entries, directoryRepairEvidence{
				KeyID: key.ID,
				HMAC:  hex.EncodeToString(value[:]),
			})
			clear(value[:])
		}
	}
	return marshalDirectoryEvidence(entries)
}

func marshalDirectoryCountEvidence(
	keys []DirectoryKey,
	counts map[string]int64,
) (string, error) {
	ids, values := directoryKeyEvidence(keys, counts)
	if len(ids) < 1 || len(ids) > 65 {
		return "", ErrRuntimeInvalid
	}
	entries := make([]directoryCountEvidence, len(ids))
	for index := range ids {
		entries[index] = directoryCountEvidence{KeyID: ids[index], Count: values[index]}
	}
	return marshalDirectoryEvidence(entries)
}

func marshalDirectoryEvidence(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encoded := string(payload)
	clear(payload)
	return encoded, nil
}
