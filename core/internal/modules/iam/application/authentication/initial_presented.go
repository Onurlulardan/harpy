package authentication

import (
	"crypto/subtle"
	"encoding/base64"
)

type PresentedInitialPassword struct {
	selector [16]byte
	secret   [32]byte
	csrf     [32]byte
	active   bool
}

func ParseInitialPassword(cookie, csrf []byte) (*PresentedInitialPassword, error) {
	if len(cookie) != initialCookieBytes || cookie[22] != '.' ||
		len(csrf) != initialCSRFWireBytes {
		return nil, ErrInitialPasswordState
	}
	result := &PresentedInitialPassword{active: true}
	if !decodeInitialSegment(cookie[:22], result.selector[:]) ||
		!decodeInitialSegment(cookie[23:], result.secret[:]) ||
		!decodeInitialSegment(csrf, result.csrf[:]) ||
		result.selector == [16]byte{} || result.secret == [32]byte{} || result.csrf == [32]byte{} {
		result.Destroy()
		return nil, ErrInitialPasswordState
	}
	return result, nil
}

func decodeInitialSegment(wire, target []byte) bool {
	count, err := base64.RawURLEncoding.Strict().Decode(target, wire)
	return err == nil && count == len(target) &&
		base64.RawURLEncoding.EncodeToString(target) == string(wire)
}

func (presented *PresentedInitialPassword) Selector() ([16]byte, error) {
	if presented == nil || !presented.active {
		return [16]byte{}, ErrInitialPasswordState
	}
	return presented.selector, nil
}

func (presented *PresentedInitialPassword) matches(
	secretHMAC, csrfHMAC [32]byte,
	secretKey, csrfKey [32]byte,
	binding initialBinding,
) bool {
	if presented == nil || !presented.active {
		return false
	}
	actualSecret := initialSecretHMAC(secretKey, binding.tenantID, binding.userID,
		binding.operationID, binding.credentialID, binding.credentialVersion,
		presented.selector, presented.secret)
	actualCSRF := initialCSRFHMAC(csrfKey, presented.selector, presented.csrf)
	matched := subtle.ConstantTimeCompare(actualSecret[:], secretHMAC[:]) == 1 &&
		subtle.ConstantTimeCompare(actualCSRF[:], csrfHMAC[:]) == 1
	clear(actualSecret[:])
	clear(actualCSRF[:])
	return matched
}

func (presented *PresentedInitialPassword) Destroy() {
	if presented == nil {
		return
	}
	clear(presented.selector[:])
	clear(presented.secret[:])
	clear(presented.csrf[:])
	presented.active = false
}
