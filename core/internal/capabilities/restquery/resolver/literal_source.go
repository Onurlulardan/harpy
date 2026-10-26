package resolver

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type literalSource struct {
	json   []byte
	text   string
	isJSON bool
}

func readLiteralSource(literal api.Literal) (literalSource, restqueryv1.IssueCode) {
	jsonValue, hasJSON, jsonErr := literal.JSONValue()
	textValue, hasText, textErr := literal.TextValue()
	if jsonErr != nil || textErr != nil || hasJSON == hasText {
		return literalSource{}, restqueryv1.IssueValueNotCanonical
	}
	return literalSource{json: jsonValue, text: textValue, isJSON: hasJSON}, ""
}

func (source literalSource) stringValue() (string, restqueryv1.IssueCode) {
	if !source.isJSON {
		return source.text, ""
	}
	var value string
	if json.Unmarshal(source.json, &value) != nil {
		return "", restqueryv1.IssueTypeMismatch
	}
	return value, ""
}

func (source literalSource) booleanValue() (bool, restqueryv1.IssueCode) {
	if !source.isJSON {
		if source.text == "true" {
			return true, ""
		}
		if source.text == "false" {
			return false, ""
		}
		return false, restqueryv1.IssueValueNotCanonical
	}
	var value bool
	if json.Unmarshal(source.json, &value) != nil {
		return false, restqueryv1.IssueTypeMismatch
	}
	return value, ""
}

func (source literalSource) jsonDocument() []byte {
	if source.isJSON {
		return append([]byte(nil), source.json...)
	}
	return []byte(source.text)
}
