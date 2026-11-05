package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type runner interface {
	Output(context.Context, []string) ([]byte, error)
	Run(context.Context, []string) error
	Combined(context.Context, []string) ([]byte, int, error)
}

type processRunner struct{}

func (processRunner) Output(ctx context.Context, argv []string) ([]byte, error) {
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Stderr = os.Stderr
	return command.Output()
}

func (processRunner) Run(ctx context.Context, argv []string) error {
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func (processRunner) Combined(ctx context.Context, argv []string) ([]byte, int, error) {
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	output, err := command.CombinedOutput()
	if err == nil {
		return output, 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, exitErr.ExitCode(), err
	}
	return output, 1, err
}

func waitOperator(ctx context.Context, commands runner) error {
	if ctx == nil || ctx.Err() != nil || commands == nil {
		return errInvalidTenant
	}
	output, code, err := commands.Combined(ctx, operatorInitializeArgs())
	_, _ = os.Stderr.Write(output)
	if code == 0 && err == nil {
		return nil
	}
	if code == 5 && bytes.Contains(output, []byte("waiting for host execution")) {
		return provisionTenant(ctx, commands)
	}
	if err != nil {
		return fmt.Errorf("core-operator exited %d: %w", code, err)
	}
	return fmt.Errorf("core-operator exited %d", code)
}

func provisionTenant(ctx context.Context, commands runner) error {
	if ctx == nil || ctx.Err() != nil || commands == nil {
		return errInvalidTenant
	}
	output, err := commands.Output(ctx, showProvisionArgs())
	if err != nil {
		return fmt.Errorf("read tenant provisioning request: %w", err)
	}
	tenantID, err := parseTenantID(output)
	if err != nil {
		return err
	}
	if err = commands.Run(ctx, provisionArgs(tenantID)); err != nil {
		return fmt.Errorf("run tenant provisioning: %w", err)
	}
	return nil
}
