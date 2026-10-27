package restoreorchestrationv1

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

func BuildTrust(input TrustInput) (Trust, error) {
	wire := trustWire{
		Schema: TrustSchema, EnvironmentID: input.EnvironmentID,
		DeploymentBindingSHA256: input.DeploymentBindingSHA256,
		Generation:              input.Generation, KeyID: input.KeyID, Algorithm: Algorithm,
		PublicKey: base64.RawURLEncoding.EncodeToString(input.PublicKey),
	}
	document, err := canonical(wire)
	if err != nil {
		return Trust{}, err
	}
	return ParseTrust(document)
}

func ParseTrust(document []byte) (Trust, error) {
	canonicalDocument, err := canonicalDocument(document)
	if err != nil {
		return Trust{}, err
	}
	var wire trustWire
	decoder := json.NewDecoder(bytes.NewReader(canonicalDocument))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&wire) != nil || decoder.Decode(new(any)) != io.EOF ||
		!validTrust(wire) {
		return Trust{}, fmt.Errorf("restore orchestrator trust is invalid")
	}
	public, _ := base64.RawURLEncoding.Strict().DecodeString(wire.PublicKey)
	return Trust{
		wire: wire, canonical: canonicalDocument, digest: digest(canonicalDocument),
		publicKey: ed25519.PublicKey(public),
	}, nil
}

func canonicalDocument(document []byte) ([]byte, error) {
	if len(document) == 0 || len(document) > MaximumDocumentBytes {
		return nil, fmt.Errorf("restore orchestration document is invalid")
	}
	var value any
	if json.Unmarshal(document, &value) != nil {
		return nil, fmt.Errorf("restore orchestration document is invalid")
	}
	canonicalValue, err := canonical(value)
	if err != nil || !bytes.Equal(canonicalValue, document) {
		return nil, fmt.Errorf("restore orchestration document is invalid")
	}
	return canonicalValue, nil
}
