package restoreorchestrationv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func canonical(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("restore orchestration encoding failed")
	}
	return canonicaljsonv1.CanonicalizeObject(raw, MaximumDocumentBytes)
}

func digest(document []byte) string {
	value := sha256.Sum256(document)
	return hex.EncodeToString(value[:])
}

func signatureDomain(purpose string, document []byte) []byte {
	prefix := []byte("harpy.restore-orchestrator." + purpose + ".v1")
	return append(append(prefix, 0), document...)
}
