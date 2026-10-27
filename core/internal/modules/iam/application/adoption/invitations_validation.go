package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateInvitations() {
	mappings := make(map[uuid.UUID]InvitationMapping)
	invitationIDs := make(map[uuid.UUID]uuid.UUID)
	tokenIDs := make(map[uuid.UUID]uuid.UUID)
	for _, mapping := range state.mapping.Invitations {
		if _, duplicate := mappings[mapping.SourceInvitationID]; duplicate {
			state.add(issue("mapping_duplicate", "invitations", mapping.SourceInvitationID.String()))
		}
		mappings[mapping.SourceInvitationID] = mapping
		state.claimInvitationID(invitationIDs, mapping.InvitationID, mapping.SourceInvitationID)
		state.claimInvitationID(tokenIDs, mapping.TokenID, mapping.SourceInvitationID)
	}
	seen := make(map[uuid.UUID]bool)
	for _, source := range state.snapshot.Invitations {
		state.validateInvitation(source, mappings[source.ID])
		if source.ID == uuid.Nil || seen[source.ID] {
			state.add(issue("invitation_identity_invalid", "invitations", source.ID.String()))
		}
		seen[source.ID] = true
	}
	for sourceID := range mappings {
		if !seen[sourceID] {
			state.add(issue("mapping_source_missing", "invitations", sourceID.String()))
		}
	}
}

func (state *validationState) validateInvitation(
	source SourceInvitation, mapping InvitationMapping,
) {
	if _, err := state.normalizers.email(source.Email); err != nil {
		state.add(issue("invitation_email_invalid", "invitations", source.ID.String()))
	}
	roleID, err := uuid.Parse(source.RoleID)
	if err != nil || state.roles[roleID].ID == uuid.Nil {
		state.add(issue("invitation_role_dangling", "invitations", source.ID.String()))
	}
	switch strings.ToUpper(strings.TrimSpace(source.State)) {
	case "PENDING":
		if mapping.SourceInvitationID == uuid.Nil || mapping.InvitationID == uuid.Nil ||
			mapping.TokenID == uuid.Nil {
			state.add(issue("invitation_reissue_mapping_missing", "invitations", source.ID.String()))
		}
	case "ACCEPTED", "EXPIRED":
		if mapping.SourceInvitationID != uuid.Nil {
			state.add(issue("historical_invitation_mapping_unexpected", "invitations", source.ID.String()))
		}
	default:
		state.add(issue("invitation_state_invalid", "invitations", source.ID.String()))
	}
}

func (state *validationState) claimInvitationID(
	seen map[uuid.UUID]uuid.UUID, id, sourceID uuid.UUID,
) {
	if id == uuid.Nil || sourceID == uuid.Nil {
		state.add(issue("invitation_mapping_invalid", "invitations", sourceID.String()))
		return
	}
	if prior, duplicate := seen[id]; duplicate && prior != sourceID {
		state.add(issue("invitation_id_collision", "invitations", id.String()))
	}
	seen[id] = sourceID
}
