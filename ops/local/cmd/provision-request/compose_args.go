package main

func composeArgs(extra ...string) []string {
	args := []string{
		"docker", "compose",
		"--env-file", ".env.local",
		"-f", "ops/local/compose.infra.yml",
		"-f", "ops/local/compose.tools.yml",
		"-f", "ops/local/compose.runtime.yml",
		"-f", "ops/local/compose.recovery.yml",
		"-f", "ops/local/compose.inventory.yml",
		"-f", "ops/local/compose.lifecycle.yml",
	}
	return append(args, extra...)
}

func showProvisionArgs() []string {
	return composeArgs(
		"run", "-T", "--rm", "--no-deps",
		"core-operator", "show-provision-request",
	)
}

func provisionArgs(tenantID string) []string {
	return composeArgs(
		"run", "--rm", "--no-deps",
		"-e", "HARPY_PROVISION_TENANT_ID="+tenantID,
		"tenant-provisioning",
	)
}
