package platformlicense

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

type queryFixture struct {
	statement string
	arguments []any
	scan      rowFixture
}

func (fixture *queryFixture) access() DatabaseAccess {
	return DatabaseAccess{QueryRow: func(
		_ context.Context,
		statement string,
		arguments ...any,
	) Row {
		fixture.statement = statement
		fixture.arguments = arguments
		return fixture.scan
	}}
}

func fixtureStore(t *testing.T, query *queryFixture) *Store {
	t.Helper()
	store, err := NewStore(query.access())
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func testV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
