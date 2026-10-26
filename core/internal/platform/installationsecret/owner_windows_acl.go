//go:build windows

package installationsecret

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func allowedSIDs() (*windows.SID, *windows.SID, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, nil, err
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return nil, nil, err
	}
	return user.User.Sid, system, nil
}

func validateOwnerOnly(handle windows.Handle, user, system *windows.SID) error {
	descriptor, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil || descriptor == nil {
		return ErrFenced
	}
	control, _, err := descriptor.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return ErrFenced
	}
	owner, _, err := descriptor.Owner()
	if err != nil || (!owner.Equals(user) && !owner.Equals(system)) {
		return ErrFenced
	}
	dacl, _, err := descriptor.DACL()
	expected := uint16(2)
	if user.Equals(system) {
		expected = 1
	}
	if err != nil || dacl == nil || dacl.AceCount != expected {
		return ErrFenced
	}
	seenUser, seenSystem := false, false
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err = windows.GetAce(dacl, index, &ace); err != nil || ace == nil ||
			ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Mask != fileAllAccess {
			return ErrFenced
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		switch {
		case sid.Equals(user) && !seenUser:
			seenUser = true
		case sid.Equals(system) && !seenSystem:
			seenSystem = true
		default:
			return ErrFenced
		}
	}
	if !seenUser || (!user.Equals(system) && !seenSystem) {
		return ErrFenced
	}
	return nil
}
