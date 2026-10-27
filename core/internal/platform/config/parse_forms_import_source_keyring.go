package config

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

func loadFormsImportSourceKEKKeyring(values environmentValues) (
	FormsImportSourceKEKKeyring, error,
) {
	path, found := values.lookup(formsImportSourceFCARTKeyringFileKey)
	if !found {
		return FormsImportSourceKEKKeyring{}, configError(
			formsImportSourceFCARTKeyringFileKey, "is required")
	}
	document, err := readSecretFile(formsImportSourceFCARTKeyringFileKey,
		path, secretFileLimit)
	if err != nil {
		return FormsImportSourceKEKKeyring{}, err
	}
	return parseFormsImportSourceKEKKeyring([]byte(document))
}

func parseFormsImportSourceKEKKeyring(raw []byte) (FormsImportSourceKEKKeyring, error) {
	if !utf8.Valid(raw) || rejectDuplicateJSONMembers(raw) != nil {
		return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var document fcartKEKKeyringDocument
	if decoder.Decode(&document) != nil || document.Version != 1 ||
		document.Retained == nil || len(document.Retained) > 7 {
		return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
	}
	canonical, err := json.Marshal(document)
	if err != nil || !bytes.Equal(canonical, raw) {
		return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
	}
	current, valid := decodeFormsImportSourceKEK(document.Current)
	if !valid {
		return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
	}
	retained := make([]FormsImportSourceKEK, 0, len(document.Retained))
	for _, item := range document.Retained {
		key, valid := decodeFormsImportSourceKEK(item)
		if !valid {
			return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
		}
		retained = append(retained, key)
	}
	keyring := FormsImportSourceKEKKeyring{current: current,
		retained: retained, valid: true}
	if keyring.Validate() != nil {
		return FormsImportSourceKEKKeyring{}, invalidFormsImportSourceKeyring()
	}
	return keyring, nil
}

func decodeFormsImportSourceKEK(document fcartKEKDocument) (
	FormsImportSourceKEK, bool,
) {
	decoded, valid := decodeFCARTKEK(document)
	if !valid {
		return FormsImportSourceKEK{}, false
	}
	return FormsImportSourceKEK{id: decoded.id, material: decoded.material, valid: true}, true
}

func invalidFormsImportSourceKeyring() error {
	return configError(formsImportSourceFCARTKeyringFileKey,
		"must satisfy forms-import-source-fcart-kek-keyring.v1")
}
