package api

import internal "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool"

type Manager = internal.Manager
type ReservationIdentity = internal.ReservationIdentity
type ActiveLeaseLookup = internal.ActiveLeaseLookup
type ReapResult = internal.ReapResult
type Lease = internal.Lease

var (
	ErrCapacity        = internal.ErrCapacity
	ErrAlreadyReserved = internal.ErrAlreadyReserved
)

func NewManager(root string) (*Manager, error) {
	return internal.NewManager(root)
}
