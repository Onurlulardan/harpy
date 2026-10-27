package cursor

import (
	"bytes"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestUUIDEnumRelationAndNullLastValuesRoundTrip(t *testing.T) {
	fixture := baseContextFixture(t)
	fixture.sorts = []SortDefinition{
		mustSort(t, testFieldID, api.DirectionAscending, api.LogicalTypeUUID, 36),
		mustSort(t, "66666666-6666-4666-8666-666666666666", api.DirectionAscending, api.LogicalTypeEnum, 16),
		mustSort(t, "77777777-7777-4777-8777-777777777777", api.DirectionDescending, api.LogicalTypeRelation, 36),
		mustSort(t, "88888888-8888-4888-8888-888888888888", api.DirectionAscending, api.LogicalTypeText, 1),
	}
	context := buildContext(t, fixture)
	values := []LastValue{
		scalarLast(t, api.LogicalTypeUUID, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		scalarLast(t, api.LogicalTypeEnum, "active"),
		scalarLast(t, api.LogicalTypeRelation, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
		NewNullLastValue(),
	}
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{12}, nonceBytes)),
		func() time.Time { return testNow },
	)
	token, err := codec.Encode(context, values)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := codec.Decode(token, context)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _ := claims.LastValues()
	for index := 0; index < 3; index++ {
		expected, _, _ := values[index].value.Scalar()
		actual, _, _ := decoded[index].value.Scalar()
		if actual != expected {
			t.Fatalf("value %d = %q, want %q", index, actual, expected)
		}
	}
	if null, _ := decoded[3].IsNull(); !null {
		t.Fatal("null typed value was not preserved")
	}
}
