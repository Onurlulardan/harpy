package authorization

import "github.com/google/uuid"

type ActorFact struct {
	ID                   uuid.UUID
	Status               string
	TokenVersion         int64
	AuthorizationVersion int64
}

func (fact ActorFact) valid() bool {
	return fact.ID != uuid.Nil && validActorStatus(fact.Status) &&
		fact.TokenVersion > 0 && fact.AuthorizationVersion >= 0
}

type ResourceFact struct {
	ID        uuid.UUID
	State     string
	Protected bool
}

func (fact ResourceFact) valid() bool {
	return fact.ID != uuid.Nil && validCatalogState(fact.State)
}

type ActionFact struct {
	ID        uuid.UUID
	State     string
	Protected bool
}

func (fact ActionFact) valid() bool {
	return fact.ID != uuid.Nil && validCatalogState(fact.State)
}

func validActorStatus(value string) bool {
	switch value {
	case "pending", "active", "suspended", "disabled", "tombstoned":
		return true
	default:
		return false
	}
}

func validCatalogState(value string) bool {
	return value == "active" || value == "retired"
}
