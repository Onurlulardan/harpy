package httpquery

import (
	"reflect"
	"sort"
	"testing"
)

func TestScanJSONMembersFindsDuplicatesAtEveryDepth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"root", `{"field":1,"field":2}`, []string{"/field"}},
		{"nested", `{"where":{"op":"eq","op":"ne"}}`, []string{"/where/op"}},
		{
			"array objects",
			`{"items":[{"id":1,"id":2},{"id":3,"id":4}]}`,
			[]string{"/items/0/id", "/items/1/id"},
		},
		{
			"literal object",
			`{"where":{"value":{"secret":1,"secret":2}}}`,
			[]string{"/where/value/secret"},
		},
		{"escaped pointer", `{"a/b~c":1,"a/b~c":2}`, []string{"/a~1b~0c"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			scan, err := scanJSONMembers([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}
			actual := scan.duplicatePaths()
			sort.Strings(actual)
			if !reflect.DeepEqual(actual, test.want) {
				t.Fatalf("duplicate paths = %v, want %v", actual, test.want)
			}
		})
	}
}

func TestScanJSONMembersAcceptsValidDocument(t *testing.T) {
	t.Parallel()
	document := `{"select":["id"],"where":{"field":"status","value":{"x":1}},"page":null}`
	scan, err := scanJSONMembers([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	if paths := scan.duplicatePaths(); len(paths) != 0 {
		t.Fatalf("valid document has duplicate members: %v", paths)
	}
}
