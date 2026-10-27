package main

import "fmt"

func ownedProcessAlive(receipt processReceipt) (bool, error) {
	identity, alive := processIdentity(receipt.PID)
	if !alive {
		return false, nil
	}
	if identity != receipt.ProcessIdentity {
		return false, fmt.Errorf(
			"PID %d was reused by another process; it will not be terminated",
			receipt.PID,
		)
	}
	return true, nil
}
