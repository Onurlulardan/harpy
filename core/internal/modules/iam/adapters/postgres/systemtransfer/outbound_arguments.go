package systemtransfer

import app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"

func outboundAttemptArguments(schema string, attempt app.ExportAttempt) []any {
	return []any{schema, attempt.TenantID, attempt.ActorID, attempt.ExportID,
		attempt.JobID, attempt.AttemptID, attempt.LeaseOwnerID,
		attempt.AttemptGeneration, string(attempt.Resource), attempt.TokenVersion,
		attempt.AuthorizationVersion}
}

func outboundEffectArguments(schema string,
	effect app.ExportArtifactEffectStart,
) []any {
	arguments := outboundAttemptArguments(schema, effect.Attempt)
	return append(arguments, effect.EffectID, effect.EffectReceiptID,
		effect.ObjectKey, effect.Draft.KeyID, effect.Draft.HeaderSHA256[:],
		effect.Draft.CiphertextSHA256[:], effect.Draft.CiphertextBytes,
		append([]byte(nil), effect.Draft.WrappedDEK...), effect.Draft.WrapNonce[:],
		effect.RequestSHA256[:], effect.StartedAt)
}
