package packageregistry

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Reader interface {
	GetPackage(context.Context, string) (Package, error)
	ReadPackageBySlug(context.Context, string) (Package, error)
	ListPackages(context.Context, int, string) ([]Package, error)
	ResolveVersion(context.Context, string, string) (SealedVersion, error)
	CompileCandidates(context.Context, string, int) ([]SealedVersion, error)
}

type PayloadReader interface {
	VersionPayloads(
		context.Context,
		string,
		string,
	) (SealedVersion, error)
}

type Binder interface {
	Bind(context.Context, platformtx.TenantScope) (Reader, error)
}

type Registrar interface {
	Register(context.Context, SealedVersion) error
}
