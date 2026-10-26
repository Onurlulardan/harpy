package published_test

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func TestPublishedManifestBindsTypedTemplateWithoutChangingQueryShape(t *testing.T) {
	manifest, err := published.DecodeParameterManifest([]byte(`{
"contract":"restquery-parameters.v1","maximum_page_size":100,"parameters":[
{"name":"minimum_total","type":"integer","arity":"single","required":true,"minimum":0,"maximum":1000}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	literal, err := api.NewJSONLiteral([]byte(`"25"`))
	if err != nil {
		t.Fatal(err)
	}
	bound, err := published.BindPOST(
		manifest, map[string][]api.Literal{"minimum_total": {literal}}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := published.ResolveTemplate([]byte(`{
"select":["id","total"],
"where":{"and":[{"field":"total","op":"gte","value":{"$param":"minimum_total"}}]},
"sort":[],"page":{"size":50,"after":null},"count":"none","include":[]}`), bound, httpdecode.POST())
	if err != nil {
		var queryError api.QueryError
		if errors.As(err, &queryError) {
			t.Fatalf("template issues = %#v", queryError.Issues())
		}
		t.Fatal(err)
	}
	where, present := request.Where()
	children := where.Children()
	if !present || len(children) != 1 {
		t.Fatalf("resolved request = %#v", request)
	}
	predicate, predicatePresent := children[0].Predicate()
	if !predicatePresent || len(predicate.Values()) != 1 ||
		request.Page().Size() != 50 {
		t.Fatalf("resolved request = %#v", request)
	}
}

func TestPublishedManifestAndTemplateFailClosedOnContractDrift(t *testing.T) {
	if _, err := published.DecodeParameterManifest([]byte(
		`{"contract":"restquery-parameters.v1","maximum_page_size":50,"parameters":[],"extra":true}`,
	)); err == nil {
		t.Fatal("unknown manifest member was accepted")
	}
	manifest, err := published.DecodeParameterManifest([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	bound, err := published.BindPOST(manifest, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = published.ResolveTemplate(
		[]byte(`{"select":{"$param":"shape"}}`), bound, httpdecode.POST(),
	)
	if err == nil {
		t.Fatal("shape placeholder was accepted")
	}
}
