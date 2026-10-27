package deploymentpolicy

import _ "embed"

//go:embed sql/platform/get.sql
var getDeploymentPolicySQL string

//go:embed sql/platform/initialize.sql
var initializeDeploymentPolicySQL string
