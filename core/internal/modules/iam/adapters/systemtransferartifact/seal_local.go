package systemtransferartifact

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type sealedDraft struct {
	lease           *spoolapi.Lease
	keyIdentity     fcartv1.KeyIdentity
	envelope        fcartv1.KeyEnvelope
	ciphertext      fcartv1.CiphertextMetadata
	plaintextBytes  int64
	plaintextSHA256 [32]byte
	objectKey       string
	analysis        systemtransfer.SourceAnalysis
}

type localSealRequest struct {
	tenantID, artifactID, leaseOwner uuid.UUID
	generation                       int64
	format, prefix                   string
	plaintext                        io.Reader
	maximumPlaintextBytes            int64
	stableReservation, analyze       bool
	keys                             *fcartv1.EnvelopeKeyring
}

func (runtime *Runtime) sealLocal(ctx context.Context,
	request localSealRequest) (draft sealedDraft, err error) {
	if ctx == nil || ctx.Err() != nil || request.tenantID == uuid.Nil ||
		request.artifactID == uuid.Nil || request.leaseOwner == uuid.Nil ||
		request.generation < 1 || request.generation > 3 || request.plaintext == nil ||
		request.maximumPlaintextBytes < 1 || request.keys == nil {
		return draft, errInvalid
	}
	lease, err := runtime.acquireSpool(request.artifactID, request.generation,
		request.leaseOwner, request.stableReservation)
	if err != nil {
		return draft, err
	}
	keep := false
	defer func() {
		if !keep {
			err = errors.Join(err, lease.Discard())
		}
	}()
	writer, err := lease.OpenWriter()
	if err != nil {
		return draft, errors.Join(errUnavailable, err)
	}
	identity := fcartv1.KeyIdentity{TenantID: request.tenantID,
		ArtifactID: request.artifactID, AttemptGeneration: uint64(request.generation)}
	dek, envelope, err := request.keys.Generate(ctx, identity)
	if err != nil {
		return draft, errors.Join(errUnavailable, err, writer.Close())
	}
	defer clear(dek[:])
	format, err := tabulartransfer.ParseFormat(request.format)
	if err != nil {
		return draft, errors.Join(errInvalid, err, writer.Close())
	}
	header, err := tabulartransfer.NewArtifactHeader(identity, format)
	if err != nil {
		return draft, errors.Join(errInvalid, err, writer.Close())
	}
	hash := sha256.New()
	capture := &analysisCapture{}
	evidence := io.Writer(hash)
	if request.analyze {
		evidence = io.MultiWriter(hash, capture)
	}
	limited := &io.LimitedReader{R: request.plaintext,
		N: request.maximumPlaintextBytes + 1}
	plaintext := io.TeeReader(limited, evidence)
	encrypted, encryptErr := tabulartransfer.Encrypt(writer, plaintext, header, dek)
	closeErr := writer.Close()
	if encryptErr != nil || closeErr != nil {
		return draft, errors.Join(errUnavailable, encryptErr, closeErr)
	}
	if encrypted.PlaintextBytes < 1 ||
		encrypted.PlaintextBytes > uint64(request.maximumPlaintextBytes) {
		return draft, errPlaintextLimit
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	analysis := systemtransfer.SourceAnalysis{}
	if request.analyze {
		captured, truncated := capture.snapshot()
		analysis, err = analyzeSource(request.format, captured, truncated)
		if err != nil {
			return draft, err
		}
	}
	draft = sealedDraft{lease: lease, keyIdentity: identity,
		envelope: envelope.OwnedCopy(), ciphertext: encrypted.CiphertextMetadata,
		plaintextBytes: int64(encrypted.PlaintextBytes), plaintextSHA256: digest,
		objectKey: fmt.Sprintf("%s/v1/%s/%d.fcart", request.prefix,
			request.artifactID.String(), request.generation), analysis: analysis}
	keep = true
	return draft, nil
}
