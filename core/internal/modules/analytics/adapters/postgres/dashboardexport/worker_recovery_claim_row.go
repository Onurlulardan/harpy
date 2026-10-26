package dashboardexport

import (
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type recoveryClaimRow struct {
	disposition                                                   string
	generation                                                    *int64
	leaseUntil, effectStartedAt, expiresAt                        *time.Time
	format, objectKey, keyID, contentType, filename               *string
	wrappedDEK, wrapNonce, headerSHA, ciphertextSHA, plaintextSHA []byte
	ciphertextBytes, plaintextBytes, rows                         *int64
	cancellationRequested                                         *bool
}

func (stored *recoveryClaimRow) targets() []any {
	return []any{&stored.disposition, &stored.generation, &stored.leaseUntil,
		&stored.format, &stored.objectKey, &stored.keyID, &stored.wrappedDEK,
		&stored.wrapNonce, &stored.headerSHA, &stored.ciphertextSHA,
		&stored.plaintextSHA, &stored.ciphertextBytes, &stored.plaintextBytes,
		&stored.rows, &stored.contentType, &stored.filename,
		&stored.effectStartedAt, &stored.cancellationRequested, &stored.expiresAt}
}

func (stored recoveryClaimRow) decode(candidate exportapp.UnknownCandidate,
	cleanupOwner uuid.UUID,
) (exportapp.UnknownCleanupClaim, error) {
	claim := exportapp.UnknownCleanupClaim{Candidate: candidate,
		Disposition:         exportapp.UnknownClaimDisposition(stored.disposition),
		CleanupLeaseOwnerID: cleanupOwner}
	if claim.Disposition == exportapp.UnknownBusy ||
		claim.Disposition == exportapp.UnknownResolved {
		if stored.generation != nil || stored.leaseUntil != nil {
			return exportapp.UnknownCleanupClaim{}, exportapp.ErrUnavailable
		}
		return claim, nil
	}
	if claim.Disposition != exportapp.UnknownClaimed || !stored.complete() ||
		*stored.generation < 1 || !stored.leaseUntil.After(candidate.UnknownAt) ||
		stored.effectStartedAt.IsZero() || stored.expiresAt.IsZero() {
		return exportapp.UnknownCleanupClaim{}, exportapp.ErrUnavailable
	}
	draft, err := stored.draft(candidate.AttemptGeneration)
	if err != nil {
		return exportapp.UnknownCleanupClaim{}, err
	}
	claim.CleanupGeneration = *stored.generation
	claim.CleanupLeaseUntil = stored.leaseUntil.UTC()
	claim.Draft = &draft
	claim.CancellationRequested = *stored.cancellationRequested
	claim.ExpiresAt = stored.expiresAt.UTC()
	return claim, nil
}

func (stored recoveryClaimRow) complete() bool {
	return stored.generation != nil && stored.leaseUntil != nil &&
		stored.format != nil && stored.objectKey != nil && stored.keyID != nil &&
		stored.ciphertextBytes != nil && stored.plaintextBytes != nil &&
		stored.rows != nil && stored.contentType != nil && stored.filename != nil &&
		stored.effectStartedAt != nil && stored.cancellationRequested != nil &&
		stored.expiresAt != nil
}

func (stored recoveryClaimRow) draft(generation int64) (exportapp.ArtifactDraft, error) {
	format := exportapp.Format(*stored.format)
	if !format.Valid() || *stored.contentType != format.ContentType() ||
		*stored.objectKey == "" || *stored.keyID == "" ||
		len(stored.wrappedDEK) != 48 || len(stored.wrapNonce) != fcartv1.NonceBytes ||
		len(stored.headerSHA) != 32 || len(stored.ciphertextSHA) != 32 ||
		len(stored.plaintextSHA) != 32 || *stored.ciphertextBytes < 35 ||
		*stored.plaintextBytes < 1 || *stored.plaintextBytes > exportapp.MaximumPlaintextBytes ||
		*stored.rows < 0 || *stored.rows > exportapp.MaximumRows || *stored.filename == "" {
		return exportapp.ArtifactDraft{}, exportapp.ErrUnavailable
	}
	var nonce [fcartv1.NonceBytes]byte
	var header, ciphertext, plaintext [32]byte
	copy(nonce[:], stored.wrapNonce)
	copy(header[:], stored.headerSHA)
	copy(ciphertext[:], stored.ciphertextSHA)
	copy(plaintext[:], stored.plaintextSHA)
	return exportapp.ArtifactDraft{AttemptGeneration: generation,
		ObjectKey: *stored.objectKey,
		Envelope: fcartv1.KeyEnvelope{KEKID: *stored.keyID,
			Ciphertext: append([]byte(nil), stored.wrappedDEK...), Nonce: nonce},
		HeaderSHA256: header, CiphertextSHA256: ciphertext,
		PlaintextSHA256: plaintext, CiphertextBytes: *stored.ciphertextBytes,
		PlaintextBytes: *stored.plaintextBytes, Rows: *stored.rows,
		ContentType: *stored.contentType, Filename: *stored.filename}, nil
}
