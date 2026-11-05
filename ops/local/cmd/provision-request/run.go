package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type runner interface {
	Output(context.Context, []string) ([]byte, error)
	Run(context.Context, []string) error
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
