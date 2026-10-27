package platformauthority

import "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"

type Runtime struct {
	Binding   api.BindingBootstrap
	Initial   api.InitialAuthority
	Protected api.ProtectedAuthority
	Signer    api.SignerLifecycle
	AirGap    api.AirGapAuthority
}
