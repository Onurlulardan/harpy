package httpquery

import (
	"testing"
)

func TestPublishedGETAndPOSTNormalizeToSameTypedBindings(t *testing.T) {
	manifest := publishedTestManifest(t)
	getBound, err := NewPublishedGETDecoder().Decode(
		"param[count]=7&param[flags][0]=true&param[flags][1]=false&"+
			"param[flags][2]=true&param[payload]=%7B%22a%22%3A1%7D&"+
			"page[size]=10&page[after]=opaque",
		manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	postBound, err := NewPublishedPOSTDecoder().Decode([]byte(
		`{"params":{"count":"7","flags":[true,false,true],`+
			`"payload":{"a":1}},"page":{"size":10,"after":"opaque"}}`,
	), manifest)
	if err != nil {
		t.Fatal(err)
	}
	requirePublishedEqual(t, getBound, postBound)
	requirePublishedScalar(t, getBound, "count", "7")
	requirePublishedScalar(t, getBound, "label", "fallback")
	flags, found, err := getBound.Lookup("flags")
	if err != nil || !found || len(flags) != 2 {
		t.Fatalf("flags = %#v, found = %v, err = %v", flags, found, err)
	}
	payload, found, err := getBound.Lookup("payload")
	if err != nil || !found || len(payload) != 1 {
		t.Fatalf("payload = %#v, found = %v, err = %v", payload, found, err)
	}
	document, isJSON, err := payload[0].JSON()
	if err != nil || !isJSON || string(document) != `{"a":1}` {
		t.Fatalf("payload = %s, JSON = %v, err = %v", document, isJSON, err)
	}
}

func TestPublishedRawGETOrderHasNoSemanticMeaning(t *testing.T) {
	manifest := publishedTestManifest(t)
	left, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&param[flags][0]=true&param[flags][1]=false", manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	right, err := NewPublishedGETDecoder().Decode(
		"param[flags][1]=false&param[count]=1&param[flags][0]=true", manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	requirePublishedEqual(t, left, right)
}

func TestPublishedTextValuesNormalizeNFCInBothTransports(t *testing.T) {
	manifest := publishedTestManifest(t)
	getBound, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&param[label]=e%CC%81", manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	postBound, err := NewPublishedPOSTDecoder().Decode(
		[]byte(`{"params":{"count":"1","label":"e\u0301"}}`), manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	requirePublishedEqual(t, getBound, postBound)
	requirePublishedScalar(t, getBound, "label", "\u00e9")
}

func TestPublishedPOSTTreatsJSONArrayAsSingleJSONParameter(t *testing.T) {
	manifest := publishedTestManifest(t)
	bound, err := NewPublishedPOSTDecoder().Decode(
		[]byte(`{"params":{"count":"1","payload":[1,2]}}`), manifest,
	)
	if err != nil {
		t.Fatal(err)
	}
	values, found, err := bound.Lookup("payload")
	if err != nil || !found || len(values) != 1 {
		t.Fatalf("payload = %#v, found = %v, err = %v", values, found, err)
	}
	document, isJSON, _ := values[0].JSON()
	if !isJSON || string(document) != `[1,2]` {
		t.Fatalf("payload = %s", document)
	}
}
