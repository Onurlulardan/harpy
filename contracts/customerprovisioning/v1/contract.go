package customerprovisioningv1

import (
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
)

const (
	CommandProvisionCustomerTenant = "tenancy.provision_customer_tenant"
	CapabilityCreate               = "platform.tenants.create"
	CapabilityRead                 = "platform.tenants.read"
	CapabilityAdminSetupReissue    = "platform.tenants.admin_setup.reissue"
	SetupCredentialPurpose         = "initial_admin_setup.v1"
)

var ErrPurpose = errors.New("customer provisioning step-up purpose is invalid")

func ProvisionStepUpPurpose(
	controlTenantID uuid.UUID,
	idempotencyKey, requestSHA256 string,
) ([sha256.Size]byte, error) {
	if controlTenantID == uuid.Nil || len(idempotencyKey) < 1 || len(idempotencyKey) > 128 ||
		len(requestSHA256) != sha256.Size*2 {
		return [sha256.Size]byte{}, ErrPurpose
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/tenancy/provision-customer/v1\x00"))
	_, _ = hash.Write(controlTenantID[:])
	_, _ = hash.Write([]byte(idempotencyKey))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(requestSHA256))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func AdminSetupReissueStepUpPurpose(
	controlTenantID, targetTenantID uuid.UUID,
	idempotencyKey string,
) ([sha256.Size]byte, error) {
	if controlTenantID == uuid.Nil || targetTenantID == uuid.Nil ||
		controlTenantID == targetTenantID || len(idempotencyKey) < 1 ||
		len(idempotencyKey) > 128 {
		return [sha256.Size]byte{}, ErrPurpose
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/tenancy/admin-setup-reissue/v1\x00"))
	_, _ = hash.Write(controlTenantID[:])
	_, _ = hash.Write(targetTenantID[:])
	_, _ = hash.Write([]byte(idempotencyKey))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
