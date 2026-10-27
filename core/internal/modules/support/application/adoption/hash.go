package adoption

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func EmptySourceRoot() [32]byte { return sha256.Sum256(nil) }

func PageSHA256(records []Record) [32]byte {
	var document strings.Builder
	document.Grow(len(records) * 64)
	for _, record := range records {
		document.WriteString(record.SourceSHA256)
	}
	return sha256.Sum256([]byte(document.String()))
}

func FoldSHA256(previous, page [32]byte) [32]byte {
	return sha256.Sum256([]byte(hex.EncodeToString(previous[:]) + hex.EncodeToString(page[:])))
}

func canonicalDocument(value any, maximum int) (json.RawMessage, [32]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, [32]byte{}, ErrInvalid
	}
	document, err := canonicaljsonv1.Canonicalize(raw, maximum)
	if err != nil {
		return nil, [32]byte{}, ErrInvalid
	}
	return document, sha256.Sum256(document), nil
}

func hexSHA(value [32]byte) string { return hex.EncodeToString(value[:]) }
