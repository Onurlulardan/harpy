package iam

func loginEnvironmentThrottleDimensions(
	tenantSlug, identifier, peerIP string,
) map[string][]string {
	return map[string][]string{
		"tenant_slug":   {tenantSlug},
		"identifier":    {identifier},
		"ip":            {peerIP},
		"identifier_ip": {identifier, peerIP},
	}
}

func loginTenantThrottleDimensions(
	identifier, peerIP string,
) map[string][]string {
	return map[string][]string{
		"identifier":    {identifier},
		"ip":            {peerIP},
		"identifier_ip": {identifier, peerIP},
	}
}

func refreshThrottleDimensions(selector, peerIP string) map[string][]string {
	return map[string][]string{
		"selector": {selector},
		"ip":       {peerIP},
	}
}
