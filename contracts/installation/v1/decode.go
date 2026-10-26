package installationv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Parse(document []byte) (Manifest, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumBytes)
	if err != nil {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	var wire manifestWire
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	if err := requireEOF(decoder); err != nil || !validWire(wire) {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	environmentID, ok := parseUUID(wire.EnvironmentID)
	if !ok || requestDigest(wire) != wire.RequestSHA256 {
		return Manifest{}, fmt.Errorf("installation manifest is invalid")
	}
	digest := sha256.Sum256(canonical)
	return Manifest{wire: wire, envID: environmentID, canonical: canonical,
		sha256: hex.EncodeToString(digest[:])}, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}
