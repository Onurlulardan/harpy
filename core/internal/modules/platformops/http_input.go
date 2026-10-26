package platformops

import (
	"encoding/base64"
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type recoveryClaimDocument struct {
	Claim struct {
		Algorithm          string `json:"algorithm"`
		PublicKeyBase64URL string `json:"public_key_base64url"`
	} `json:"claim"`
}

type recoveryKitAckDocument struct {
	ArtifactSHA256    string `json:"artifact_sha256"`
	AckProofBase64URL string `json:"ack_proof_base64url"`
}

func decodeRecoveryClaim(
	w http.ResponseWriter,
	r *http.Request,
) (recoveryClaimDocument, []byte, error) {
	var document recoveryClaimDocument
	if err := httpinput.DecodeJSON(w, r, 2048, &document); err != nil {
		return recoveryClaimDocument{}, nil, err
	}
	key, err := base64.RawURLEncoding.Strict().DecodeString(
		document.Claim.PublicKeyBase64URL)
	if err != nil || len(key) != 32 ||
		base64.RawURLEncoding.EncodeToString(key) != document.Claim.PublicKeyBase64URL ||
		document.Claim.Algorithm != platformopsapi.RecoveryClaimAlgorithm {
		clear(key)
		return recoveryClaimDocument{}, nil, httpinput.ErrInvalid
	}
	return document, key, nil
}

func decodeRecoveryKitAck(
	w http.ResponseWriter,
	r *http.Request,
) (recoveryKitAckDocument, []byte, error) {
	var document recoveryKitAckDocument
	if err := httpinput.DecodeJSON(w, r, 2048, &document); err != nil {
		return recoveryKitAckDocument{}, nil, err
	}
	proof, err := base64.RawURLEncoding.Strict().DecodeString(document.AckProofBase64URL)
	if err != nil || len(proof) < 32 || len(proof) > 96 ||
		base64.RawURLEncoding.EncodeToString(proof) != document.AckProofBase64URL ||
		!lowercaseSHA256(document.ArtifactSHA256) {
		clear(proof)
		return recoveryKitAckDocument{}, nil, httpinput.ErrInvalid
	}
	return document, proof, nil
}

func lowercaseSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') &&
			(character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
