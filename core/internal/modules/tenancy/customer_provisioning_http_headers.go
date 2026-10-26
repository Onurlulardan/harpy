package tenancy

import (
	"net/http"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func customerProvisioningCredential(r *http.Request) provisioning.CredentialEnvelope {
	if r == nil {
		return provisioning.CredentialEnvelope{}
	}
	return provisioning.CredentialEnvelope{
		AuthorizationValues: r.Header.Values("Authorization"),
		TenantIDValues:      r.Header.Values("X-Tenant-ID"),
		TenantSlugValues:    r.Header.Values("X-Tenant-Slug"),
	}
}

func customerProvisioningIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	values := r.Header.Values("Idempotency-Key")
	if len(values) != 1 || len(values[0]) < 1 || len(values[0]) > 128 {
		return "", false
	}
	for _, character := range values[0] {
		if character < 0x21 || character > 0x7e {
			return "", false
		}
	}
	return values[0], true
}
