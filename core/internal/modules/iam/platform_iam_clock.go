package iam

import "time"

type iamPlatformClock struct{}

func (iamPlatformClock) Now() time.Time { return time.Now().UTC() }
