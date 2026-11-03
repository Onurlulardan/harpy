package initialize

import "context"

func UnavailableGoose(context.Context) error {
	return errGooseNotInstalled
}

func PermitApplication(context.Context) error { return nil }
