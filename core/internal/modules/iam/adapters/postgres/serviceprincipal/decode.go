package serviceprincipal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	"github.com/google/uuid"
)

func decodeEvidence(
	raw []byte,
) (serviceapp.Principal, serviceapp.GrantRevision, error) {
	document, err := decodeEvidenceDocument(raw)
	if err != nil {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, err
	}
	principal := serviceapp.Principal{
		ID: document.Principal.ID, Kind: serviceapp.Kind(document.Principal.Kind),
		State:                serviceapp.PrincipalState(document.Principal.State),
		AuthorizationVersion: document.Principal.AuthorizationVersion,
		Revision:             document.Principal.Revision,
	}
	capability, err := serviceapp.ParseCapabilityKey(document.Grant.Capability)
	if err != nil {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, err
	}
	manifest, err := hex.DecodeString(document.Grant.ManifestSHA256)
	if err != nil || len(manifest) != 32 {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	var resourceID uuid.UUID
	if document.Grant.ResourceID != nil {
		resourceID = *document.Grant.ResourceID
	}
	var validUntil time.Time
	if document.Grant.ValidUntil != nil {
		validUntil = *document.Grant.ValidUntil
	}
	var digest [32]byte
	copy(digest[:], manifest)
	grant := serviceapp.GrantRevision{
		GrantID: document.Grant.ID, Revision: document.Grant.Revision,
		PreviousRevision: document.Grant.PreviousRevision,
		PrincipalID:      document.Grant.PrincipalID, Capability: capability,
		Scope: serviceapp.ResourceScope{
			Kind:  serviceapp.ScopeKind(document.Grant.ScopeKind),
			Owner: document.Grant.Owner, ResourceKind: document.Grant.ResourceKind,
			ResourceID: resourceID,
		},
		ValidFrom: document.Grant.ValidFrom, ValidUntil: validUntil,
		State: serviceapp.GrantState(document.Grant.State), ManifestSHA256: digest,
	}
	if principal.Validate() != nil || grant.Validate() != nil {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	return principal, grant, nil
}

func decodeEvidenceDocument(raw []byte) (evidenceDocument, error) {
	var document evidenceDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return evidenceDocument{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return evidenceDocument{}, errEvidenceRead
	}
	return document, nil
}
