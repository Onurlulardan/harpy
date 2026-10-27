package adoption

import "github.com/google/uuid"

func validMapping() Mapping {
	id := fixtureIDs
	generated := func(suffix string) uuid.UUID {
		return uuid.MustParse("20000000-0000-0000-0000-" + suffix)
	}
	resourceTarget := generated("000000000001")
	actionTarget := generated("000000000002")
	permissionTarget := generated("000000000003")
	bindingTarget := generated("000000000004")
	return Mapping{
		ContractVersion: MappingContractVersion,
		TenantID:        id.tenant,
		Identity: IdentityMapping{
			Users: []UserMapping{{
				UserID: id.user, EmailIdentifierID: pointer(generated("000000000005")),
				OperatorModeID: pointer(generated("000000000006")),
				MembershipID:   pointer(generated("000000000007")),
				Assignments: []AssignmentMapping{{
					RoleID: id.role, AssignmentID: generated("000000000008"),
				}},
			}},
			ProtectedRoleID:       id.role,
			ProtectedUserID:       id.user,
			ProtectedPermissionID: permissionTarget,
			ProtectedBindingID:    bindingTarget,
			ProtectedPolicyID:     generated("000000000009"),
		},
		Resources: []ResourceMapping{{
			SourceID: id.usersResource, TargetID: resourceTarget,
			ResourceKey: "users", OwnerDomain: "iam", ResourceKind: "user",
		}},
		Actions: []ActionMapping{{
			SourceID: id.readAction, TargetID: actionTarget, ActionKey: "read",
		}},
		Permissions: []PermissionMapping{{
			SourcePermissionID: id.permission, ResourceID: resourceTarget,
			ActionID: actionTarget, PermissionID: permissionTarget, BindingID: bindingTarget,
		}},
		PermissionRequests: []PermissionRequestMapping{{
			SourceRequestID: id.request, GroupID: generated("00000000000a"),
			RequestID: generated("00000000000b"),
		}},
		Invitations: []InvitationMapping{{
			SourceInvitationID: id.invitation, InvitationID: generated("00000000000c"),
			TokenID: generated("00000000000d"),
		}},
		ProfileSchemas: []ProfileSchemaMapping{{
			ResourceKind: "user", SchemaVersionID: id.schema,
		}},
		ProfileFields: []ProfileFieldMapping{{
			ResourceKind: "user", SourceKey: "nickname", FieldID: id.field,
		}},
		LayoutFields: []LayoutFieldMapping{{
			ResourceKind: "user", SourceKey: "nickname", LayoutID: id.layout, FieldID: id.field,
		}},
	}
}
