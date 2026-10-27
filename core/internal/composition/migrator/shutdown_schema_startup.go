package migrator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func shutdownSchemaStartup(settings config.Runtime, runtime *telemetry.Runtime) {
	if runtime == nil {
		return
	}
	ctx, cancel := context.WithTimeout(
		context.Background(), settings.Telemetry().ExportTimeout(),
	)
	defer cancel()
	_ = runtime.Shutdown(ctx)
}
