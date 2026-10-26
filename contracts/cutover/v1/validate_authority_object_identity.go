package cutoverv1

import (
	"fmt"
	"regexp"
	"strings"
)

var authorityIdentityArgumentsPattern = regexp.MustCompile(
	`^(?:[a-z][a-z0-9_]*(?:,[a-z][a-z0-9_]*)*)?$`,
)

func validAuthorityObjectIdentity(
	object AuthorityAdoptionObject, index int, seen map[string]struct{},
) bool {
	if !validAuthorityIdentityFields(
		object.Kind, object.SchemaName, object.ObjectName, object.IdentityArguments,
	) {
		return false
	}
	if index < len(authorityObjectOrder) && !matchesBaseAuthorityObject(object, index) {
		return false
	}
	if index >= len(authorityObjectOrder) &&
		(object.Kind == "schema" || object.SchemaName != "cutover_authority" ||
			(object.Kind == "function" && object.IdentityArguments == "") ||
			!strings.HasSuffix(object.ObjectName, fmt.Sprintf("_v%d", object.IntroducedGeneration))) {
		return false
	}
	key := authorityObjectIdentity(
		object.Kind, object.SchemaName, object.ObjectName, object.IdentityArguments,
	)
	if _, exists := seen[key]; exists {
		return false
	}
	seen[key] = struct{}{}
	return true
}

func validAuthorityReceiptIdentity(
	object AuthorityAdoptionReceiptObject, index int,
) bool {
	if !validAuthorityIdentityFields(
		object.Kind, object.SchemaName, object.ObjectName, object.IdentityArguments,
	) {
		return false
	}
	if index < len(authorityObjectOrder) {
		expected := authorityObjectOrder[index]
		return object.Kind == expected[0] && object.SchemaName == expected[1] &&
			object.ObjectName == expected[2] && object.IdentityArguments == expected[3]
	}
	return object.Kind != "schema" && object.SchemaName == "cutover_authority" &&
		(object.Kind != "function" || object.IdentityArguments != "") &&
		strings.HasSuffix(object.ObjectName, fmt.Sprintf("_v%d", object.IntroducedGeneration))
}

func validAuthorityIdentityFields(kind, schema, name, arguments string) bool {
	return (kind == "schema" || kind == "table" || kind == "function") &&
		sqlNamePattern.MatchString(schema) && sqlNamePattern.MatchString(name) &&
		len(arguments) <= 1024 && authorityIdentityArgumentsPattern.MatchString(arguments) &&
		(kind == "function" || arguments == "")
}

func authorityObjectIdentity(kind, schema, name, arguments string) string {
	return kind + ":" + schema + ":" + name + ":" + arguments
}
