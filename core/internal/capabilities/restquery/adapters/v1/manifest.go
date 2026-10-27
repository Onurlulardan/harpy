package v1query

import (
	"fmt"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type VerifiedManifestSource interface {
	VerifiedV1QueryManifest() (ManifestSnapshot, error)
}

type FilterShape uint8

const (
	FilterShapeScalar FilterShape = iota + 1
	FilterShapeCollection
)

type FilterFieldBinding struct {
	Name    string
	Shape   FilterShape
	AliasID string
}

type SortFieldBinding struct {
	Name    string
	AliasID string
}

type OperatorBinding struct {
	Operator string
	AliasID  string
}

type QueryKeyBinding struct {
	Key     string
	AliasID string
}

type ManifestSnapshot struct {
	ManifestID           string
	OwnerModule          string
	Domain               string
	DefaultLimit         int
	MaximumPageSize      int
	DefaultSortField     string
	DefaultSortDirection string
	FilterFields         map[string]FilterFieldBinding
	SortFields           map[string]SortFieldBinding
	OperatorAliases      map[string]OperatorBinding
	QueryKeyAliases      map[string]QueryKeyBinding
}

type manifest struct {
	id              string
	identity        contractretirement.Identity
	defaultLimit    int
	maximumPageSize int
	defaultSort     api.RequestSort
	filterFields    map[string]fieldBinding
	sortFields      map[string]fieldBinding
	operatorAliases map[string]operatorBinding
	queryKeyAliases map[string]queryKeyBinding
	validated       bool
}

type fieldBinding struct {
	name    string
	shape   FilterShape
	aliasID string
}

type operatorBinding struct {
	operator api.Operator
	aliasID  string
}

type queryKeyBinding struct {
	key     string
	aliasID string
}

type decoder struct{ manifest manifest }

func newDecoder(source VerifiedManifestSource) (decoder, error) {
	if nilManifestSource(source) {
		return decoder{}, fmt.Errorf("verified V1 query manifest source is required")
	}
	snapshot, err := source.VerifiedV1QueryManifest()
	if err != nil {
		return decoder{}, fmt.Errorf("load verified V1 query manifest: %w", err)
	}
	validated, err := newManifest(snapshot)
	if err != nil {
		return decoder{}, err
	}
	return decoder{manifest: validated}, nil
}

func nilManifestSource(source VerifiedManifestSource) bool {
	if source == nil {
		return true
	}
	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (decoder decoder) Validate() error {
	return decoder.manifest.Validate()
}

func (decoder) String() string                         { return v1Redaction }
func (decoder) GoString() string                       { return v1Redaction }
func (decoder decoder) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (decoder) MarshalJSON() ([]byte, error)           { return rejectV1JSON() }
func (ManifestSnapshot) String() string                { return v1Redaction }
func (ManifestSnapshot) GoString() string              { return v1Redaction }
func (snapshot ManifestSnapshot) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (ManifestSnapshot) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
func (FilterFieldBinding) String() string             { return v1Redaction }
func (FilterFieldBinding) GoString() string           { return v1Redaction }
func (binding FilterFieldBinding) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (FilterFieldBinding) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
func (SortFieldBinding) String() string                 { return v1Redaction }
func (SortFieldBinding) GoString() string               { return v1Redaction }
func (binding SortFieldBinding) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (SortFieldBinding) MarshalJSON() ([]byte, error)          { return rejectV1JSON() }
func (OperatorBinding) String() string                         { return v1Redaction }
func (OperatorBinding) GoString() string                       { return v1Redaction }
func (binding OperatorBinding) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (OperatorBinding) MarshalJSON() ([]byte, error)           { return rejectV1JSON() }
func (QueryKeyBinding) String() string                         { return v1Redaction }
func (QueryKeyBinding) GoString() string                       { return v1Redaction }
func (binding QueryKeyBinding) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (QueryKeyBinding) MarshalJSON() ([]byte, error)           { return rejectV1JSON() }
