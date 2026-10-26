package postgres

import (
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

type ExternalEffectPermit struct {
	owner     *workerIdentity
	claim     JobClaim
	request   effectsv1.Request
	startedAt time.Time
}

func newExternalEffectPermit(
	owner *workerIdentity,
	claim JobClaim,
	request effectsv1.Request,
	startedAt time.Time,
) (ExternalEffectPermit, error) {
	permit := ExternalEffectPermit{
		owner: owner, claim: claim, request: request, startedAt: startedAt,
	}
	if !permit.valid(owner) {
		return ExternalEffectPermit{}, ErrExternalEffectPermit
	}
	return permit, nil
}

func (permit ExternalEffectPermit) Request() effectsv1.Request {
	document, err := permit.request.MarshalJSON()
	if err != nil {
		return effectsv1.Request{}
	}
	value, err := effectsv1.DecodeRequest(document)
	if err != nil {
		return effectsv1.Request{}
	}
	return value
}

func (permit ExternalEffectPermit) StartedAt() time.Time { return permit.startedAt }

func (permit ExternalEffectPermit) valid(owner *workerIdentity) bool {
	if permit.owner != owner || !permit.claim.valid(owner) ||
		permit.claim.effectClass != "external" || permit.startedAt.IsZero() ||
		permit.request.Validate() != nil ||
		permit.request.RequestedAt != permit.startedAt.UTC().Format(time.RFC3339Nano) {
		return false
	}
	request := permit.request
	return request.EffectKind == permit.claim.kind &&
		request.EffectRevision == permit.claim.payloadRevision &&
		request.IdempotencyKey == permit.claim.effectKey &&
		request.PayloadReference == permit.claim.payloadReference &&
		request.PayloadSHA256 == permit.claim.payloadSHA256
}
