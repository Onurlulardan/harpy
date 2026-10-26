package cutoverv1

type authorityNamespaceObject struct {
	Kind              string `json:"kind"`
	SchemaName        string `json:"schema_name"`
	ObjectName        string `json:"object_name"`
	IdentityArguments string `json:"identity_arguments"`
	Owner             string `json:"owner"`
	DefinitionSHA256  string `json:"definition_sha256"`
	ACLSHA256         string `json:"acl_sha256"`
}

func AuthorityOperationPreNamespaceSHA256(objects []AuthorityAdoptionObject) string {
	states := make([]authorityNamespaceObject, len(objects))
	for index, object := range objects {
		states[index] = authorityNamespaceObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, Owner: object.PreAdoptionOwner,
			DefinitionSHA256: object.PreDefinitionSHA256, ACLSHA256: object.PreAdoptionACLSHA256,
		}
	}
	return authorityCatalogDigest(states)
}

func AuthorityOperationPostNamespaceSHA256(objects []AuthorityAdoptionObject) string {
	states := make([]authorityNamespaceObject, len(objects))
	for index, object := range objects {
		states[index] = authorityNamespaceObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, Owner: object.AdoptedOwner,
			DefinitionSHA256: object.AdoptedDefinitionSHA256, ACLSHA256: object.AdoptedACLSHA256,
		}
	}
	return authorityCatalogDigest(states)
}

func AuthorityReceiptNamespaceSHA256(objects []AuthorityAdoptionReceiptObject) string {
	states := make([]authorityNamespaceObject, len(objects))
	for index, object := range objects {
		states[index] = authorityNamespaceObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, Owner: object.Owner,
			DefinitionSHA256: object.DefinitionSHA256, ACLSHA256: object.ACLSHA256,
		}
	}
	return authorityCatalogDigest(states)
}
