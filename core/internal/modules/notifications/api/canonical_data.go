package api

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
)

func CanonicalDataDigest(values []TemplateValue) ([32]byte, bool) {
	if len(values) > 64 {
		return [32]byte{}, false
	}
	copyValues := append([]TemplateValue(nil), values...)
	sort.Slice(copyValues, func(left, right int) bool {
		return copyValues[left].Key < copyValues[right].Key
	})
	hash := sha256.New()
	var total int
	for index, value := range copyValues {
		if value.Key == "" || len(value.Key) > 64 ||
			index > 0 && copyValues[index-1].Key == value.Key {
			return [32]byte{}, false
		}
		rendered, ok := value.Render()
		total += len(value.Key) + len(rendered)
		if !ok || total > 16*1024 {
			return [32]byte{}, false
		}
		writeCanonicalPart(hash, value.Key)
		writeCanonicalPart(hash, string(value.Kind))
		writeCanonicalPart(hash, rendered)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, true
}

type canonicalWriter interface{ Write([]byte) (int, error) }

func writeCanonicalPart(writer canonicalWriter, value string) {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write([]byte(value))
}
