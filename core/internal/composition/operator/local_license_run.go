package operator

import "context"

func runLocalLicense(ctx context.Context) error {
	return reconcileLocalDevelopmentLicense(ctx)
}
