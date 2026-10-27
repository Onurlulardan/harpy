//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMFoundationAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			t.Run("catalog", func(t *testing.T) {
				assertIAMFoundationCatalog(t, ctx, database)
			})
			t.Run("privileges", func(t *testing.T) {
				assertIAMFoundationHasNoRuntimeDML(t, ctx, database)
			})
			t.Run("platform_auth", func(t *testing.T) {
				assertIAMPlatformAuthConstraints(t, ctx, database)
			})
			t.Run("directory", func(t *testing.T) {
				assertIAMLoginDirectoryConstraints(t, ctx, database)
			})
			t.Run("profile", func(t *testing.T) {
				assertIAMProfileConstraints(t, ctx, database)
			})
			t.Run("identifiers", func(t *testing.T) {
				assertIAMIdentifierConstraints(t, ctx, database)
			})
			t.Run("organization_roles", func(t *testing.T) {
				assertIAMOrganizationRoleConstraints(t, ctx, database)
			})
			t.Run("password_credentials", func(t *testing.T) {
				assertIAMPasswordCredentialConstraints(t, ctx, database)
			})
			t.Run("human_access", func(t *testing.T) {
				assertIAMHumanAccessConstraints(t, ctx, database)
			})
			t.Run("invitation_token", func(t *testing.T) {
				assertIAMInvitationTokenConstraints(t, ctx, database)
			})
			t.Run("session_refresh", func(t *testing.T) {
				assertIAMSessionRefreshConstraints(t, ctx, database)
			})
			t.Run("authenticator_challenge", func(t *testing.T) {
				assertIAMAuthenticatorChallengeConstraints(t, ctx, database)
			})
		})
	}
}
