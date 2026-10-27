package application

import (
	"sort"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func presentationLocales(definition layoutapi.Definition) []string {
	locales := map[string]struct{}{"en": {}}
	for _, surface := range definition.Surfaces {
		collectLocales(locales, surface.Translations)
		for _, section := range surface.Sections {
			collectLocales(locales, section.Translations)
			for _, field := range section.Fields {
				collectLocales(locales, field.Translations)
			}
		}
	}
	result := make([]string, 0, len(locales))
	for locale := range locales {
		result = append(result, locale)
	}
	sort.Strings(result)
	return result
}

func collectLocales(target map[string]struct{}, values []layoutapi.Translation) {
	for _, translation := range values {
		target[translation.Locale] = struct{}{}
	}
}

func presentationTranslations(
	input []layoutapi.Translation,
	defaultLabel string,
) []presentation.Translation {
	result := make([]presentation.Translation, 0, len(input)+1)
	hasEnglish := false
	for _, translation := range input {
		result = append(result, presentation.Translation{
			Locale: translation.Locale, Label: translation.Label, Help: translation.Help,
		})
		hasEnglish = hasEnglish || translation.Locale == "en"
	}
	if !hasEnglish {
		result = append(result, presentation.Translation{Locale: "en", Label: defaultLabel})
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Locale < result[right].Locale
	})
	return result
}
