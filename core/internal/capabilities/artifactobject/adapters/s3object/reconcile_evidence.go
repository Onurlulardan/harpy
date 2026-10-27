package s3object

import (
	"crypto/sha256"
	"encoding/binary"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

type reconcileDigestWriter interface{ Write([]byte) (int, error) }

func reconcileEvidenceSHA(
	expected artifactobject.ReconcileExpected,
	before, after reconciliationSnapshot,
	state artifactobject.ReconciliationState,
	headRequestID string,
) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("artifact-object.reconcile-exact.v1\x00"))
	reconcileDigestString(hash, expected.Key)
	hash.Write(expected.CiphertextSHA256[:])
	hash.Write(expected.HeaderSHA256[:])
	reconcileDigestInt64(hash, expected.CiphertextBytes)
	reconcileDigestString(hash, string(state))
	reconcileDigestSnapshot(hash, before)
	reconcileDigestString(hash, headRequestID)
	reconcileDigestSnapshot(hash, after)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func reconcileDigestSnapshot(hash reconcileDigestWriter, snapshot reconciliationSnapshot) {
	if snapshot.present {
		_, _ = hash.Write([]byte{1})
	} else {
		_, _ = hash.Write([]byte{0})
	}
	reconcileDigestString(hash, snapshot.version)
	reconcileDigestInt64(hash, snapshot.bytes)
	reconcileDigestInt64(hash, snapshot.lastModified.UnixNano())
	reconcileDigestString(hash, snapshot.requestID)
}

func reconcileDigestString(hash reconcileDigestWriter, value string) {
	reconcileDigestInt64(hash, int64(len(value)))
	_, _ = hash.Write([]byte(value))
}

func reconcileDigestInt64(hash reconcileDigestWriter, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	_, _ = hash.Write(encoded[:])
}
