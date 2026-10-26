package licenseauthority

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	keyIDPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	digestPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	channelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
)

func strictDecode(document []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func parseExactUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func parseExactInstant(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.Nanosecond() != 0 || parsed.Location() != time.UTC ||
		parsed.Format(time.RFC3339) != value {
		return time.Time{}, false
	}
	return parsed, true
}

func validText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "\x00\r\n")
}

func validDelivery(value DeliveryKind) bool {
	return value == DeliveryConnected || value == DeliveryAirGap || value == DeliveryDevelopment
}

func validStatus(value Status) bool {
	return value == StatusActive || value == StatusSuspended ||
		value == StatusRevoked || value == StatusExpired
}
