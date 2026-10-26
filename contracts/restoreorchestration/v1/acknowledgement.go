package restoreorchestrationv1

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func BuildAcknowledgement(
	input AcknowledgementInput,
	trust Trust,
	private ed25519.PrivateKey,
) (Acknowledgement, error) {
	if len(private) != ed25519.PrivateKeySize ||
		!private.Public().(ed25519.PublicKey).Equal(trust.publicKey) {
		return Acknowledgement{}, fmt.Errorf(
			"restore orchestrator acknowledgement signer is invalid",
		)
	}
	wire := acknowledgementWire{
		Schema: AcknowledgementSchema, Purpose: input.Purpose,
		EnvironmentID: input.EnvironmentID, OperationID: input.OperationID,
		AttemptID: input.AttemptID, PlanSHA256: input.PlanSHA256,
		Sequence: input.Sequence, IssuedAt: input.IssuedAt.Format(timeFormat),
		ExpiresAt:                     input.ExpiresAt.Format(timeFormat),
		PreviousDatabaseSystemID:      input.PreviousDatabaseSystemID,
		NewDatabaseSystemID:           input.NewDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: input.PreviousDatabaseLocatorSHA256,
		NewDatabaseLocatorSHA256:      input.NewDatabaseLocatorSHA256,
		StoppedServices:               append([]string(nil), input.StoppedServices...),
		SwitchOutcome:                 input.SwitchOutcome, KeyGeneration: trust.Generation(),
		KeyID: trust.KeyID(), Algorithm: Algorithm,
	}
	if !validAcknowledgement(wire, trust, input.IssuedAt) {
		return Acknowledgement{}, fmt.Errorf(
			"restore orchestrator acknowledgement input is invalid",
		)
	}
	unsigned, err := canonical(wire)
	if err != nil {
		return Acknowledgement{}, err
	}
	wire.Signature = base64.RawURLEncoding.EncodeToString(
		ed25519.Sign(private, signatureDomain(input.Purpose, unsigned)),
	)
	document, err := canonical(wire)
	if err != nil {
		return Acknowledgement{}, err
	}
	return ParseAcknowledgement(document, trust, Expected{
		Purpose: input.Purpose, EnvironmentID: input.EnvironmentID,
		OperationID: input.OperationID, AttemptID: input.AttemptID,
		PlanSHA256:                    input.PlanSHA256,
		PreviousDatabaseSystemID:      input.PreviousDatabaseSystemID,
		NewDatabaseSystemID:           input.NewDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: input.PreviousDatabaseLocatorSHA256,
		NewDatabaseLocatorSHA256:      input.NewDatabaseLocatorSHA256,
	}, input.IssuedAt)
}

func ParseAcknowledgement(
	document []byte,
	trust Trust,
	expected Expected,
	now time.Time,
) (Acknowledgement, error) {
	canonicalValue, err := canonicalDocument(document)
	if err != nil {
		return Acknowledgement{}, err
	}
	var wire acknowledgementWire
	decoder := json.NewDecoder(bytes.NewReader(canonicalValue))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&wire) != nil || decoder.Decode(new(any)) != io.EOF ||
		!validAcknowledgement(wire, trust, now) ||
		!matchesExpected(wire, expected) {
		return Acknowledgement{}, fmt.Errorf(
			"restore orchestrator acknowledgement is invalid",
		)
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(wire.Signature)
	unsigned := wire
	unsigned.Signature = ""
	unsignedDocument, encodeErr := canonical(unsigned)
	if err != nil || len(signature) != ed25519.SignatureSize ||
		encodeErr != nil || !ed25519.Verify(
		trust.publicKey, signatureDomain(wire.Purpose, unsignedDocument), signature,
	) {
		return Acknowledgement{}, fmt.Errorf(
			"restore orchestrator acknowledgement signature is invalid",
		)
	}
	issuedAt, _ := time.Parse(timeFormat, wire.IssuedAt)
	expiresAt, _ := time.Parse(timeFormat, wire.ExpiresAt)
	return Acknowledgement{
		wire: wire, canonical: canonicalValue, digest: digest(canonicalValue),
		issuedAt: issuedAt, expiresAt: expiresAt,
	}, nil
}
