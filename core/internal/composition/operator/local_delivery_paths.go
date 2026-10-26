package operator

import "path/filepath"

const (
	adapterID         = "local-terminal"
	claimTokenName    = ".delivery-claim-token"
	recoveryTokenName = ".delivery-recovery-token"
)

func (runtime *deliveryRuntime) passwordPath() string {
	return runtime.settings.InitialPassword().Path().Value()
}

func (runtime *deliveryRuntime) statePath(name string) string {
	return filepath.Join(filepath.Dir(runtime.passwordPath()), name)
}
