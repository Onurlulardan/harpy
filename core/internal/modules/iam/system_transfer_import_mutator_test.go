package iam

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestSystemTransferMutatorCannotBypassInvitationUserCreate(t *testing.T) {
	fieldID, err := publicresource.ParseID("a1000000-0000-4000-8000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	values, err := publicresource.NewFieldValues(map[publicresource.ID]json.RawMessage{
		fieldID: json.RawMessage(`"person@example.invalid"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	input, err := publicresource.NewCreateInput(
		publicresource.Users, values, []byte("not-used-by-import"), "import-row-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	mutator := &systemTransferResourceMutator{inner: &identitySystemResourcePort{}}
	if _, err = mutator.PrepareCreate(t.Context(), input); !errors.Is(
		err, publicresource.ErrForbidden,
	) {
		t.Fatalf("user create error=%v", err)
	}
}

func TestSystemTransferMutatorPortExposesOnlyCreateAndUpdate(t *testing.T) {
	typeOf := reflect.TypeOf((*systemtransfer.ResourceMutator)(nil)).Elem()
	wanted := map[string]bool{"PrepareCreate": true, "Create": true, "Update": true}
	if typeOf.NumMethod() != len(wanted) {
		t.Fatalf("method count=%d", typeOf.NumMethod())
	}
	for index := 0; index < typeOf.NumMethod(); index++ {
		if !wanted[typeOf.Method(index).Name] {
			t.Fatalf("unexpected bypass method=%s", typeOf.Method(index).Name)
		}
	}
}
