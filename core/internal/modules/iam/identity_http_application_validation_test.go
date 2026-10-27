package iam

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/google/uuid"
)

func TestIdentityApplicationCRUDAndStateRejectInvalidContractsBeforeIO(t *testing.T) {
	application := &identityHTTPApplication{}
	claims := accesstoken.Claims{}
	tests := []struct {
		name string
		run  func() error
	}{
		{"create user", func() error {
			_, err := application.CreateUser(context.Background(), claims, IdentityUserCreate{})
			return err
		}},
		{"get user", func() error {
			_, err := application.GetUser(context.Background(), claims, uuid.Nil)
			return err
		}},
		{"update user", func() error {
			_, err := application.UpdateUser(context.Background(), claims, IdentityUserUpdate{})
			return err
		}},
		{"delete user", func() error {
			return application.DeleteUser(context.Background(), claims, IdentityMutation{})
		}},
		{"state user", func() error {
			_, err := application.ChangeUserState(
				context.Background(), claims, IdentityStateMutation{State: "ACTIVE"},
			)
			return err
		}},
		{"create role", func() error {
			_, err := application.CreateRole(context.Background(), claims, IdentityRoleCreate{})
			return err
		}},
		{"get role", func() error {
			_, err := application.GetRole(context.Background(), claims, uuid.Nil)
			return err
		}},
		{"update role", func() error {
			_, err := application.UpdateRole(context.Background(), claims, IdentityRoleUpdate{})
			return err
		}},
		{"delete role", func() error {
			return application.DeleteRole(context.Background(), claims, IdentityMutation{})
		}},
		{"state role", func() error {
			_, err := application.ChangeRoleState(
				context.Background(), claims, IdentityStateMutation{State: "ACTIVE"},
			)
			return err
		}},
		{"create organization", func() error {
			_, err := application.CreateOrganization(
				context.Background(), claims, IdentityOrganizationCreate{},
			)
			return err
		}},
		{"get organization", func() error {
			_, err := application.GetOrganization(context.Background(), claims, uuid.Nil)
			return err
		}},
		{"update organization", func() error {
			_, err := application.UpdateOrganization(
				context.Background(), claims, IdentityOrganizationUpdate{},
			)
			return err
		}},
		{"delete organization", func() error {
			return application.DeleteOrganization(context.Background(), claims, IdentityMutation{})
		}},
		{"state organization", func() error {
			_, err := application.ChangeOrganizationState(
				context.Background(), claims, IdentityStateMutation{State: "ACTIVE"},
			)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, ErrIdentityHTTPInvalid) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestIdentityApplicationQueryFailsClosedWithoutRuntime(t *testing.T) {
	var application *identityHTTPApplication
	_, err := application.Query(
		context.Background(), accesstoken.Claims{}, IdentityUsers, queryapp.Request{},
	)
	if !errors.Is(err, ErrIdentityHTTPUnavailable) {
		t.Fatalf("error=%v", err)
	}
}
