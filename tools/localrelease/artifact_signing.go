package main

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func runtimeSignature(
	authority signingAuthority, document []byte, issuedAt time.Time,
) (runtimev1.AttestationSignature, error) {
	signature, err := authority.signer.SignAt(document, issuedAt)
	if err != nil {
		return runtimev1.AttestationSignature{}, fmt.Errorf("sign artifact: %w", err)
	}
	return runtimev1.AttestationSignature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}, nil
}

func receiptSignature(
	authority signingAuthority, contractVersion, kind string,
	payload any, issuedAt time.Time,
) (runtimev1.AttestationSignature, error) {
	envelope, err := attestation.MarshalCanonical(map[string]any{
		"contract_version": contractVersion, "kind": kind, "payload": payload,
	})
	if err != nil {
		return runtimev1.AttestationSignature{}, fmt.Errorf("encode receipt envelope: %w", err)
	}
	return runtimeSignature(authority, envelope, issuedAt)
}

func cutoverSignature(
	authority signingAuthority, document []byte, issuedAt time.Time,
) (cutoverv1.Signature, error) {
	signature, err := authority.signer.SignAt(document, issuedAt)
	if err != nil {
		return cutoverv1.Signature{}, fmt.Errorf("sign cutover artifact: %w", err)
	}
	return cutoverv1.Signature{
		Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
	}, nil
}
