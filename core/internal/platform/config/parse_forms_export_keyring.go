package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
)

type fcartKEKDocument struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

type fcartKEKKeyringDocument struct {
	Version  int                `json:"version"`
	Current  fcartKEKDocument   `json:"current"`
	Retained []fcartKEKDocument `json:"retained"`
}

func loadFCARTKEKKeyring(values environmentValues) (FCARTKEKKeyring, error) {
	return loadFCARTKEKKeyringAt(values, formsExportFCARTKeyringFileKey,
		"forms-export-fcart-kek-keyring.v1")
}

func loadFCARTKEKKeyringAt(values environmentValues, key, contract string) (
	FCARTKEKKeyring, error,
) {
	path, found := values.lookup(key)
	if !found {
		return FCARTKEKKeyring{}, configError(key, "is required")
	}
	document, err := readSecretFile(key, path, secretFileLimit)
	if err != nil {
		return FCARTKEKKeyring{}, err
	}
	return parseFCARTKEKKeyringFor([]byte(document), key, contract)
}

func parseFCARTKEKKeyring(raw []byte) (FCARTKEKKeyring, error) {
	return parseFCARTKEKKeyringFor(raw, formsExportFCARTKeyringFileKey,
		"forms-export-fcart-kek-keyring.v1")
}

func parseFCARTKEKKeyringFor(raw []byte, environmentKey, contract string) (
	FCARTKEKKeyring, error,
) {
	if !utf8.Valid(raw) || rejectDuplicateJSONMembers(raw) != nil {
		return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var document fcartKEKKeyringDocument
	if decoder.Decode(&document) != nil || document.Version != 1 || document.Retained == nil ||
		len(document.Retained) > 7 {
		return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
	}
	canonical, err := json.Marshal(document)
	if err != nil || !bytes.Equal(canonical, raw) {
		return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
	}
	current, ok := decodeFCARTKEK(document.Current)
	if !ok {
		return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
	}
	retained := make([]FCARTKEK, 0, len(document.Retained))
	seenIDs := map[string]struct{}{current.id: {}}
	seenMaterial := map[[32]byte]struct{}{current.material: {}}
	for _, item := range document.Retained {
		key, valid := decodeFCARTKEK(item)
		_, duplicateID := seenIDs[key.id]
		_, duplicateMaterial := seenMaterial[key.material]
		if !valid || duplicateID || duplicateMaterial {
			return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
		}
		seenIDs[key.id] = struct{}{}
		seenMaterial[key.material] = struct{}{}
		retained = append(retained, key)
	}
	keyring := FCARTKEKKeyring{current: current, retained: retained, valid: true}
	if keyring.Validate() != nil {
		return FCARTKEKKeyring{}, invalidFCARTKEKKeyringFor(environmentKey, contract)
	}
	return keyring, nil
}

func decodeFCARTKEK(document fcartKEKDocument) (FCARTKEK, bool) {
	if !validFCARTKeyID(document.KID) {
		return FCARTKEK{}, false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(document.Key)
	if err != nil || len(decoded) != 32 ||
		base64.RawURLEncoding.EncodeToString(decoded) != document.Key || allZero(decoded) {
		clear(decoded)
		return FCARTKEK{}, false
	}
	var material [32]byte
	copy(material[:], decoded)
	clear(decoded)
	return FCARTKEK{id: document.KID, material: material, valid: true}, true
}

func invalidFCARTKEKKeyring() error {
	return invalidFCARTKEKKeyringFor(formsExportFCARTKeyringFileKey,
		"forms-export-fcart-kek-keyring.v1")
}

func invalidFCARTKEKKeyringFor(key, contract string) error {
	return configError(key, "must satisfy "+contract)
}
