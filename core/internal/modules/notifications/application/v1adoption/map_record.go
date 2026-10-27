package v1adoption

import (
	"crypto/sha256"
	"strings"

	"github.com/google/uuid"
)

var (
	deliveryNamespace = uuid.MustParse("4ba35b13-07e7-4dd6-b1f7-21361658c655")
	attemptNamespace  = uuid.MustParse("ac526adc-e48c-42da-9ba8-b0b2873391bc")
)

func mapRecord(plan Plan, source SourceRecord) (MappedRecord, string, bool) {
	canonical := sourceDigest(source)
	if source.Channel == "websocket" {
		return MappedRecord{CanonicalDigest: canonical}, "", true
	}
	if source.ID == uuid.Nil || source.TenantID != plan.TenantID || source.UserID == nil ||
		*source.UserID == uuid.Nil || source.CreatedAt == nil || source.CreatedAt.IsZero() ||
		source.InboxRevision < 1 {
		return MappedRecord{CanonicalDigest: canonical}, "V1_IDENTITY_INVALID", false
	}
	notificationType := strings.TrimSpace(source.NotificationType)
	subject, body := strings.TrimSpace(source.Title), strings.TrimSpace(source.Body)
	if notificationType == "" || len(notificationType) > 96 || subject == "" ||
		len(subject) > 512 || body == "" || len(body) > 64*1024 {
		return MappedRecord{CanonicalDigest: canonical}, "V1_CONTENT_INVALID", false
	}
	channel := ""
	if source.Channel == "email" || source.Channel == "mail" {
		channel = "email"
	} else if source.Channel != "flow" {
		return MappedRecord{CanonicalDigest: canonical}, "V1_CHANNEL_UNKNOWN", false
	}
	createdAt := source.CreatedAt.UTC()
	var readAt = source.ReadAt
	if source.Read {
		if readAt == nil || readAt.IsZero() {
			value := createdAt
			readAt = &value
		} else {
			value := readAt.UTC()
			readAt = &value
		}
	} else {
		readAt = nil
	}
	effectKey := sha256.Sum256([]byte("notifications.v1.effect\x00" + source.ID.String()))
	provider := sha256.Sum256([]byte("notifications.v1.provider\x00" + source.ID.String()))
	mapped := MappedRecord{MessageID: source.ID, RecipientID: *source.UserID,
		NotificationType: notificationType, Category: legacyCategory(notificationType, source.Channel),
		Subject: subject, Body: body, Channel: channel, InboxRevision: source.InboxRevision,
		CreatedAt: createdAt, ReadAt: readAt, CanonicalDigest: canonical,
		EffectKeyHash: effectKey, SubjectDigest: sha256.Sum256([]byte(subject)),
		BodyDigest: sha256.Sum256([]byte(body)), ProviderReferenceDigest: provider}
	if channel != "" {
		mapped.DeliveryID = uuid.NewSHA1(deliveryNamespace, []byte(source.ID.String()+"\x00"+channel))
		mapped.AttemptID = uuid.NewSHA1(attemptNamespace, []byte(mapped.DeliveryID.String()+"\x001"))
	}
	return mapped, "", false
}

func legacyCategory(notificationType, channel string) string {
	value := strings.ToLower(notificationType)
	switch {
	case channel == "flow" || strings.Contains(value, "workflow"):
		return "workflow_action"
	case strings.Contains(value, "security") || strings.Contains(value, "password") ||
		strings.Contains(value, "auth") || strings.Contains(value, "login") ||
		strings.Contains(value, "verification"):
		return "security"
	case strings.Contains(value, "billing") || strings.Contains(value, "invoice"):
		return "billing"
	case strings.Contains(value, "document") || strings.Contains(value, "export"):
		return "document"
	case strings.Contains(value, "product") || strings.Contains(value, "marketing"):
		return "product"
	default:
		return "system"
	}
}
