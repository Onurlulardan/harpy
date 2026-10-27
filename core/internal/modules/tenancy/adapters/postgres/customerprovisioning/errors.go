package customerprovisioning

import (
	"errors"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

type sqlStateError interface{ SQLState() string }

func classify(err error, notFound RowNotFound) error {
	if err == nil {
		return nil
	}
	if notFound != nil && notFound(err) {
		return provisioning.ErrNotFound
	}
	var state sqlStateError
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "23505":
			return provisioning.ErrConflict
		case "02000":
			return provisioning.ErrNotFound
		case "42501":
			return provisioning.ErrForbidden
		}
	}
	return provisioning.ErrUnavailable
}
