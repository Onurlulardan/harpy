package recordkernel

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

type TextStorageCompanionV1 struct {
	variant string
	profile string
	value   string
	null    bool
}

func MaterializeTextStorageCompanionV1(
	value Value,
	variant string,
	profile string,
) (TextStorageCompanionV1, error) {
	wantProfile := map[string]string{
		"folded": "unicode_default_casefold_v1",
		"accent": "accent_insensitive_v1",
	}[variant]
	if value.TypeID() != TypeText || wantProfile == "" || profile != wantProfile {
		return TextStorageCompanionV1{}, ErrInputInvalid
	}
	if value.IsNull() {
		if !bytes.Equal(value.CanonicalJSON(), []byte("null")) {
			return TextStorageCompanionV1{}, ErrInputInvalid
		}
		return TextStorageCompanionV1{variant: variant, profile: profile, null: true}, nil
	}
	var text string
	if json.Unmarshal(value.CanonicalJSON(), &text) != nil {
		return TextStorageCompanionV1{}, ErrInputInvalid
	}
	if variant == "folded" {
		text = textprofile.UnicodeDefaultCasefoldV1(text)
	} else {
		text = textprofile.AccentInsensitiveV1(text)
	}
	return TextStorageCompanionV1{variant: variant, profile: profile, value: text}, nil
}

func (companion TextStorageCompanionV1) DatabaseValue() (string, bool) {
	if companion.variant == "" || companion.profile == "" || companion.null {
		return "", false
	}
	return companion.value, true
}
