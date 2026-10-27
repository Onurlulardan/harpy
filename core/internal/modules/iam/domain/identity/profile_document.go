package identity

import (
	"bytes"
)

const MaximumProfileBytes = 1 << 20

type ProfileResource string

const (
	ProfileUser         ProfileResource = "user"
	ProfileOrganization ProfileResource = "organization"
	ProfileRole         ProfileResource = "role"
)

func (resource ProfileResource) Valid() bool {
	return resource == ProfileUser || resource == ProfileOrganization ||
		resource == ProfileRole
}

type ProfileDocument struct {
	resource ProfileResource
	schemaID IdentityID
	values   []byte
	digest   [32]byte
}

func NewProfileDocument(
	resource ProfileResource, schemaID IdentityID, values []byte, digest [32]byte,
) (ProfileDocument, error) {
	if !resource.Valid() || schemaID == (IdentityID{}) || len(values) < 2 ||
		len(values) > MaximumProfileBytes || digest == [32]byte{} {
		return ProfileDocument{}, ErrInvalidRecord
	}
	return ProfileDocument{
		resource: resource, schemaID: schemaID,
		values: bytes.Clone(values), digest: digest,
	}, nil
}

func (document ProfileDocument) Resource() ProfileResource { return document.resource }
func (document ProfileDocument) SchemaID() IdentityID      { return document.schemaID }
func (document ProfileDocument) Values() []byte            { return bytes.Clone(document.values) }
func (document ProfileDocument) SHA256() [32]byte          { return document.digest }

func (document ProfileDocument) Validate(resource ProfileResource) error {
	if document.resource != resource {
		return ErrInvalidRecord
	}
	_, err := NewProfileDocument(
		document.resource, document.schemaID, document.values, document.digest,
	)
	return err
}
