package platformiam

import (
	"crypto/ed25519"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type AcknowledgementTrust interface {
	ResolveRegistryKey(string) (ed25519.PublicKey, bool)
}

type OwnerTransferService struct {
	store platformports.OwnerTransferStore[platformtx.TenantScope]
	clock Clock
	trust AcknowledgementTrust
}

func NewOwnerTransferService(
	store platformports.OwnerTransferStore[platformtx.TenantScope],
	clock Clock,
	trust AcknowledgementTrust,
) (*OwnerTransferService, error) {
	if nilPort(store) || nilPort(clock) || nilPort(trust) {
		return nil, platformdomain.ErrStorage
	}
	return &OwnerTransferService{store: store, clock: clock, trust: trust}, nil
}
