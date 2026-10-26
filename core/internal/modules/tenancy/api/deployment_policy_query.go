package api

import "context"

type DeploymentPolicyQuery interface {
	GetDeploymentPolicy(context.Context) (DeploymentPolicy, error)
}
