package systemtransferartifact

import (
	"crypto/sha256"
	"encoding/binary"
)

func exportUploadRequestSHA(
	key string,
	digest [32]byte,
	bytes int64,
) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("iam.system-transfer.export.upload.v1\x00"))
	writeDigestString(hash, key)
	hash.Write(digest[:])
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(bytes))
	hash.Write(encoded[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func exportUploadResultSHA(
	key, version, providerRequestID string,
	bytes int64,
) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("iam.system-transfer.export.upload-result.v1\x00"))
	writeDigestString(hash, key)
	writeDigestString(hash, version)
	writeDigestString(hash, providerRequestID)
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(bytes))
	hash.Write(encoded[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type digestWriter interface{ Write([]byte) (int, error) }

func writeDigestString(destination digestWriter, value string) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(len(value)))
	_, _ = destination.Write(encoded[:])
	_, _ = destination.Write([]byte(value))
}
