package systemtransfer

import (
	"time"

	"github.com/google/uuid"
)

type rawAttempt struct {
	actorID                   uuid.UUID
	resource, duplicateAction string
	authorization             int64
	catalogVersionID          uuid.UUID
	catalogSHA, manifestSHA   []byte
	mapping                   []byte
	next, processed           int64
	sourceID                  uuid.UUID
	format                    string
	objectKey, objectVersion  string
	header, ciphertext        []byte
	ciphertextBytes           int64
	plaintext                 []byte
	plaintextBytes            int64
	keyID                     string
	wrappedDEK, wrapNonce     []byte
	providerRequestID         string
	uploadedAt                time.Time
	finished                  bool
}

func (raw *rawAttempt) targets() []any {
	return []any{
		&raw.actorID, &raw.resource, &raw.authorization, &raw.duplicateAction,
		&raw.mapping, &raw.next, &raw.processed, &raw.sourceID, &raw.format,
		&raw.objectKey, &raw.objectVersion, &raw.header, &raw.ciphertext,
		&raw.ciphertextBytes, &raw.plaintext, &raw.plaintextBytes, &raw.keyID,
		&raw.wrappedDEK, &raw.wrapNonce, &raw.providerRequestID, &raw.uploadedAt,
		&raw.finished, &raw.catalogVersionID, &raw.catalogSHA, &raw.manifestSHA,
	}
}

func workerDigest(value []byte) ([32]byte, bool) {
	var digest [32]byte
	if len(value) != len(digest) {
		return digest, false
	}
	copy(digest[:], value)
	return digest, true
}

func workerNonce(value []byte) ([12]byte, bool) {
	var nonce [12]byte
	if len(value) != len(nonce) {
		return nonce, false
	}
	copy(nonce[:], value)
	return nonce, true
}
