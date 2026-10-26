package platformops

import (
	"encoding/base64"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type recoveryAnchorView struct {
	AnchorID         string `json:"anchor_id"`
	Generation       int64  `json:"generation"`
	CredentialID     string `json:"credential_id_base64url"`
	RPID             string `json:"rp_id"`
	UserVerification string `json:"user_verification"`
	BackupEligible   bool   `json:"backup_eligible"`
	BackupState      bool   `json:"backup_state"`
	State            string `json:"state"`
}

func representRecoveryAnchors(values []platformopsapi.RecoveryAnchor) []recoveryAnchorView {
	result := make([]recoveryAnchorView, len(values))
	for index := range values {
		result[index] = recoveryAnchorView{AnchorID: values[index].AnchorID,
			Generation:       values[index].Generation,
			CredentialID:     base64.RawURLEncoding.EncodeToString(values[index].CredentialID),
			RPID:             values[index].RPID,
			UserVerification: values[index].UserVerification,
			BackupEligible:   values[index].BackupEligible,
			BackupState:      values[index].BackupState, State: values[index].State}
	}
	return result
}

type recoveryAnchorPublicKey struct {
	Challenge              string                    `json:"challenge"`
	RP                     recoveryAnchorRP          `json:"rp"`
	User                   recoveryAnchorUser        `json:"user"`
	PubKeyCredParams       []recoveryAnchorAlgorithm `json:"pubKeyCredParams"`
	Timeout                int                       `json:"timeout"`
	Attestation            string                    `json:"attestation"`
	AuthenticatorSelection struct {
		AuthenticatorAttachment string `json:"authenticatorAttachment"`
		ResidentKey             string `json:"residentKey"`
		RequireResidentKey      bool   `json:"requireResidentKey"`
		UserVerification        string `json:"userVerification"`
	} `json:"authenticatorSelection"`
}

type recoveryAnchorRP struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type recoveryAnchorUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type recoveryAnchorAlgorithm struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

func recoveryAnchorOptionsView(value platformopsapi.RecoveryAnchorRegistrationOptions) any {
	publicKey := recoveryAnchorPublicKey{Challenge: value.Challenge,
		Timeout: 300000, Attestation: "direct"}
	publicKey.RP.ID, publicKey.RP.Name = value.RPID, "Harpy Recovery"
	publicKey.User.ID, publicKey.User.Name = value.UserID, "instance-owner"
	publicKey.User.DisplayName = "Harpy Instance Owner"
	publicKey.PubKeyCredParams = append(publicKey.PubKeyCredParams,
		recoveryAnchorAlgorithm{Type: "public-key", Alg: -7},
		recoveryAnchorAlgorithm{Type: "public-key", Alg: -257})
	publicKey.AuthenticatorSelection.AuthenticatorAttachment = "cross-platform"
	publicKey.AuthenticatorSelection.ResidentKey = "required"
	publicKey.AuthenticatorSelection.RequireResidentKey = true
	publicKey.AuthenticatorSelection.UserVerification = "required"
	return struct {
		OperationID string                  `json:"operation_id"`
		PublicKey   recoveryAnchorPublicKey `json:"public_key"`
	}{OperationID: value.OperationID, PublicKey: publicKey}
}
