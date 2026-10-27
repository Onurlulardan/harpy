package iam

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type portalUserCreateDocument struct {
	Email           string `json:"email,omitempty"`
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	SubjectType     string `json:"subject_type"`
	SubjectID       string `json:"subject_id"`
	ProfileFormSlug string `json:"portal_profile_form_slug"`
	Status          string `json:"status,omitempty"`
}

type portalUserStatusDocument struct {
	Status string `json:"status"`
}
type portalUserEmptyDocument struct{}

func portalUserPathAndKey(r *http.Request) (uuid.UUID, string, bool) {
	id, idValid := identityPathID(r)
	key, keyValid := identityIdempotencyKey(r)
	return id, key, idValid && keyValid
}

func portalUserMutation(r *http.Request) (PortalUserMutationInput, bool) {
	mutation, valid := identityMutationFromRequest(r)
	return PortalUserMutationInput{ID: mutation.ID,
		ExpectedRevision: mutation.ExpectedRevision,
		IdempotencyKey:   mutation.IdempotencyKey}, valid
}

func normalizePortalStatus(value string) string { return strings.ToUpper(strings.TrimSpace(value)) }
