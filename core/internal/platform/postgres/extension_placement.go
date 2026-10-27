package postgres

import (
	"fmt"
	"strings"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func validateExtensionPlacement(
	origin runtimev1.DatabaseOrigin,
	schema string,
	owner string,
) error {
	wantSchema := "platform_ext"
	if origin == runtimev1.DatabaseOriginV1Adopted {
		wantSchema = "public"
	}
	invalidFreshOwner := origin == runtimev1.DatabaseOriginFresh &&
		owner != "harpy_extension_admin"
	invalidAdoptedOwner := origin == runtimev1.DatabaseOriginV1Adopted &&
		strings.HasPrefix(owner, "harpy_") && owner != "harpy_extension_admin" &&
		owner != "harpy_legacy_extension_owner"
	if schema != wantSchema || !schemaNamePattern.MatchString(owner) ||
		invalidFreshOwner || invalidAdoptedOwner {
		return fmt.Errorf("%w: extension placement", ErrStartupPreflight)
	}
	return nil
}
