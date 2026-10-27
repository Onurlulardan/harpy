package accesstoken

type headerDocument struct {
	Type      string `json:"typ"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type claimsDocument struct {
	Issuer                     string   `json:"iss"`
	Audience                   string   `json:"aud"`
	Subject                    string   `json:"sub"`
	Session                    string   `json:"sid"`
	Tenant                     string   `json:"tenant_id"`
	Client                     string   `json:"client_id"`
	IssuedAt                   int64    `json:"iat"`
	NotBefore                  int64    `json:"nbf"`
	ExpiresAt                  int64    `json:"exp"`
	JTI                        string   `json:"jti"`
	TokenVersion               int64    `json:"token_version"`
	AuthorizationVersion       int64    `json:"authorization_version"`
	PlatformAuthorizationClock int64    `json:"platform_authorization_clock"`
	SecurityEpoch              string   `json:"security_epoch"`
	AccessMode                 string   `json:"access_mode"`
	PortalBinding              string   `json:"portal_binding_id,omitempty"`
	AuthenticationTime         int64    `json:"auth_time"`
	Assurance                  string   `json:"acr"`
	Methods                    []string `json:"amr"`
	StepUpPurposeHash          string   `json:"step_up_purpose_hash,omitempty"`
}

func (claims Claims) document() claimsDocument {
	methods := make([]string, len(claims.methods))
	for index, method := range claims.methods {
		methods[index] = string(method)
	}
	document := claimsDocument{
		Issuer: claims.issuer, Audience: claims.audience,
		Subject: claims.subject.String(), Session: claims.session.String(),
		Tenant: claims.tenant.String(), Client: string(claims.client),
		IssuedAt: claims.issuedAt, NotBefore: claims.issuedAt,
		ExpiresAt: claims.expiresAt, JTI: claims.jti.String(),
		TokenVersion:               claims.tokenVersion,
		AuthorizationVersion:       claims.authorizationVersion,
		PlatformAuthorizationClock: claims.platformAuthorizationClock,
		SecurityEpoch:              encodeBase64URL(claims.securityEpoch[:]),
		AccessMode:                 string(claims.accessMode), AuthenticationTime: claims.authenticationTime,
		Assurance: string(claims.assurance), Methods: methods,
	}
	if claims.hasPortalBinding {
		document.PortalBinding = claims.portalBinding.String()
	}
	if claims.hasStepUpPurpose {
		document.StepUpPurposeHash = encodeBase64URL(claims.stepUpPurposeHash[:])
	}
	return document
}
