package dashboardexport

import (
	"errors"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

func mapDispatchError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{exportapp.ErrDispatchQuotaWait,
		exportapp.ErrDispatchCandidateChanged, exportapp.ErrDispatchLeaseStale,
		exportapp.ErrInvalid, exportapp.ErrForbidden} {
		if errors.Is(err, known) {
			return known
		}
	}
	var state sqlStateError
	if !errors.As(err, &state) {
		return exportapp.ErrUnavailable
	}
	switch state.SQLState() {
	case "P6046":
		return exportapp.ErrDispatchCandidateChanged
	case "P6047":
		return exportapp.ErrDispatchLeaseStale
	case "P6111":
		return exportapp.ErrDispatchQuotaWait
	case "42501":
		return exportapp.ErrForbidden
	case "22023", "22P02", "23514":
		return exportapp.ErrInvalid
	default:
		// P6048/P6049 and ambiguous database failures retain the exact spool.
		return exportapp.ErrUnavailable
	}
}
