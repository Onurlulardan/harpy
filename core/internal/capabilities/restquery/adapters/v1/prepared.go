package v1query

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type Prepared struct {
	decoder  decoder
	pending  pendingResult
	evidence contractretirement.RuntimePort
	dialect  Dialect
	state    *preparedState
	valid    bool
}

type preparedState struct{ consumed atomic.Bool }

func Prepare(
	ctx context.Context,
	source VerifiedManifestSource,
	evidence contractretirement.RuntimePort,
	dialect Dialect,
	rawQuery string,
) (Prepared, error) {
	if ctx == nil || ctx.Err() != nil || nilInterface(evidence) ||
		dialect.Validate() != nil {
		return Prepared{}, fmt.Errorf("V1 query preparation is invalid")
	}
	decoded, err := newDecoder(source)
	if err != nil {
		return Prepared{}, err
	}
	if err = admitPreparedManifest(ctx, decoded.manifest, evidence); err != nil {
		return Prepared{}, err
	}
	pending, err := decoded.decode(rawQuery, dialect)
	if err != nil {
		return Prepared{}, err
	}
	prepared := Prepared{decoder: decoded, pending: pending, evidence: evidence,
		dialect: dialect, state: &preparedState{}, valid: true}
	if err = prepared.Validate(); err != nil {
		return Prepared{}, err
	}
	return prepared, nil
}

func (prepared Prepared) Validate() error {
	if !prepared.valid || prepared.state == nil || nilInterface(prepared.evidence) ||
		prepared.dialect.Validate() != nil || prepared.decoder.Validate() != nil ||
		prepared.pending.Validate() != nil ||
		prepared.pending.manifestID != prepared.decoder.manifest.id {
		return fmt.Errorf("prepared V1 query is invalid")
	}
	return nil
}

func (prepared Prepared) PagePlan() (PagePlan, error) {
	if err := prepared.Validate(); err != nil {
		return PagePlan{}, err
	}
	return prepared.pending.PagePlan()
}

func (prepared Prepared) Accept(
	ctx context.Context, authorizer AuthorizationPort,
) (Accepted, error) {
	if ctx == nil || ctx.Err() != nil || prepared.Validate() != nil ||
		nilInterface(authorizer) || !prepared.state.consumed.CompareAndSwap(false, true) {
		return Accepted{}, fmt.Errorf("prepared V1 query is invalid")
	}
	request, _ := prepared.pending.Request()
	plan, err := authorizer.ResolveAndAuthorizeV1(ctx, request)
	if err != nil {
		return Accepted{}, err
	}
	if err = validateAcceptedPlan(plan, prepared.pending.page); err != nil {
		return Accepted{}, fmt.Errorf("V1 authorized plan is invalid")
	}
	coordinator := Coordinator{decoder: prepared.decoder, evidence: prepared.evidence}
	deprecated, err := coordinator.recordHits(ctx, prepared.pending)
	if err != nil {
		return Accepted{}, err
	}
	return newAccepted(plan, prepared.pending.page, deprecated)
}

func admitPreparedManifest(
	ctx context.Context, manifest manifest, evidence contractretirement.RuntimePort,
) error {
	aliasIDs := manifest.aliasIDs()
	if len(aliasIDs) == 0 {
		return nil
	}
	batch, _, err := buildEvidenceBatch(manifest, aliasIDs)
	if err != nil {
		return err
	}
	decision, err := evidence.AdmitManifest(ctx, batch)
	if err != nil || decision.Validate() != nil {
		if err != nil {
			return fmt.Errorf("admit V1 alias manifest: %w", err)
		}
		return fmt.Errorf("admit V1 alias manifest: invalid decision")
	}
	retired, err := decision.RetiredCount()
	if err != nil || retired != 0 {
		return contractretirement.NewFault(contractretirement.FaultRetired)
	}
	return nil
}

func (Prepared) String() string                          { return v1Redaction }
func (Prepared) GoString() string                        { return v1Redaction }
func (prepared Prepared) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (Prepared) MarshalJSON() ([]byte, error)            { return rejectV1JSON() }
