package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

const documentsArchiveReservationKey = "HARPY_DOCUMENTS_ARCHIVE_RESERVATION_TTL"

var errInvalidDocumentsWorker = errors.New("Documents worker config is invalid")

type DocumentsWorker struct {
	artifacts     DocumentsArtifacts
	legacy        ArtifactS3
	archiveTTL    time.Duration
	clamAVAddress string
	valid         bool
}

func loadDocumentsWorker(values environmentValues,
	environment Environment,
) (DocumentsWorker, error) {
	artifacts, err := loadDocumentsArtifacts(values, environment)
	if err != nil {
		return DocumentsWorker{}, err
	}
	legacy, err := loadArtifactS3(values, environment, documentsLegacyStorageKeys())
	if err != nil {
		return DocumentsWorker{}, err
	}
	if _, found := values.lookup(documentsArchiveReservationKey); !found {
		return DocumentsWorker{}, configError(documentsArchiveReservationKey, "is required")
	}
	archiveTTL, err := parseDuration(values, documentsArchiveReservationKey, "",
		2*time.Hour, 48*time.Hour)
	if err != nil {
		return DocumentsWorker{}, err
	}
	address, found := values.lookup(documentsClamAVAddressKey)
	if !found || !validClamAVAddress(address) {
		return DocumentsWorker{}, configError(documentsClamAVAddressKey,
			"must be a canonical host and port")
	}
	target, targetErr := artifacts.Storage()
	if targetErr != nil || sameArtifactBucket(target, legacy) {
		return DocumentsWorker{}, configError(documentsLegacyBucketKey,
			"must differ from the Documents target bucket")
	}
	return DocumentsWorker{artifacts: artifacts, legacy: legacy,
		archiveTTL: archiveTTL, clamAVAddress: address, valid: true}, nil
}

func (settings DocumentsWorker) Validate() error {
	if !settings.valid || settings.artifacts.Validate() != nil ||
		settings.legacy.Validate() != nil ||
		settings.archiveTTL < 2*time.Hour || settings.archiveTTL > 48*time.Hour ||
		!validClamAVAddress(settings.clamAVAddress) {
		return errInvalidDocumentsWorker
	}
	return nil
}

func (settings DocumentsWorker) ArchiveReservationTTL() (time.Duration, error) {
	if err := settings.Validate(); err != nil || settings.archiveTTL < 2*time.Hour ||
		settings.archiveTTL > 48*time.Hour {
		return 0, errInvalidDocumentsWorker
	}
	return settings.archiveTTL, nil
}

func (settings DocumentsWorker) Artifacts() (DocumentsArtifacts, error) {
	if err := settings.Validate(); err != nil {
		return DocumentsArtifacts{}, err
	}
	return settings.artifacts, nil
}

func (settings DocumentsWorker) LegacySource() (ArtifactS3, error) {
	if err := settings.Validate(); err != nil {
		return ArtifactS3{}, err
	}
	return settings.legacy, nil
}

func (settings DocumentsWorker) ClamAVAddress() (string, error) {
	if err := settings.Validate(); err != nil {
		return "", err
	}
	return settings.clamAVAddress, nil
}

func sameArtifactBucket(left, right ArtifactS3) bool {
	leftEndpoint, _, leftBucket, leftOwner, _, leftErr := left.Connection()
	rightEndpoint, _, rightBucket, rightOwner, _, rightErr := right.Connection()
	return leftErr == nil && rightErr == nil && leftEndpoint == rightEndpoint &&
		leftBucket == rightBucket && leftOwner == rightOwner
}

func (DocumentsWorker) String() string   { return "[REDACTED DOCUMENTS WORKER CONFIG]" }
func (DocumentsWorker) GoString() string { return "[REDACTED DOCUMENTS WORKER CONFIG]" }
func (DocumentsWorker) LogValue() slog.Value {
	return slog.StringValue("[REDACTED DOCUMENTS WORKER CONFIG]")
}
func (DocumentsWorker) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED DOCUMENTS WORKER CONFIG]")
}
