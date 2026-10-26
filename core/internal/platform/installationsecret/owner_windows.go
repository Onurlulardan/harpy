//go:build windows

package installationsecret

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const fileAllAccess windows.ACCESS_MASK = 0x001f01ff

func createOwnerOnly(path string) (*os.File, error) {
	user, system, err := allowedSIDs()
	if err != nil {
		return nil, err
	}
	sddl := fmt.Sprintf("O:%sD:P(A;;FA;;;%s)(A;;FA;;;SY)", user.String(), user.String())
	if user.Equals(system) {
		sddl = "O:SYD:P(A;;FA;;;SY)"
	}
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return nil, err
	}
	attributes := windows.SecurityAttributes{
		Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: descriptor,
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE,
		0, &attributes, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), path)
	if err = validateOwnerOnly(handle, user, system); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return nil, err
	}
	return file, nil
}

func openOwnerOnly(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil,
		windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	user, system, sidErr := allowedSIDs()
	if sidErr != nil || validateOwnerOnly(handle, user, system) != nil {
		windows.CloseHandle(handle)
		return nil, ErrFenced
	}
	return os.NewFile(uintptr(handle), path), nil
}

func publishNoReplace(staging, target string) error {
	from, fromErr := windows.UTF16PtrFromString(staging)
	to, toErr := windows.UTF16PtrFromString(target)
	if fromErr != nil || toErr != nil {
		return ErrInvalid
	}
	if err := windows.MoveFileEx(from, to, windows.MOVEFILE_WRITE_THROUGH); err != nil {
		if _, statErr := os.Lstat(target); statErr == nil {
			return ErrExists
		}
		return ErrInterrupted
	}
	return nil
}
