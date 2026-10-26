package operator

import (
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

var ErrEnvironmentFinalizerInvalid = errors.New(
	"environment activation finalizer input is invalid",
)

type EnvironmentFinalizerInput struct {
	TenantName, TenantSlug       string
	AdminEmail, AdminDisplayName string
	DirectoryKeyID               string
	IdentifierHMAC               [32]byte
	DirectorySourceSHA256        [32]byte
	InitialCredential            iam.PreparedInstallationCredential
	ActivatedAt                  time.Time
}

func validEnvironmentFinalizerInput(input EnvironmentFinalizerInput) bool {
	return input.TenantName != "" && input.TenantSlug != "" &&
		input.AdminEmail != "" && input.AdminDisplayName != "" &&
		input.DirectoryKeyID != "" && input.IdentifierHMAC != [32]byte{} &&
		input.DirectorySourceSHA256 != [32]byte{} &&
		input.InitialCredential.Valid() &&
		!input.ActivatedAt.IsZero() && input.ActivatedAt.Location() == time.UTC
}
