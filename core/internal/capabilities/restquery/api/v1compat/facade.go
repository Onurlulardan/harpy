package v1compat

import (
	"context"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type Dialect = v1query.Dialect
type Prepared = v1query.Prepared
type PagePlan = v1query.PagePlan
type ManifestSnapshot = v1query.ManifestSnapshot
type VerifiedManifestSource = v1query.VerifiedManifestSource
type FilterShape = v1query.FilterShape
type FilterFieldBinding = v1query.FilterFieldBinding
type SortFieldBinding = v1query.SortFieldBinding
type OperatorBinding = v1query.OperatorBinding
type QueryKeyBinding = v1query.QueryKeyBinding
type RetirementPort = contractretirement.RuntimePort
type RetirementIdentity = contractretirement.Identity
type RetirementKey = contractretirement.Key
type RetirementBatch = contractretirement.Batch
type RetirementDecision = contractretirement.Decision

const (
	DialectDynamicList = v1query.DialectDynamicList
	DialectOpenAPIGet  = v1query.DialectOpenAPIGet

	FilterShapeScalar     = v1query.FilterShapeScalar
	FilterShapeCollection = v1query.FilterShapeCollection

	WarningHeaderName      = v1query.WarningHeaderName
	DeprecatedWarningValue = v1query.DeprecatedWarningValue
)

func Prepare(
	ctx context.Context,
	source VerifiedManifestSource,
	evidence contractretirement.RuntimePort,
	dialect Dialect,
	rawQuery string,
) (Prepared, error) {
	return v1query.Prepare(ctx, source, evidence, dialect, rawQuery)
}

func ValidateManifestSnapshot(snapshot ManifestSnapshot) error {
	return v1query.ValidateManifestSnapshot(snapshot)
}

func NewRetirementIdentity(owner, domain string) (RetirementIdentity, error) {
	return contractretirement.NewIdentity(owner, domain)
}

func ParseRetirementKey(value string) (RetirementKey, error) {
	return contractretirement.ParseKey(value)
}

func NewRetirementBatch(
	identity RetirementIdentity, keys []RetirementKey,
) (RetirementBatch, error) {
	return contractretirement.NewBatch(identity, keys)
}

func SealRetirementDecision(
	batch RetirementBatch, retired []RetirementKey,
) (RetirementDecision, error) {
	return contractretirement.SealDecision(batch, retired)
}
