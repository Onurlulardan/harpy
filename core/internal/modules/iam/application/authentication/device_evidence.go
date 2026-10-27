package authentication

import "regexp"

var deviceKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

type DeviceEvidence struct {
	Kind          string
	BrowserFamily string
	OSFamily      string
	IPKeyID       string
	IPHMAC        [32]byte
}

func (evidence DeviceEvidence) validate() error {
	if evidence.Kind != "desktop" && evidence.Kind != "mobile" &&
		evidence.Kind != "tablet" && evidence.Kind != "unknown" {
		return ErrAuthenticationContract
	}
	if !deviceKeyPattern.MatchString(evidence.BrowserFamily) ||
		!deviceKeyPattern.MatchString(evidence.OSFamily) ||
		!validHMACKeyID(evidence.IPKeyID) ||
		evidence.IPHMAC == [32]byte{} {
		return ErrAuthenticationContract
	}
	return nil
}
