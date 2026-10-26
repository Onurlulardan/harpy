package postgres

import (
	"fmt"
	"testing"
)

type sqlStateFixture string

func (value sqlStateFixture) Error() string    { return string(value) }
func (value sqlStateFixture) SQLState() string { return string(value) }

func TestErrorHasSQLStateRecognizesWrappedCarrier(t *testing.T) {
	err := fmt.Errorf("reconcile failed: %w", sqlStateFixture("42501"))
	if !ErrorHasSQLState(err, "23505", "42501") {
		t.Fatal("expected wrapped SQL state to match")
	}
	if ErrorHasSQLState(err, "08006") {
		t.Fatal("unexpected SQL state match")
	}
}
