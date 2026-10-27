package postgres

import (
	"context"
	"testing"
)

func TestReadDevelopmentBindingInputRejectsInvalidEnvironment(t *testing.T) {
	_, err := (*Store)(nil).ReadDevelopmentBindingInput(context.Background(), "invalid")
	if err == nil {
		t.Fatal("invalid development binding environment was accepted")
	}
}
