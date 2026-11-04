package postgres

import (
	"reflect"
	"testing"
)

func TestOpenDoesNotTakeSchemaContract(t *testing.T) {
	t.Parallel()
	openType := reflect.TypeOf(Open)
	if openType.NumIn() != 3 {
		t.Fatalf("Open parameters = %d, type %v", openType.NumIn(), openType)
	}
}
