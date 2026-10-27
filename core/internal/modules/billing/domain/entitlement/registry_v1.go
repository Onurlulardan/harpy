package entitlement

type Unit string

const (
	UnitBoolean Unit = "boolean"
	UnitCount   Unit = "count"
	UnitByte    Unit = "byte"
)

type RegistryEntry struct {
	Definition Definition
	Unit       Unit
	Owner      string
}

func RegistryV1() []RegistryEntry {
	return []RegistryEntry{
		integerEntry("forms.submissions", Counter, CalendarMonthUTC, "forms", UnitCount),
		integerEntry("iam.active_users", Gauge, Never, "iam", UnitCount),
		integerEntry("forms.active_forms", Gauge, Never, "forms", UnitCount),
		integerEntry("documents.storage_bytes", Gauge, Never, "documents", UnitByte),
		booleanEntry("workflow.enabled", "workflow"),
		booleanEntry("api.access", "platform"),
	}
}

func RegistryV1ByKey() map[Key]RegistryEntry {
	entries := RegistryV1()
	result := make(map[Key]RegistryEntry, len(entries))
	for _, entry := range entries {
		result[entry.Definition.Key()] = entry
	}
	return result
}

func integerEntry(
	keyValue string,
	meter MeterKind,
	reset ResetPolicy,
	owner string,
	unit Unit,
) RegistryEntry {
	key, _ := NewKey(keyValue)
	definition, _ := NewDefinition(key, Integer, meter, reset, Hard, true)
	return RegistryEntry{Definition: definition, Unit: unit, Owner: owner}
}

func booleanEntry(keyValue string, owner string) RegistryEntry {
	key, _ := NewKey(keyValue)
	definition, _ := NewDefinition(
		key, Boolean, NoMeter, Never, NoEnforcement, false,
	)
	return RegistryEntry{Definition: definition, Unit: UnitBoolean, Owner: owner}
}
