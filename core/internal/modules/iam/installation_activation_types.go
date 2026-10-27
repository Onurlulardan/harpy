package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"strings"
)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationIdentityInput struct {
	OperationID, EnvironmentID, TenantID [16]byte
	FirstUserID, RootOrganizationID      [16]byte
	ProtectedAdminBindingID              [16]byte
	Email, DisplayName, TenantName       string
	SchemaName                           string
	DirectoryKeyID                       string
	IdentifierHMAC, SourceSHA256         [32]byte
}

type installationIdentityIDs struct {
	userSchema, userField, userSchemaV2      [16]byte
	userFirstName, userLastName, userPhone   [16]byte
	organizationSchema                       [16]byte
	organizationField, roleSchema, roleField [16]byte
	identifier, accessMode, membership       [16]byte
	role, permission, rolePermission         [16]byte
}

func installationIDs(operationID [16]byte) installationIdentityIDs {
	return installationIdentityIDs{
		userSchema:         deriveInstallationID(operationID, "user-profile-schema"),
		userField:          deriveInstallationID(operationID, "user-display-name-field"),
		userSchemaV2:       deriveInstallationID(operationID, "user-profile-schema-v2"),
		userFirstName:      deriveInstallationID(operationID, "user-first-name-field"),
		userLastName:       deriveInstallationID(operationID, "user-last-name-field"),
		userPhone:          deriveInstallationID(operationID, "user-phone-field"),
		organizationSchema: deriveInstallationID(operationID, "organization-profile-schema"),
		organizationField:  deriveInstallationID(operationID, "organization-display-name-field"),
		roleSchema:         deriveInstallationID(operationID, "role-profile-schema"),
		roleField:          deriveInstallationID(operationID, "role-display-name-field"),
		identifier:         deriveInstallationID(operationID, "first-admin-email"),
		accessMode:         deriveInstallationID(operationID, "first-admin-operator-console"),
		membership:         deriveInstallationID(operationID, "root-organization-membership"),
		role:               deriveInstallationID(operationID, "protected-tenant-admin-role"),
		permission:         deriveInstallationID(operationID, "protected-tenant-all-permission"),
		rolePermission:     deriveInstallationID(operationID, "protected-role-permission-binding"),
	}
}

func InstallationFirstEmailIdentifierID(operationID [16]byte) [16]byte {
	if zeroInstallationID(operationID) {
		return [16]byte{}
	}
	return installationIDs(operationID).identifier
}

func ProtectedAdminBindingSetSHA256(
	operationID, tenantID, userID, protectedBindingID [16]byte,
) string {
	if zeroInstallationID(operationID) || zeroInstallationID(tenantID) ||
		zeroInstallationID(userID) || zeroInstallationID(protectedBindingID) {
		return ""
	}
	ids := installationIDs(operationID)
	hash := sha256.New()
	for _, value := range [][16]byte{tenantID, userID, protectedBindingID,
		ids.role, ids.permission, ids.rolePermission} {
		text := installationUUID(value)
		_, _ = hash.Write([]byte{byte(len(text)), 0})
		_, _ = hash.Write([]byte(text))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func deriveInstallationID(operationID [16]byte, label string) [16]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy:iam:installation:v1"))
	_, _ = hash.Write(operationID[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(label))
	var result [16]byte
	copy(result[:], hash.Sum(nil))
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}

func validInstallationIdentity(input InstallationIdentityInput) bool {
	for _, value := range [][16]byte{input.OperationID, input.EnvironmentID,
		input.TenantID, input.FirstUserID, input.RootOrganizationID,
		input.ProtectedAdminBindingID} {
		if zeroInstallationID(value) {
			return false
		}
	}
	parsed, err := mail.ParseAddress(input.Email)
	return err == nil && parsed.Name == "" && parsed.Address == input.Email &&
		input.Email == strings.ToLower(input.Email) && len(input.Email) <= 254 &&
		input.DirectoryKeyID != "" && len(input.DirectoryKeyID) <= 64 &&
		strings.TrimSpace(input.DirectoryKeyID) == input.DirectoryKeyID &&
		len(input.SchemaName) >= 3 && len(input.SchemaName) <= 63 &&
		strings.HasPrefix(input.SchemaName, "t_") &&
		validInstallationText(input.DisplayName, 128) &&
		validInstallationText(input.TenantName, 128)
}

func validInstallationText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value
}

func zeroInstallationID(value [16]byte) bool { return value == [16]byte{} }

func installationUUID(value [16]byte) string {
	raw := make([]byte, 32)
	hex.Encode(raw, value[:])
	return string(raw[:8]) + "-" + string(raw[8:12]) + "-" +
		string(raw[12:16]) + "-" + string(raw[16:20]) + "-" + string(raw[20:])
}
