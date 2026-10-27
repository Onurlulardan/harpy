package systemresourcelayouts

import (
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

type Resource = publicresource.Resource
type ID = publicresource.ID
type ActorEvidence = publicresource.ActorEvidence
type Admission = publicresource.Admission
type LayoutAccessRequest = publicresource.LayoutAccessRequest
type LayoutAccessDecision = publicresource.LayoutAccessDecision

const (
	ResourceUsers         = publicresource.Users
	ResourceRoles         = publicresource.Roles
	ResourceOrganizations = publicresource.Organizations
)

var (
	ErrLayoutAccessAuthentication = publicresource.ErrLayoutAccessAuthentication
	ErrLayoutAccessForbidden      = publicresource.ErrLayoutAccessForbidden
	ErrLayoutAccessUnavailable    = publicresource.ErrLayoutAccessUnavailable
	NewLayoutAccessDecision       = publicresource.NewLayoutAccessDecision
	ParseID                       = publicresource.ParseID
	NewActorEvidence              = publicresource.NewActorEvidence
	NewAdmission                  = publicresource.NewAdmission
)

func ParseResource(value string) (Resource, bool) {
	return publicresource.ParseResource(value)
}

func ParseRevisionID(value string) (string, bool) {
	parsed, err := uuid.Parse(value)
	return value, err == nil && parsed != uuid.Nil && parsed.Version() == 7 &&
		parsed.Variant() == uuid.RFC4122 && parsed.String() == value
}

type Definition = layoutapi.Definition
type DefinitionContract = layoutapi.DefinitionContract
type SurfaceDefinition = layoutapi.SurfaceDefinition
type SurfaceKind = layoutapi.SurfaceKind
type Section = layoutapi.Section
type Translation = layoutapi.Translation
type Revision = layoutapi.Revision
type RevisionState = layoutapi.RevisionState
type LayoutResourceKind = layoutapi.ResourceKind
type UpdateInput = layoutapi.UpdateInput
type DeleteInput = layoutapi.DeleteInput

const (
	DefinitionContractV1 = layoutapi.DefinitionContractV1
	DefinitionContractV2 = layoutapi.DefinitionContractV2
	SurfaceRecordCreate  = layoutapi.SurfaceRecordCreate
	SurfaceRecordUpdate  = layoutapi.SurfaceRecordUpdate
	SurfaceRecordDetail  = layoutapi.SurfaceRecordDetail
	SurfaceRecordList    = layoutapi.SurfaceRecordList
	SurfaceRecordFilter  = layoutapi.SurfaceRecordFilter
	SurfacePrint         = layoutapi.SurfacePrint
	RevisionDraft        = layoutapi.RevisionDraft
	RevisionPublished    = layoutapi.RevisionPublished
	LayoutResourceUser   = layoutapi.ResourceUser
)

var (
	ErrLayoutInvalid     = layoutapi.ErrInvalid
	ErrLayoutNotFound    = layoutapi.ErrNotFound
	ErrLayoutConflict    = layoutapi.ErrConflict
	ErrLayoutUnavailable = layoutapi.ErrUnavailable
)

type PresentationDocument = presentation.Document
type PresentationSurface = presentation.Surface

const PresentationRegistryContract = presentation.RegistryContractVersion
