package csrf

import "time"

const (
	RefreshCookieName = "__Secure-harpy-refresh"
	RefreshCookiePath = "/api/auth"
	CSRFCookieName    = "__Host-harpy-csrf"
	CSRFCookiePath    = "/"
	HeaderName        = "X-CSRF-Token"
	Lifetime          = 15 * time.Minute
	ClockSkew         = 60 * time.Second
	MaximumWireSize   = 192
	MaximumOriginSize = 512
	keyBytes          = 32
	nonceBytes        = 32
)
