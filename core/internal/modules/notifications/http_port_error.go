package notifications

import (
	"errors"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
)

func httpPortError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, application.ErrInvalid), errors.Is(err, api.ErrIntentInvalid):
		return notificationhttp.ErrInvalid
	case errors.Is(err, application.ErrForbidden), errors.Is(err, api.ErrForbidden):
		return notificationhttp.ErrForbidden
	case errors.Is(err, application.ErrNotFound), errors.Is(err, api.ErrRecipientNotFound),
		errors.Is(err, api.ErrTemplateNotFound):
		return notificationhttp.ErrNotFound
	case errors.Is(err, api.ErrEffectKeyReused):
		return api.ErrEffectKeyReused
	case errors.Is(err, application.ErrConflict):
		return notificationhttp.ErrConflict
	default:
		return notificationhttp.ErrUnavailable
	}
}
