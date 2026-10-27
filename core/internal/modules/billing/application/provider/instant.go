package providerapp

import "time"

func canonicalInstant(value time.Time) (time.Time, error) {
	if value.IsZero() || value.Nanosecond()%int(time.Microsecond) != 0 {
		return time.Time{}, ErrWebhookRequest
	}
	return time.UnixMicro(value.UnixMicro()).UTC(), nil
}
