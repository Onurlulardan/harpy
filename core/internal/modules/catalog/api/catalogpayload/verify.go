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

const MaximumPayloadBytes = 4 << 20

var (
	uuidV7Pattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
	stableIDPattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	)
	slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
)

type DemoForm struct {
	Slug               string            `json:"slug"`
	LegacyFormConfigID string            `json:"legacy_form_config_id"`
	Rows               []json.RawMessage `json:"rows"`
}

type Payload struct {
	ContractVersion   string     `json:"contract_version"`
	ComponentID       string     `json:"component_id"`
	ModuleSlug        string     `json:"module_slug"`
	SourceContract    string     `json:"source_contract"`
	ExcludedFormSlugs []string   `json:"excluded_form_slugs"`
	DemoForms         []DemoForm `json:"demo_forms"`
}

type Verified struct {
	payload  Payload
	document []byte
}

func Verify(document []byte, componentID, expectedSHA256 string) (Verified, error) {
	if err := catalogpackagev1.ValidateCanonicalPayload(
		document, expectedSHA256, MaximumPayloadBytes,
	); err != nil {
		return Verified{}, fmt.Errorf("catalog demo payload is invalid")
	}
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		validate(payload, componentID) != nil {
		return Verified{}, fmt.Errorf("catalog demo payload contract is invalid")
	}
	return Verified{payload: clone(payload), document: append([]byte(nil), document...)}, nil
}

func validate(payload Payload, componentID string) error {
	expectedSource, sourceErr := catalogpackagev1.FirstPartySourceContract(
		payload.ModuleSlug,
	)
	if payload.ContractVersion != catalogpackagev1.ContractCatalog ||
		payload.ComponentID != componentID ||
		!uuidV7Pattern.MatchString(payload.ComponentID) ||
		sourceErr != nil || payload.SourceContract != expectedSource ||
		len(payload.DemoForms) == 0 {
		return fmt.Errorf("catalog demo identity is invalid")
	}
	seen := map[string]bool{}
	for _, form := range payload.DemoForms {
		if !slugPattern.MatchString(form.Slug) || seen[form.Slug] ||
			!validStableID(form.LegacyFormConfigID) ||
			len(form.Rows) == 0 {
			return fmt.Errorf("catalog demo form is invalid")
		}
		seen[form.Slug] = true
		for _, row := range form.Rows {
			if err := validateRow(row); err != nil {
				return err
			}
		}
	}
	excluded := map[string]bool{}
	for _, slug := range payload.ExcludedFormSlugs {
		if !slugPattern.MatchString(slug) ||
			excluded[slug] || seen[slug] {
			return fmt.Errorf("catalog demo exclusion is invalid")
		}
		excluded[slug] = true
	}
	return nil
}

func validateRow(row json.RawMessage) error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(row, MaximumPayloadBytes)
	if err != nil || !bytes.Equal(canonical, row) {
		return fmt.Errorf("catalog demo row is not canonical")
	}
	var identity struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(row, &identity) != nil || !validStableID(identity.ID) {
		return fmt.Errorf("catalog demo row identity is invalid")
	}
	return nil
}

func validStableID(value string) bool {
	return value != "00000000-0000-0000-0000-000000000000" &&
		stableIDPattern.MatchString(value)
}

func clone(payload Payload) Payload {
	payload.ExcludedFormSlugs = append([]string(nil), payload.ExcludedFormSlugs...)
	payload.DemoForms = append([]DemoForm(nil), payload.DemoForms...)
	for formIndex := range payload.DemoForms {
		rows := make([]json.RawMessage, len(payload.DemoForms[formIndex].Rows))
		for rowIndex := range rows {
			rows[rowIndex] = append(
				json.RawMessage(nil), payload.DemoForms[formIndex].Rows[rowIndex]...,
			)
		}
		payload.DemoForms[formIndex].Rows = rows
	}
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
