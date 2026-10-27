package application

import "testing"

func TestAddTemplateRejectsEmptyTypeAndPersistsDraft(t *testing.T) {
	application, stores, scope, _ := newIntentFixture(t)
	if _, err := application.AddTemplate(t.Context(), scope,
		CreateTemplateInput{}); err != ErrInvalid {
		t.Fatalf("empty type error = %v", err)
	}
	draft, err := application.AddTemplate(t.Context(), scope,
		CreateTemplateInput{NotificationType: "document.archive.ready"})
	if err != nil || !draft.ID.IsV7() || stores.createdTemplate.ID != draft.ID {
		t.Fatalf("AddTemplate() = %+v, %v", draft, err)
	}
}
