package config

import "fmt"

const deploymentModeEnvironmentKey = "DEPLOYMENT_MODE"

type DeploymentMode string

const (
	DeploymentSaaS   DeploymentMode = "saas"
	DeploymentOnPrem DeploymentMode = "onpremise"
)

func loadDeploymentMode(values environmentValues) (DeploymentMode, error) {
	raw, found := values.lookup(deploymentModeEnvironmentKey)
	if !found {
		return DeploymentOnPrem, nil
	}
	mode := DeploymentMode(raw)
	if mode != DeploymentSaaS && mode != DeploymentOnPrem {
		return "", configError(deploymentModeEnvironmentKey, "must be saas or onpremise")
	}
	return mode, nil
}

func (mode DeploymentMode) Validate() error {
	if mode != DeploymentSaaS && mode != DeploymentOnPrem {
		return fmt.Errorf("deployment mode is invalid")
	}
	return nil
}
