package iam

import (
	"net/http"

	"github.com/google/uuid"
)

func rbacRevisionedMutation(r *http.Request) (RBACMutation, bool) {
	id, idValid := identityPathID(r)
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	return rbacMutationFromRequest(id, key, revision),
		idValid && keyValid && revisionValid
}

func rbacCreateKey(r *http.Request) (string, bool) {
	return identityIdempotencyKey(r)
}

func rbacPathID(r *http.Request) (uuid.UUID, bool) {
	return identityPathID(r)
}
