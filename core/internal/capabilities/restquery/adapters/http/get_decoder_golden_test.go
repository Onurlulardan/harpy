package httpquery

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestGETDecoderCanonicalRecordGolden(t *testing.T) {
	rawQuery := "sort[0][direction]=desc&filter[1][value]=10&select[1]=status&" +
		"filter[0][values][1]=DONE&page[after]=cursor&filter[0][op]=in&" +
		"search[term]=e%CC%81&filter[1][field]=total&select[0]=id&" +
		"filter[0][field]=status&sort[0][field]=created_at&count=exact&" +
		"filter[1][op]=gte&filter[0][values][0]=OPEN&page[size]=25"
	request := decodeGETSuccess(t, rawQuery)
	selects := request.Select()
	if len(selects) != 2 || selects[0] != "id" || selects[1] != "status" {
		t.Fatalf("select order changed: %#v", selects)
	}
	where, present := request.Where()
	if !present || where.Kind() != api.RequestFilterAnd || len(where.Children()) != 2 {
		t.Fatal("GET filters did not normalize to flat AND")
	}
	first, _ := where.Children()[0].Predicate()
	if first.Field() != "status" || first.Operator() != api.OperatorIn {
		t.Fatal("first predicate changed")
	}
	firstValues := first.Values()
	if len(firstValues) != 2 || getTextValue(t, firstValues[0]) != "OPEN" ||
		getTextValue(t, firstValues[1]) != "DONE" {
		t.Fatal("indexed operand order changed")
	}
	second, _ := where.Children()[1].Predicate()
	if second.Field() != "total" || second.Operator() != api.OperatorGTE ||
		getTextValue(t, second.Values()[0]) != "10" {
		t.Fatal("second predicate changed")
	}
	sorts := request.Sort()
	if len(sorts) != 1 || sorts[0].Field() != "created_at" ||
		sorts[0].Direction() != api.DirectionDescending {
		t.Fatal("sort changed")
	}
	search, present := request.Search()
	if !present || search.Term() != "é" {
		t.Fatalf("search was not NFC-normalized: %q", search.Term())
	}
	if request.Page().Size() != 25 || request.Count() != api.CountModeExact {
		t.Fatal("page or count changed")
	}
	if after, present := request.Page().After(); !present || after != "cursor" {
		t.Fatal("cursor changed")
	}
}

func TestGETDecoderKeepsCommaAndDecodesPercentOnce(t *testing.T) {
	rawQuery := "filter[0][field]=x&filter[0][op]=eq&" +
		"filter[0][value]=a%2Cb%2520"
	request := decodeGETSuccess(t, rawQuery)
	where, _ := request.Where()
	predicate, _ := where.Children()[0].Predicate()
	if value := getTextValue(t, predicate.Values()[0]); value != "a,b%20" {
		t.Fatalf("literal was split or decoded twice: %q", value)
	}
}

func TestGETDecoderRawParameterOrderIsNotSemantic(t *testing.T) {
	left := decodeGETSuccess(t,
		"select[1]=b&filter[0][op]=eq&select[0]=a&"+
			"filter[0][value]=x&filter[0][field]=f")
	right := decodeGETSuccess(t,
		"filter[0][field]=f&select[0]=a&filter[0][value]=x&"+
			"select[1]=b&filter[0][op]=eq")
	if left.Select()[0] != right.Select()[0] || left.Select()[1] != right.Select()[1] {
		t.Fatal("raw parameter order changed select semantics")
	}
	leftWhere, _ := left.Where()
	rightWhere, _ := right.Where()
	leftPredicate, _ := leftWhere.Children()[0].Predicate()
	rightPredicate, _ := rightWhere.Children()[0].Predicate()
	if leftPredicate.Field() != rightPredicate.Field() ||
		getTextValue(t, leftPredicate.Values()[0]) !=
			getTextValue(t, rightPredicate.Values()[0]) {
		t.Fatal("raw parameter order changed filter semantics")
	}
}
