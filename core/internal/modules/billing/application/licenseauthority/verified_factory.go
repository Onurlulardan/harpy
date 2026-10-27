package licenseauthority

func verifiedFrom(
	payload parsedPayload,
	envelope wireEnvelope,
	document, canonical []byte,
) VerifiedSnapshot {
	return VerifiedSnapshot{
		licenseID: payload.licenseID, organizationID: payload.organizationID,
		environmentID: payload.environmentID, issuer: payload.wire.Issuer,
		keyID: envelope.KeyID, digest: envelope.PayloadSHA256,
		releaseChannel: payload.wire.ReleaseChannel, delivery: payload.wire.DeliveryKind,
		status: payload.wire.Status, sequence: payload.wire.Sequence,
		issuedAt: payload.issuedAt, notBefore: payload.notBefore,
		refreshAfter: payload.refreshAfter, grantValidUntil: payload.grantValidUntil,
		graceUntil: payload.graceUntil, commercialExpiresAt: payload.commercialExpiresAt,
		features: FeatureClaims{
			FormBuilder:      payload.wire.Features.FormBuilder,
			WorkflowBuilder:  payload.wire.Features.WorkflowBuilder,
			APIBuilder:       payload.wire.Features.APIBuilder,
			DashboardBuilder: payload.wire.Features.DashboardBuilder,
			AIAssistant:      payload.wire.Features.AIAssistant,
		},
		limits: LimitClaims{
			MaxUsers:        payload.wire.Limits.MaxUsers,
			MaxForms:        payload.wire.Limits.MaxForms,
			MaxWorkflows:    payload.wire.Limits.MaxWorkflows,
			MaxAPIEndpoints: payload.wire.Limits.MaxAPIEndpoints,
			MaxDashboards:   payload.wire.Limits.MaxDashboards,
			MaxTenants:      payload.wire.Limits.MaxTenants,
		},
		canonicalPayload: append([]byte(nil), canonical...),
		signedDocument:   append([]byte(nil), document...),
	}
}
