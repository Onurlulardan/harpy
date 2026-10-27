package main

import cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"

func cutoverObjects() []cutoverv1.AuthorityAdoptionObject {
	identities := cutoverObjectIdentities()
	objects := make([]cutoverv1.AuthorityAdoptionObject, len(identities))
	for index, identity := range identities {
		objects[index] = cutoverv1.AuthorityAdoptionObject{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Action: "adopt", IntroducedGeneration: 1,
			PreAdoptionOwner: "harpy_owner", AdoptedOwner: "harpy_cutover_authority_owner",
			PreDefinitionSHA256: localFixedDigest, AdoptedDefinitionSHA256: localFixedDigest,
			PreAdoptionACLSHA256: localFixedDigest, AdoptedACLSHA256: localFixedDigest,
		}
	}
	return objects
}

func cutoverDependencies() []cutoverv1.AuthorityCatalogDependency {
	identities := cutoverDependencyIdentities()
	dependencies := make([]cutoverv1.AuthorityCatalogDependency, len(identities))
	for index, identity := range identities {
		dependencies[index] = cutoverv1.AuthorityCatalogDependency{
			Kind: identity[0], SchemaName: identity[1], ObjectName: identity[2],
			IdentityArguments: identity[3], Owner: "harpy_owner",
			DefinitionSHA256: localFixedDigest, ACLSHA256: localFixedDigest,
		}
	}
	return dependencies
}

func cutoverReceiptObjects(
	objects []cutoverv1.AuthorityAdoptionObject,
) []cutoverv1.AuthorityAdoptionReceiptObject {
	receipts := make([]cutoverv1.AuthorityAdoptionReceiptObject, len(objects))
	for index, object := range objects {
		receipts[index] = cutoverv1.AuthorityAdoptionReceiptObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, IntroducedGeneration: object.IntroducedGeneration,
			ObjectOID: uint32(17000 + index), Owner: object.AdoptedOwner,
			DefinitionSHA256: object.AdoptedDefinitionSHA256, ACLSHA256: object.AdoptedACLSHA256,
		}
	}
	return receipts
}

func cutoverReceiptDependencies(
	dependencies []cutoverv1.AuthorityCatalogDependency,
) []cutoverv1.AuthorityCatalogDependencyReceipt {
	receipts := make([]cutoverv1.AuthorityCatalogDependencyReceipt, len(dependencies))
	for index, dependency := range dependencies {
		receipts[index] = cutoverv1.AuthorityCatalogDependencyReceipt{
			AuthorityCatalogDependency: dependency, ObjectOID: uint32(18000 + index),
		}
	}
	return receipts
}

func cutoverOwnerRole(objectCount int) cutoverv1.AuthorityOwnerRoleState {
	return cutoverv1.AuthorityOwnerRoleState{
		ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
		OwnedObjectCount: objectCount,
	}
}

func cutoverMigratorAtRest() cutoverv1.AuthorityQuiescedRoleState {
	return cutoverv1.AuthorityQuiescedRoleState{
		Login: true, ConnectionLimit: -1, PasswordState: "scram_sha_256",
		ValidUntil: "infinity", MembershipCount: 1,
	}
}
