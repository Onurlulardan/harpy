package platformops

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

const recoveryOperationNamespace = "harpy/platform-recovery-operation/v1"

func recoveryOperationID(environmentID, actorID uuid.UUID, kind, key string) uuid.UUID {
	hash := sha256.New()
	for _, part := range [][]byte{
		[]byte(recoveryOperationNamespace), environmentID[:], actorID[:],
		[]byte(kind), []byte(key),
	} {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write(part)
	}
	value := hash.Sum(nil)
	var result uuid.UUID
	copy(result[:], value[:16])
	clear(value)
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func recoveryRequestSHA256(value any) (string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(document)
	clear(document)
	return hex.EncodeToString(digest[:]), nil
}
