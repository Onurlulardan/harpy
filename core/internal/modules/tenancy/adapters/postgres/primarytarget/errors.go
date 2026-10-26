package primarytarget

import "errors"

var errTargetUnavailable = errors.New(
	"tenancy primary tenant target storage is unavailable",
)
