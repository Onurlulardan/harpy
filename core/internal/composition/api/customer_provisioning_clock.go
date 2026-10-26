package api

import "time"

type customerProvisioningClock struct{}

func (customerProvisioningClock) Now() time.Time { return time.Now().UTC() }
