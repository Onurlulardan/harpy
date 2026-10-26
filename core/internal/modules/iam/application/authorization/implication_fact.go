package authorization

import "github.com/google/uuid"

type ImplicationFact struct {
	ID                     uuid.UUID
	Kind                   string
	State                  string
	SourcePermissionID     uuid.UUID
	SourcePermissionState  string
	SourceResourceID       uuid.UUID
	SourceActionID         uuid.UUID
	SourceResourceState    string
	SourceActionState      string
	ImpliedPermissionID    uuid.UUID
	ImpliedPermissionState string
	ImpliedResourceID      uuid.UUID
	ImpliedActionID        uuid.UUID
	ImpliedResourceState   string
	ImpliedActionState     string
}

func (fact ImplicationFact) valid() bool {
	return fact.ID != uuid.Nil && validImplicationKind(fact.Kind) &&
		validCatalogState(fact.State) && fact.SourcePermissionID != uuid.Nil &&
		validCatalogState(fact.SourcePermissionState) &&
		fact.SourceResourceID != uuid.Nil && fact.SourceActionID != uuid.Nil &&
		validCatalogState(fact.SourceResourceState) &&
		validCatalogState(fact.SourceActionState) &&
		fact.ImpliedPermissionID != uuid.Nil &&
		validCatalogState(fact.ImpliedPermissionState) &&
		fact.ImpliedResourceID != uuid.Nil && fact.ImpliedActionID != uuid.Nil &&
		validCatalogState(fact.ImpliedResourceState) &&
		validCatalogState(fact.ImpliedActionState)
}

func validImplicationKind(kind string) bool {
	return kind == "v1_form_parent_v1" || kind == "forms_record_operation_v1"
}

func (fact ImplicationFact) activeFor(resourceID, actionID uuid.UUID) bool {
	return fact.valid() && fact.State == "active" &&
		fact.SourcePermissionState == "active" &&
		fact.SourceResourceState == "active" && fact.SourceActionState == "active" &&
		fact.ImpliedPermissionState == "active" &&
		fact.ImpliedResourceState == "active" && fact.ImpliedActionState == "active" &&
		fact.ImpliedResourceID == resourceID && fact.ImpliedActionID == actionID
}
