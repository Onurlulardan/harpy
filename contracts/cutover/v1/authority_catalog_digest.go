package cutoverv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type authorityCatalogContractObject struct {
	Kind                 string `json:"kind"`
	SchemaName           string `json:"schema_name"`
	ObjectName           string `json:"object_name"`
	IdentityArguments    string `json:"identity_arguments"`
	IntroducedGeneration int    `json:"introduced_generation"`
	Owner                string `json:"owner"`
	DefinitionSHA256     string `json:"definition_sha256"`
	ACLSHA256            string `json:"acl_sha256"`
}

type authorityCatalogInstanceObject struct {
	authorityCatalogContractObject
	ObjectOID uint32 `json:"object_oid"`
}

func AuthorityOperationCatalogContractSHA256(objects []AuthorityAdoptionObject) string {
	contract := make([]authorityCatalogContractObject, len(objects))
	for index, object := range objects {
		contract[index] = authorityCatalogContractObject{
			Kind: object.Kind, SchemaName: object.SchemaName,
			ObjectName: object.ObjectName, IdentityArguments: object.IdentityArguments,
			IntroducedGeneration: object.IntroducedGeneration, Owner: object.AdoptedOwner,
			DefinitionSHA256: object.AdoptedDefinitionSHA256,
			ACLSHA256:        object.AdoptedACLSHA256,
		}
	}
	return authorityCatalogDigest(contract)
}

func AuthorityReceiptCatalogContractSHA256(objects []AuthorityAdoptionReceiptObject) string {
	contract := make([]authorityCatalogContractObject, len(objects))
	for index, object := range objects {
		contract[index] = receiptContractObject(object)
	}
	return authorityCatalogDigest(contract)
}

func AuthorityReceiptCatalogInstanceSHA256(objects []AuthorityAdoptionReceiptObject) string {
	instances := make([]authorityCatalogInstanceObject, len(objects))
	for index, object := range objects {
		instances[index] = authorityCatalogInstanceObject{
			authorityCatalogContractObject: receiptContractObject(object),
			ObjectOID:                      object.ObjectOID,
		}
	}
	return authorityCatalogDigest(instances)
}

func receiptContractObject(object AuthorityAdoptionReceiptObject) authorityCatalogContractObject {
	return authorityCatalogContractObject{
		Kind: object.Kind, SchemaName: object.SchemaName,
		ObjectName: object.ObjectName, IdentityArguments: object.IdentityArguments,
		IntroducedGeneration: object.IntroducedGeneration, Owner: object.Owner,
		DefinitionSHA256: object.DefinitionSHA256, ACLSHA256: object.ACLSHA256,
	}
}

func authorityCatalogDigest(value any) string {
	if length, ok := authorityCatalogLength(value); ok && length == 0 {
		return emptyCatalogSHA256
	}
	document, err := json.Marshal(value)
	if err != nil {
		panic("authority catalog contains non-JSON data")
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

func authorityCatalogLength(value any) (int, bool) {
	switch typed := value.(type) {
	case []authorityCatalogContractObject:
		return len(typed), true
	case []authorityCatalogInstanceObject:
		return len(typed), true
	default:
		return 0, false
	}
}
