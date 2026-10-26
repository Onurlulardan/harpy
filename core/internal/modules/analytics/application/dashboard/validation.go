package dashboard

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"regexp"
	"sort"
	"time"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	"github.com/google/uuid"
)

var namePattern = regexp.MustCompile(`^[\pL\pN][\pL\pN _.-]{0,99}$`)
var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func validateSnapshot(value Snapshot) error {
	dashboardID, err := analyticsdashboard.NewID([16]byte(value.DashboardID))
	if err != nil {
		return ErrUnavailable
	}
	versionID, err := analyticsdashboard.NewVersionID([16]byte(value.VersionID))
	if err != nil || value.VersionNumber == 0 || value.RootRevision == 0 ||
		value.VersionRevision == 0 || !value.Lifecycle.Valid() || !value.Visibility.Valid() ||
		!namePattern.MatchString(value.Name) || !validTime(value.CreatedAt) ||
		!validTime(value.UpdatedAt) || value.UpdatedAt.Before(value.CreatedAt) {
		return ErrUnavailable
	}
	layout, err := NewDocument(value.LayoutDocument)
	if err != nil || layout.SHA256() != value.LayoutSHA256 || value.ManifestSHA256 == [32]byte{} {
		return ErrUnavailable
	}
	widgets, err := domainWidgets(value.Widgets)
	if err != nil {
		return err
	}
	version, err := analyticsdashboard.NewDraft(versionID, dashboardID, value.VersionNumber,
		value.VersionRevision, value.Visibility, widgets)
	if err != nil {
		return ErrUnavailable
	}
	if value.Published {
		if _, err := version.Publish(value.VersionRevision); err != nil {
			return ErrUnavailable
		}
	}
	return nil
}

func domainWidgets(values []Widget) ([]analyticsdashboard.Widget, error) {
	result := make([]analyticsdashboard.Widget, 0, len(values))
	for _, value := range values {
		id, err := analyticsdashboard.NewID([16]byte(value.WidgetID))
		parameters, parameterErr := NewDocument(value.ParameterDocument)
		output, outputErr := NewDocument(value.OutputDocument)
		if err != nil || parameterErr != nil || outputErr != nil {
			return nil, ErrInvalid
		}
		widget, err := analyticsdashboard.NewWidget(id, [16]byte(value.QueryVersionID),
			value.Position, parameters.SHA256(), output.SHA256())
		if err != nil {
			return nil, ErrInvalid
		}
		result = append(result, widget)
	}
	return result, nil
}

func digest(parts ...[]byte) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(part)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func widgetDigest(widgets []Widget) [32]byte {
	values := append([]Widget(nil), widgets...)
	sort.Slice(values, func(i, j int) bool { return values[i].Position < values[j].Position })
	encoded, _ := json.Marshal(values)
	return sha256.Sum256(encoded)
}

func digestDocument(value Document) []byte { digest := value.SHA256(); return digest[:] }
func digestArray(value [32]byte) []byte    { return value[:] }

func aclDigest(operation string, identity ScopeIdentity, dashboardID uuid.UUID,
	principalKind string, principalID uuid.UUID, level string, revision uint64,
) [32]byte {
	revisionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(revisionBytes, revision)
	return digest([]byte(operation), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), dashboardID[:], []byte(principalKind),
		principalID[:], []byte(level), revisionBytes)
}
