package operator

import (
	"context"
	"io"
)

func runLocalDelivery(ctx context.Context, action string, stdout io.Writer) error {
	runtime, err := openDeliveryRuntime(ctx)
	if err != nil {
		return err
	}
	defer runtime.close()
	switch action {
	case "prepare":
		return runtime.prepare(ctx)
	case "deliver":
		return runtime.deliver(ctx, stdout)
	case "verify":
		return runtime.verify(ctx)
	default:
		return installationv1Error{}
	}
}
