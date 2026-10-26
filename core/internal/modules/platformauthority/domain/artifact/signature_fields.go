package artifact

import (
	"encoding/json"
)

func unsignedArtifact(document []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(document, &fields); err != nil {
		return nil, ErrInvalidDocument
	}
	if _, exists := fields["signature"]; !exists {
		return nil, ErrInvalidDocument
	}
	delete(fields, "signature")
	return canonicalValue(fields)
}

func unsignedTrustBundle(document []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(document, &fields); err != nil {
		return nil, ErrInvalidDocument
	}
	if _, digest := fields["bundle_sha256"]; !digest {
		return nil, ErrInvalidDocument
	}
	if _, signature := fields["release_signature"]; !signature {
		return nil, ErrInvalidDocument
	}
	delete(fields, "bundle_sha256")
	delete(fields, "release_signature")
	return canonicalValue(fields)
}
