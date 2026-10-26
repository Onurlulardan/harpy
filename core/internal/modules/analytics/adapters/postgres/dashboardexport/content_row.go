package dashboardexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type contentRow struct {
	tenantID, actorID, exportID, dashboardID, versionID                    uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock                 int64
	manifest, pinSHA256                                                    []byte
	format                                                                 string
	parameters, widgets                                                    []byte
	maximumRows, maximumBytes, generation                                  int64
	expiresAt                                                              time.Time
	objectKey, objectVersion, keyID                                        string
	wrappedDEK, wrapNonce, headerSHA256, ciphertextSHA256, plaintextSHA256 []byte
	ciphertextBytes, plaintextBytes, rows                                  int64
	contentType, filename                                                  string
	readyAt                                                                time.Time
}

func (stored *contentRow) destinations() []any {
	return []any{&stored.actorID, &stored.exportID, &stored.dashboardID, &stored.versionID,
		&stored.tokenVersion, &stored.authorizationVersion, &stored.authorizationClock,
		&stored.manifest, &stored.pinSHA256, &stored.format, &stored.parameters,
		&stored.widgets, &stored.maximumRows, &stored.maximumBytes, &stored.expiresAt,
		&stored.generation, &stored.objectKey, &stored.objectVersion, &stored.keyID,
		&stored.wrappedDEK, &stored.wrapNonce, &stored.headerSHA256,
		&stored.ciphertextSHA256, &stored.plaintextSHA256, &stored.ciphertextBytes,
		&stored.plaintextBytes, &stored.rows, &stored.contentType, &stored.filename,
		&stored.readyAt}
}

func (stored contentRow) decode() (exportapp.StoredContent, error) {
	pin, err := stored.decodePin()
	if err != nil || stored.objectVersion == "" || stored.readyAt.IsZero() {
		return exportapp.StoredContent{}, exportapp.ErrUnavailable
	}
	draft, err := stored.decodeDraft(pin)
	if err != nil {
		return exportapp.StoredContent{}, err
	}
	return exportapp.StoredContent{Pin: pin, Draft: draft,
		ObjectVersion: stored.objectVersion, ReadyAt: stored.readyAt.UTC()}, nil
}

func (stored contentRow) decodePin() (exportapp.Pin, error) {
	items, err := decodePinnedWidgets(stored.widgets)
	if err != nil || len(stored.manifest) != sha256.Size || len(stored.pinSHA256) != sha256.Size ||
		!json.Valid(stored.parameters) || stored.tenantID == uuid.Nil || stored.actorID == uuid.Nil ||
		stored.exportID.Version() != 7 || stored.dashboardID.Version() != 7 ||
		stored.versionID.Version() != 7 || !exportapp.Format(stored.format).Valid() ||
		stored.tokenVersion < 1 || stored.authorizationVersion < 1 ||
		stored.authorizationClock < 1 || stored.maximumRows < 1 ||
		stored.maximumRows > exportapp.MaximumRows || stored.maximumBytes < 1 ||
		stored.maximumBytes > exportapp.MaximumPlaintextBytes || stored.expiresAt.IsZero() {
		return exportapp.Pin{}, exportapp.ErrUnavailable
	}
	var manifest, pin [sha256.Size]byte
	copy(manifest[:], stored.manifest)
	copy(pin[:], stored.pinSHA256)
	return exportapp.Pin{TenantID: stored.tenantID, ActorID: stored.actorID,
		ExportID: stored.exportID, DashboardID: stored.dashboardID,
		DashboardVersionID: stored.versionID, TokenVersion: stored.tokenVersion,
		AuthorizationVersion: stored.authorizationVersion,
		AuthorizationClock:   stored.authorizationClock, DashboardManifestSHA256: manifest,
		PinSHA256: pin, Format: exportapp.Format(stored.format),
		Parameters: append([]byte(nil), stored.parameters...), Widgets: items,
		MaximumRows: stored.maximumRows, MaximumPlaintextBytes: stored.maximumBytes,
		ExpiresAt: stored.expiresAt.UTC()}, nil
}

func (stored contentRow) decodeDraft(pin exportapp.Pin) (exportapp.ArtifactDraft, error) {
	if len(stored.headerSHA256) != sha256.Size || len(stored.ciphertextSHA256) != sha256.Size ||
		len(stored.plaintextSHA256) != sha256.Size || len(stored.wrapNonce) != fcartv1.NonceBytes ||
		stored.generation < 1 || stored.generation > exportapp.MaximumAttempts ||
		stored.objectKey == "" || stored.keyID == "" || len(stored.wrappedDEK) != 48 ||
		stored.ciphertextBytes < 35 || stored.plaintextBytes < 1 ||
		stored.plaintextBytes > pin.MaximumPlaintextBytes || stored.rows < 0 ||
		stored.rows > pin.MaximumRows || stored.contentType != pin.Format.ContentType() ||
		stored.filename == "" {
		return exportapp.ArtifactDraft{}, exportapp.ErrUnavailable
	}
	var header, ciphertext, plaintext [sha256.Size]byte
	copy(header[:], stored.headerSHA256)
	copy(ciphertext[:], stored.ciphertextSHA256)
	copy(plaintext[:], stored.plaintextSHA256)
	var nonce [fcartv1.NonceBytes]byte
	copy(nonce[:], stored.wrapNonce)
	envelope := fcartv1.KeyEnvelope{KEKID: stored.keyID,
		Ciphertext: append([]byte(nil), stored.wrappedDEK...), Nonce: nonce}
	return exportapp.ArtifactDraft{AttemptGeneration: stored.generation,
		ObjectKey: stored.objectKey, Envelope: envelope, HeaderSHA256: header,
		CiphertextSHA256: ciphertext, PlaintextSHA256: plaintext,
		CiphertextBytes: stored.ciphertextBytes, PlaintextBytes: stored.plaintextBytes,
		Rows: stored.rows, ContentType: stored.contentType, Filename: stored.filename}, nil
}

func decodePinnedWidgets(document []byte) ([]exportapp.PinnedWidget, error) {
	var wire []struct {
		WidgetID       string `json:"widget_id"`
		Position       uint16 `json:"position"`
		QueryVersionID string `json:"query_version_id"`
		Parameters     string `json:"parameters"`
		Output         string `json:"output"`
	}
	if json.Unmarshal(document, &wire) != nil || len(wire) < 1 || len(wire) > 30 {
		return nil, exportapp.ErrUnavailable
	}
	result := make([]exportapp.PinnedWidget, len(wire))
	for index, item := range wire {
		widgetID, widgetErr := uuid.Parse(item.WidgetID)
		queryID, queryErr := uuid.Parse(item.QueryVersionID)
		parameters, parametersErr := hex.DecodeString(item.Parameters)
		output, outputErr := hex.DecodeString(item.Output)
		if widgetErr != nil || queryErr != nil || parametersErr != nil || outputErr != nil ||
			widgetID.Version() != 7 || queryID.Version() != 7 || int(item.Position) != index ||
			!json.Valid(parameters) || !json.Valid(output) {
			return nil, exportapp.ErrUnavailable
		}
		result[index] = exportapp.PinnedWidget{WidgetID: widgetID, QueryVersionID: queryID,
			Position: item.Position, Parameters: parameters, Output: output}
	}
	return result, nil
}
