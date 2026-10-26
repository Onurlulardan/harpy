package billingv1

import (
	"context"
	"fmt"
)

type DeploymentMode string

const (
	DeploymentSaaS   DeploymentMode = "saas"
	DeploymentOnPrem DeploymentMode = "onpremise"
)

type AvailabilityPolicy interface {
	AdmitRoute(context.Context, Route) error
}

type deploymentAvailability struct{ mode DeploymentMode }

func NewAvailabilityPolicy(mode DeploymentMode) (AvailabilityPolicy, error) {
	if mode != DeploymentSaaS && mode != DeploymentOnPrem {
		return nil, fmt.Errorf("Billing V1 deployment mode is invalid")
	}
	return deploymentAvailability{mode: mode}, nil
}

func (policy deploymentAvailability) AdmitRoute(ctx context.Context, route Route) error {
	if ctx == nil || ctx.Err() != nil || route.ID == "" {
		return ErrAvailabilityEvidence
	}
	if policy.mode == DeploymentSaaS || onPremBillingRead(route.ID) {
		return nil
	}
	return ErrRouteUnavailable
}

func onPremBillingRead(routeID string) bool {
	return routeID == "billing.route.subscription-read" ||
		routeID == "billing.route.usage-read"
}
