package provisionlaunch

import (
	"context"
	"fmt"
	"os/exec"
)

func Run(ctx context.Context, tenantID string) error {
	if tenantID == "" || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("tenant provisioning command is invalid")
	}
	command := exec.CommandContext(ctx, "docker", composeArgs(tenantID)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tenant provisioning container failed: %w", err)
	}
	_ = output
	return nil
}

func composeArgs(tenantID string) []string {
	return []string{
		"compose",
		"--env-file", ".env.local",
		"-f", "ops/local/compose.infra.yml",
		"-f", "ops/local/compose.tools.yml",
		"-f", "ops/local/compose.runtime.yml",
		"-f", "ops/local/compose.recovery.yml",
		"-f", "ops/local/compose.inventory.yml",
		"-f", "ops/local/compose.lifecycle.yml",
		"run", "--rm", "--no-deps",
		"-e", "HARPY_PROVISION_TENANT_ID=" + tenantID,
		"tenant-provisioning",
	}
}
