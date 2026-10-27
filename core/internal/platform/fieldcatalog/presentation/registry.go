package presentation

import (
	"reflect"
	"slices"
)

type WidgetDescriptor struct {
	ID         string   `json:"id"`
	FieldTypes []string `json:"field_types"`
}

type Registry struct {
	Contract string             `json:"contract"`
	Widgets  []WidgetDescriptor `json:"widgets"`
}

type registryCatalog struct {
	descriptors map[string]WidgetDescriptor
}

func DefaultRegistry() Registry {
	widgets := defaultWidgetDescriptors()
	result := make([]WidgetDescriptor, len(widgets))
	for index, widget := range widgets {
		result[index] = WidgetDescriptor{
			ID: widget.ID, FieldTypes: append([]string(nil), widget.FieldTypes...),
		}
	}
	return Registry{Contract: RegistryContractVersion, Widgets: result}
}

func loadRegistry(source Registry) (registryCatalog, error) {
	if source.Contract != RegistryContractVersion || !reflect.DeepEqual(source, DefaultRegistry()) {
		return registryCatalog{}, invalid("$.registry", "contract")
	}
	byID := make(map[string]WidgetDescriptor, len(source.Widgets))
	previous := ""
	for _, descriptor := range source.Widgets {
		if descriptor.ID == "" || previous >= descriptor.ID ||
			len(descriptor.FieldTypes) == 0 || !strictStrings(descriptor.FieldTypes) {
			return registryCatalog{}, invalid("$.registry", "descriptor order")
		}
		previous = descriptor.ID
		byID[descriptor.ID] = descriptor
	}
	return registryCatalog{descriptors: byID}, nil
}

func (catalog registryCatalog) accepts(widget, fieldType string) bool {
	descriptor, exists := catalog.descriptors[widget]
	return exists && slices.Contains(descriptor.FieldTypes, fieldType)
}

func strictStrings(values []string) bool {
	for index, value := range values {
		if value == "" || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}
