package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/diagnostics"
)

type DiagnosticRunner interface {
	RunDiagnostic(context.Context, diagnostics.Definition) ([]diagnostics.Field, error)
}

type Clock interface {
	Now() time.Time
}
