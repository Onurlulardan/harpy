package runtimev1

import "testing"

func mergedConfigCatalog(t *testing.T) any {
	t.Helper()
	base, ok := readJSONValue(t, "config.json").(map[string]any)
	if !ok {
		t.Fatal("config catalog is not an object")
	}
	tail, ok := readJSONValue(t, "config.tail.json").(map[string]any)
	if !ok {
		t.Fatal("config tail is not an object")
	}
	baseVariables, ok := base["variables"].([]any)
	tailVariables, tailOK := tail["variables"].([]any)
	if !ok || !tailOK {
		t.Fatal("config catalog variables are not arrays")
	}
	base["variables"] = append(baseVariables, tailVariables...)
	return base
}
