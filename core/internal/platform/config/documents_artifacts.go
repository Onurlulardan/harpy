package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
)

var errInvalidDocumentsArtifacts = errors.New("Documents artifacts are invalid")

type DocumentsArtifacts struct {
	storage      ArtifactS3
	keyring      DocumentsKEKKeyring
	billingActor uuid.UUID
	valid        bool
}

func loadDocumentsArtifacts(values environmentValues,
	environment Environment,
) (DocumentsArtifacts, error) {
	storage, err := loadArtifactS3(values, environment, documentsStorageKeys())
	if err != nil {
		return DocumentsArtifacts{}, err
	}
	keyring, err := loadDocumentsKEKKeyring(values)
	if err != nil {
		return DocumentsArtifacts{}, err
	}
	rawActor, found := values.lookup(documentsBillingActorKey)
	actor, actorErr := uuid.Parse(rawActor)
	if !found || actorErr != nil || actor.Version() != 7 || actor.String() != rawActor {
		return DocumentsArtifacts{}, configError(documentsBillingActorKey,
			"must be a canonical UUIDv7 IAM service principal")
	}
	return DocumentsArtifacts{storage: storage, keyring: keyring,
		billingActor: actor, valid: true}, nil
}

func (settings DocumentsArtifacts) Validate() error {
	if !settings.valid || settings.storage.Validate() != nil ||
		settings.keyring.Validate() != nil || settings.billingActor.Version() != 7 {
		return errInvalidDocumentsArtifacts
	}
	return nil
}

func (settings DocumentsArtifacts) BillingActorID() (uuid.UUID, error) {
	if err := settings.Validate(); err != nil {
		return uuid.Nil, err
	}
	return settings.billingActor, nil
}

func (settings DocumentsArtifacts) Storage() (ArtifactS3, error) {
	if err := settings.Validate(); err != nil {
		return ArtifactS3{}, err
	}
	return settings.storage, nil
}

func (settings DocumentsArtifacts) KEKKeyring() (DocumentsKEKKeyring, error) {
	if err := settings.Validate(); err != nil {
		return DocumentsKEKKeyring{}, err
	}
	return settings.keyring, nil
}

func (DocumentsArtifacts) String() string   { return "[REDACTED DOCUMENTS ARTIFACTS]" }
func (DocumentsArtifacts) GoString() string { return "[REDACTED DOCUMENTS ARTIFACTS]" }
func (DocumentsArtifacts) LogValue() slog.Value {
	return slog.StringValue("[REDACTED DOCUMENTS ARTIFACTS]")
}
func (DocumentsArtifacts) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED DOCUMENTS ARTIFACTS]")
}
