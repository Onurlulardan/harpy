package idempotency

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
)

func digestResponse(response Response) string {
	digest := sha256.New()
	writeDigestPart(digest.Write, []byte("http-idempotency-response.v1"))
	var status [8]byte
	binary.BigEndian.PutUint64(status[:], uint64(response.status))
	writeDigestPart(digest.Write, status[:])
	names := make([]string, 0, len(response.headers))
	for name := range response.headers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		writeDigestPart(digest.Write, []byte(name))
		writeDigestPart(digest.Write, []byte(response.headers[name]))
	}
	if response.body != nil {
		writeDigestPart(digest.Write, []byte("body"))
		writeDigestPart(digest.Write, response.body)
	} else {
		writeDigestPart(digest.Write, []byte("reference"))
		writeDigestPart(digest.Write, []byte(*response.resultReference))
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func writeDigestPart(
	write func([]byte) (int, error),
	value []byte,
) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = write(size[:])
	_, _ = write(value)
}
