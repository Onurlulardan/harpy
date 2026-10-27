package systemtransferartifact

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	"github.com/google/uuid"
)

type localOpenRequest struct {
	tenantID, artifactID, leaseOwner uuid.UUID
	generation                       int64
	format                           string
	object                           artifactobject.Expected
	keyID                            string
	wrappedDEK                       []byte
	wrapNonce                        [12]byte
	expectedPlaintextBytes           int64
	expectedPlaintextSHA256          *[32]byte
	keys                             *fcartv1.EnvelopeKeyring
	objects                          artifactobject.Store
}

func (runtime *Runtime) decryptLocal(
	ctx context.Context,
	request localOpenRequest,
) (lease *spoolapi.Lease, err error) {
	if !validLocalOpen(ctx, request) {
		return nil, errInvalid
	}
	lease, err = runtime.acquireSpool(request.artifactID, request.generation,
		request.leaseOwner, false)
	if err != nil {
		return nil, err
	}
	cleanupLease := lease
	keep := false
	defer func() {
		if !keep {
			err = errors.Join(err, cleanupLease.Discard())
		}
	}()
	writer, err := lease.OpenWriter()
	if err != nil {
		return nil, errors.Join(errUnavailable, err)
	}
	result, operationErr := decryptObject(ctx, request, writer)
	closeErr := writer.Close()
	if operationErr != nil || closeErr != nil ||
		int64(result.bytes) != request.expectedPlaintextBytes ||
		request.expectedPlaintextSHA256 != nil &&
			result.digest != *request.expectedPlaintextSHA256 {
		return nil, errors.Join(errIntegrity, operationErr, closeErr)
	}
	keep = true
	return lease, nil
}

type plaintextEvidence struct {
	bytes  uint64
	digest [32]byte
}

func decryptObject(ctx context.Context, request localOpenRequest,
	destination io.Writer) (plaintextEvidence, error) {
	identity := fcartv1.KeyIdentity{TenantID: request.tenantID,
		ArtifactID: request.artifactID, AttemptGeneration: uint64(request.generation)}
	envelope := fcartv1.KeyEnvelope{KEKID: request.keyID,
		Ciphertext: append([]byte(nil), request.wrappedDEK...), Nonce: request.wrapNonce}
	dek, err := request.keys.Unwrap(ctx, identity, envelope)
	if err != nil {
		return plaintextEvidence{}, errors.Join(errIntegrity, err)
	}
	defer clear(dek[:])
	format, err := tabulartransfer.ParseFormat(request.format)
	if err != nil {
		return plaintextEvidence{}, errors.Join(errIntegrity, err)
	}
	header, err := tabulartransfer.NewArtifactHeader(identity, format)
	if err != nil {
		return plaintextEvidence{}, errors.Join(errIntegrity, err)
	}
	reader, err := request.objects.OpenExact(ctx, request.object)
	if err != nil {
		return plaintextEvidence{}, errors.Join(errIntegrity, err)
	}
	hash := sha256.New()
	decrypted, decryptErr := tabulartransfer.Decrypt(io.MultiWriter(destination, hash),
		reader, fcartv1.CiphertextMetadata{Header: header,
			HeaderSHA256:     request.object.HeaderSHA256,
			CiphertextBytes:  uint64(request.object.CiphertextBytes),
			CiphertextSHA256: request.object.CiphertextSHA256}, dek)
	closeErr := reader.Close()
	var digest [32]byte
	copy(digest[:], hash.Sum(nil))
	return plaintextEvidence{bytes: decrypted.PlaintextBytes, digest: digest},
		errors.Join(decryptErr, closeErr)
}

func validLocalOpen(ctx context.Context, request localOpenRequest) bool {
	return runtimeInputContext(ctx) && request.tenantID != uuid.Nil &&
		request.artifactID != uuid.Nil && request.leaseOwner != uuid.Nil &&
		request.generation >= 1 && request.generation <= 3 &&
		(request.format == "csv" || request.format == "jsonl" || request.format == "xlsx") &&
		request.object.Identity.Key != "" && request.object.Identity.Version != "" &&
		request.object.CiphertextBytes > 0 && request.object.CiphertextSHA256 != [32]byte{} &&
		request.object.HeaderSHA256 != [32]byte{} && request.keyID != "" &&
		len(request.wrappedDEK) >= 32 && request.wrapNonce != [12]byte{} &&
		request.expectedPlaintextBytes > 0 && request.keys != nil &&
		!nilDependency(request.objects)
}

func runtimeInputContext(ctx context.Context) bool {
	return ctx != nil && ctx.Err() == nil
}
