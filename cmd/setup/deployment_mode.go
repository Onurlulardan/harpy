package main

import "errors"

func deploymentMode(profile string) (string, error) {
	switch profile {
	case "single-tenant":
		return "onpremise", nil
	case "multi-tenant":
		return "saas", nil
	default:
		return "", errors.New("profile must be single-tenant or multi-tenant")
	}
}

func deploymentNote(profile string) string {
	switch profile {
	case "single-tenant":
		return "DEPLOYMENT_MODE=onpremise (single-tenant, customer tenant cap 0)"
	case "multi-tenant":
		return "DEPLOYMENT_MODE=saas (multi-tenant, unbounded customer tenants)"
	default:
		return ""
	}
}
