package customerprovisioning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,98}[a-z0-9])?$`)

type FirstAdmin struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type CustomerRequest struct {
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	FirstAdmin FirstAdmin `json:"first_admin"`
}

func (request CustomerRequest) valid() bool {
	return validText(request.Name, 128) && slugPattern.MatchString(request.Slug) &&
		validText(request.FirstAdmin.DisplayName, 128) &&
		validEmail(request.FirstAdmin.Email)
}

func (request CustomerRequest) digest() (string, error) {
	if !request.valid() {
		return "", ErrInvalid
	}
	canonical, err := json.Marshal(request)
	if err != nil {
		return "", ErrInvalid
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func validText(value string, maximum int) bool {
	if value == "" || len(value) > maximum || !utf8.ValidString(value) ||
		strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validEmail(value string) bool {
	if value == "" || len(value) > 254 || value != strings.ToLower(value) ||
		strings.TrimSpace(value) != value || strings.Count(value, "@") != 1 {
		return false
	}
	parts := strings.Split(value, "@")
	return len(parts[0]) > 0 && len(parts[0]) <= 64 &&
		len(parts[1]) > 2 && strings.Contains(parts[1], ".") &&
		!strings.ContainsAny(value, " <>\t\r\n")
}
