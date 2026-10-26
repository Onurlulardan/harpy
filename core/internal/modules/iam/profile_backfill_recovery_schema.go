package iam

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func profileRecoverySchemasMatch(
	parent profilebackfillapp.RecoveryParent,
	source profileapp.Schema,
	target profileapp.Schema,
) bool {
	if source.ID() != parent.SourceSchemaID || target.ID() != parent.Target.ID ||
		source.ID() == target.ID() || source.Resource() != target.Resource() ||
		string(target.Resource()) != parent.Target.ResourceKind ||
		source.State() != profileapp.SchemaActive ||
		target.State() != profileapp.SchemaActive ||
		target.Version() != source.Version()+1 ||
		target.Version() != parent.Target.Version ||
		target.DefinitionSHA256() != parent.Target.DefinitionSHA256 {
		return false
	}
	fields := target.Fields()
	if len(fields) != len(parent.Target.Fields) {
		return false
	}
	for _, expected := range parent.Target.Fields {
		field, found := target.Field(expected.ID)
		if !found || field.Required() != expected.Required ||
			(field.State() == profileapp.FieldActive) != expected.Active ||
			field.SemanticSHA256() != expected.SemanticSHA256 {
			return false
		}
	}
	return true
}

func validProfileRecoverySource(
	active profileapp.Schema,
	actual profileapp.Schema,
	expectedID uuid.UUID,
) bool {
	if actual.ID() != expectedID || actual.ID() == uuid.Nil ||
		actual.Resource() != active.Resource() || actual.Version() > active.Version() {
		return false
	}
	if actual.State() == profileapp.SchemaActive {
		return actual.ID() == active.ID() &&
			actual.DefinitionSHA256() == active.DefinitionSHA256()
	}
	return actual.State() == profileapp.SchemaRetired
}
