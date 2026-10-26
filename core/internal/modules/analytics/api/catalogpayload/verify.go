package catalogpayload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const (
	MaximumPayloadBytes     = 4 << 20
	maximumNameBytes        = 100
	maximumDescriptionBytes = 2048
)

var (
	uuidV7Pattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
	slugPattern   = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type Payload struct {
	ContractVersion string          `json:"contract_version"`
	ComponentID     string          `json:"component_id"`
	AnalyticsID     string          `json:"analytics_id"`
	Kind            string          `json:"kind"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	LegacyHash      string          `json:"legacy_hash"`
	SourceContract  string          `json:"source_contract"`
	Definition      json.RawMessage `json:"definition"`
}

type Verified struct {
	payload  Payload
	document []byte
}

func Verify(document []byte, componentID, expectedSHA256 string) (Verified, error) {
	if err := catalogpackagev1.ValidateCanonicalPayload(
		document, expectedSHA256, MaximumPayloadBytes,
	); err != nil {
		return Verified{}, fmt.Errorf("analytics catalog payload is invalid")
	}
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		validate(payload, componentID) != nil {
		return Verified{}, fmt.Errorf("analytics catalog payload contract is invalid")
	}
	return Verified{payload: clone(payload), document: append([]byte(nil), document...)}, nil
}

func validate(payload Payload, componentID string) error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		payload.Definition, MaximumPayloadBytes,
	)
	_, sourceErr := catalogpackagev1.ParseFirstPartySourceContract(
		payload.SourceContract,
	)
	if payload.ContractVersion != catalogpackagev1.ContractAnalytics ||
		payload.ComponentID != componentID || payload.AnalyticsID != componentID ||
		!uuidV7Pattern.MatchString(payload.ComponentID) ||
		(payload.Kind != "endpoint" && payload.Kind != "dashboard") ||
		!slugPattern.MatchString(payload.Slug) || payload.Name == "" ||
		len([]byte(payload.Name)) > maximumNameBytes ||
		payload.Description == "" ||
		len([]byte(payload.Description)) > maximumDescriptionBytes ||
		!digestPattern.MatchString(payload.LegacyHash) ||
		sourceErr != nil ||
		err != nil || !bytes.Equal(canonical, payload.Definition) {
		return fmt.Errorf("analytics identity is invalid")
	}
	binding, bindingErr := json.Marshal(map[string]any{
		"slug": payload.Slug, "name": payload.Name,
		"description": payload.Description, "definition": payload.Definition,
	})
	bindingHash, bindingErr := canonicaljsonv1.DigestHex(binding, MaximumPayloadBytes)
	if bindingErr != nil || bindingHash != payload.LegacyHash {
		return fmt.Errorf("analytics legacy definition digest differs")
	}
	var definition map[string]json.RawMessage
	if json.Unmarshal(payload.Definition, &definition) != nil {
		return fmt.Errorf("analytics definition is invalid")
	}
	if payload.Kind == "endpoint" {
		return requireArrays(definition, "dataSources", "fields")
	}
	return requireArrays(definition, "dataSources", "widgets")
}

func requireArrays(definition map[string]json.RawMessage, keys ...string) error {
	for _, key := range keys {
		var items []json.RawMessage
		if json.Unmarshal(definition[key], &items) != nil || len(items) == 0 {
			return fmt.Errorf("analytics definition %s is empty", key)
		}
	}
	return nil
}

func clone(payload Payload) Payload {
	payload.Definition = append(json.RawMessage(nil), payload.Definition...)
	return payload
}

func (verified Verified) Value() Payload {
	if len(verified.document) == 0 {
		return Payload{}
	}
	return clone(verified.payload)
}

func (verified Verified) CanonicalJSON() []byte {
	return append([]byte(nil), verified.document...)
}
