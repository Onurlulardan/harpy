package userrelation

import (
	"context"
	"errors"
	"testing"
)

func TestReplaceRejectsInvalidCommand(t *testing.T) {
	var commands *Commands
	if _, err := commands.Replace(
		context.Background(), TenantScope{}, ReplaceCommand{},
	); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("invalid replace error = %v", err)
	}
}

func TestReplaceConstructorRejectsMissingDependencies(t *testing.T) {
	if commands, err := NewCommands(nil, nil); commands != nil || !errors.Is(err, ErrStorage) {
		t.Fatalf("commands = %+v, error = %v", commands, err)
	}
}
