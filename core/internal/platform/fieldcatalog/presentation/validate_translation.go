package presentation

import (
	"regexp"
	"unicode"
	"unicode/utf8"
)

var localePattern = regexp.MustCompile(
	`^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?$`,
)

func validLocale(value string) bool {
	return localePattern.MatchString(value)
}

func (state *validationState) validateTranslations(
	path string,
	translations []Translation,
	requireDefaultLabel bool,
) error {
	if translations == nil || len(translations) > maximumTranslations {
		return invalid(path, "translation budget")
	}
	defaultLabel := false
	previous := ""
	for _, translation := range translations {
		if _, declared := state.locales[translation.Locale]; !declared ||
			previous != "" && previous >= translation.Locale ||
			!validText(translation.Label, maximumLabelRunes) ||
			!validText(translation.Help, maximumHelpRunes) ||
			!validText(translation.Placeholder, maximumPlaceholderRunes) {
			return invalid(path, "locale order or text")
		}
		if translation.Locale == state.document.DefaultLocale && translation.Label != "" {
			defaultLabel = true
		}
		previous = translation.Locale
	}
	if requireDefaultLabel && !defaultLabel {
		return invalid(path, "default locale label")
	}
	return nil
}

func validText(value string, maximum int) bool {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > maximum {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
