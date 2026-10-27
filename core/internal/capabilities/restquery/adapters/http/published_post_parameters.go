package httpquery

import (
	"bytes"
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func decodePublishedPOSTParameters(
	root publishedPOSTRoot,
	manifest published.ParameterManifest,
	collector *publishedIssueCollector,
) map[string][]api.Literal {
	result := make(map[string][]api.Literal)
	raw, present := root["params"]
	if !present {
		return result
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		collector.add("/params", restqueryv1.IssueTypeMismatch)
		return result
	}
	var parameters map[string]json.RawMessage
	if json.Unmarshal(trimmed, &parameters) != nil || parameters == nil {
		collector.add("/params", restqueryv1.IssueTypeMismatch)
		return result
	}
	for name, value := range parameters {
		path := appendJSONPointer("/params", name)
		definition, known, err := manifest.Lookup(name)
		if err != nil || !known {
			collector.add(path, restqueryv1.IssueParameterUnknown)
			continue
		}
		arity, arityErr := definition.Arity()
		if arityErr != nil {
			collector.add(path, restqueryv1.IssueParameterArityMismatch)
			continue
		}
		if arity == published.ParameterAritySingle {
			literal, valid := decodePublishedPOSTLiteral(value, path, collector)
			if valid {
				result[name] = []api.Literal{literal}
			}
			continue
		}
		result[name] = decodePublishedPOSTList(value, path, collector)
	}
	return result
}

func decodePublishedPOSTList(
	raw json.RawMessage, path string, collector *publishedIssueCollector,
) []api.Literal {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		collector.add(path, restqueryv1.IssueParameterArityMismatch)
		return nil
	}
	var items []json.RawMessage
	if json.Unmarshal(trimmed, &items) != nil || items == nil ||
		len(items) < 1 || len(items) > 100 {
		collector.add(path, restqueryv1.IssueParameterArityMismatch)
		return nil
	}
	result := make([]api.Literal, 0, len(items))
	valid := true
	for index, item := range items {
		literal, itemValid := decodePublishedPOSTLiteral(
			item, path+"/"+strconv.Itoa(index), collector,
		)
		if !itemValid {
			valid = false
			continue
		}
		result = append(result, literal)
	}
	if !valid {
		return nil
	}
	return result
}

func decodePublishedPOSTLiteral(
	raw json.RawMessage, path string, collector *publishedIssueCollector,
) (api.Literal, bool) {
	literal, err := api.NewJSONLiteral(raw)
	if err == nil {
		return literal, true
	}
	code := restqueryv1.IssueValueNotCanonical
	if rawIsNull(raw) {
		code = restqueryv1.IssueValueNotAllowed
	}
	collector.add(path, code)
	return api.Literal{}, false
}
