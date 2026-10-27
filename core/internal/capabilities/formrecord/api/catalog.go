package recordkernel

import (
	"github.com/google/uuid"
	"slices"
)

type Catalog struct {
	tenantID              uuid.UUID
	publicationID         uuid.UUID
	publicationRevision   int64
	formID                uuid.UUID
	formVersionID         uuid.UUID
	definitionContract    string
	definitionSHA256      [32]byte
	fieldCatalogSHA256    [32]byte
	budgetProfileID       string
	budgetSHA256          [32]byte
	registrySHA256        [32]byte
	mappingCatalogSHA256  [32]byte
	recordRulesSHA256     [32]byte
	recordAttestationSHA  [32]byte
	limits                BudgetLimits
	fields                []Field
	fieldsByID            map[uuid.UUID]Field
	fieldsByName          map[string]Field
	workflowActions       map[Action]struct{}
	defaults              []defaultValue
	calculations          []calculation
	crossFieldValidations []crossFieldRule
	canonical             []byte
	sha256                [32]byte
}

func (catalog Catalog) TenantID() uuid.UUID      { return catalog.tenantID }
func (catalog Catalog) PublicationID() uuid.UUID { return catalog.publicationID }
func (catalog Catalog) PublicationRevision() int64 {
	return catalog.publicationRevision
}
func (catalog Catalog) FormID() uuid.UUID        { return catalog.formID }
func (catalog Catalog) FormVersionID() uuid.UUID { return catalog.formVersionID }
func (catalog Catalog) SHA256() [32]byte         { return catalog.sha256 }

func (catalog Catalog) CanonicalJSON() []byte {
	return append([]byte(nil), catalog.canonical...)
}

func (catalog Catalog) Fields() []Field {
	result := make([]Field, len(catalog.fields))
	for index, field := range catalog.fields {
		result[index] = cloneField(field)
	}
	return result
}

func (catalog Catalog) FieldByID(fieldID uuid.UUID) (Field, bool) {
	field, found := catalog.fieldsByID[fieldID]
	return cloneField(field), found
}

func (catalog Catalog) FieldByName(name string) (Field, bool) {
	field, found := catalog.fieldsByName[name]
	return cloneField(field), found
}

func (catalog Catalog) AllowsWorkflow(action Action) bool {
	_, allowed := catalog.workflowActions[action]
	return allowed
}

func (catalog Catalog) WorkflowActions() []Action {
	result := make([]Action, 0, len(catalog.workflowActions))
	for action := range catalog.workflowActions {
		result = append(result, action)
	}
	slices.Sort(result)
	return result
}
