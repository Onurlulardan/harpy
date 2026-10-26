package published

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func ResolveTemplate(
	document []byte, bound BoundParameters, decoder api.POSTDecoder,
) (api.QueryRequest, error) {
	if decoder == nil || bound.Validate() != nil {
		return api.QueryRequest{}, fmt.Errorf("published query template is invalid")
	}
	values, err := templateValues(bound)
	if err != nil {
		return api.QueryRequest{}, err
	}
	jsonDecoder := json.NewDecoder(bytes.NewReader(document))
	jsonDecoder.UseNumber()
	var root any
	if err = jsonDecoder.Decode(&root); err != nil {
		return api.QueryRequest{}, fmt.Errorf("published query template is invalid")
	}
	root, err = resolveTemplateNode(root, "", values)
	if err != nil {
		return api.QueryRequest{}, err
	}
	if page, present, pageErr := bound.Page(); pageErr != nil {
		return api.QueryRequest{}, pageErr
	} else if present {
		root, err = applyTemplatePage(root, page)
		if err != nil {
			return api.QueryRequest{}, err
		}
	}
	resolvedDocument, err := json.Marshal(root)
	if err != nil {
		return api.QueryRequest{}, fmt.Errorf("published query template is invalid")
	}
	return decoder.Decode(resolvedDocument)
}

func templateValues(bound BoundParameters) (map[string][]resolved.Value, error) {
	names, err := bound.Names()
	if err != nil {
		return nil, err
	}
	result := make(map[string][]resolved.Value, len(names))
	for _, name := range names {
		values, found, lookupErr := bound.Lookup(name)
		if lookupErr != nil || !found || len(values) == 0 {
			return nil, fmt.Errorf("published query template is invalid")
		}
		result[name] = values
	}
	return result, nil
}

func resolveTemplateNode(
	node any, parent string, values map[string][]resolved.Value,
) (any, error) {
	switch typed := node.(type) {
	case []any:
		for index := range typed {
			resolvedNode, err := resolveTemplateNode(typed[index], parent, values)
			if err != nil {
				return nil, err
			}
			typed[index] = resolvedNode
		}
		return typed, nil
	case map[string]any:
		if parameter, found := typed["$param"]; found {
			name, valid := parameter.(string)
			bound, available := values[name]
			if !valid || !available || len(typed) != 1 ||
				(parent != "value" && parent != "values") {
				return nil, fmt.Errorf("published query template placeholder is invalid")
			}
			return resolvedTemplateValues(bound, parent == "values")
		}
		for key, value := range typed {
			resolvedNode, err := resolveTemplateNode(value, key, values)
			if err != nil {
				return nil, err
			}
			typed[key] = resolvedNode
		}
		return typed, nil
	default:
		return node, nil
	}
}
