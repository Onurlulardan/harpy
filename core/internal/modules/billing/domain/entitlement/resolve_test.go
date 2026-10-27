package entitlement

import "testing"

func TestResolveUsesBaseSubscriptionOverridePrecedence(t *testing.T) {
	booleanKey, _ := NewKey("features.ai")
	integerKey, _ := NewKey("iam.active_users")
	booleanDefinition, _ := NewDefinition(
		booleanKey, Boolean, NoMeter, Never, NoEnforcement, false,
	)
	integerDefinition, _ := NewDefinition(
		integerKey, Integer, Counter, CalendarMonthUTC, Hard, true,
	)
	ten, _ := IntegerValue(10)
	twenty, _ := IntegerValue(20)
	definitions := map[Key]Definition{
		booleanKey: booleanDefinition, integerKey: integerDefinition,
	}
	base := map[Key]Value{booleanKey: BooleanValue(false), integerKey: ten}
	subscription := map[Key]Value{booleanKey: BooleanValue(true), integerKey: twenty}
	overrides := map[Key]Value{integerKey: UnlimitedValue()}
	resolved, err := Resolve(definitions, base, subscription, overrides)
	if err != nil {
		t.Fatal(err)
	}
	if value, ok := resolved[booleanKey].Boolean(); !ok || !value {
		t.Fatalf("resolved boolean = %t, %t", value, ok)
	}
	if !resolved[integerKey].Unlimited() {
		t.Fatal("override did not win")
	}
	resolved[booleanKey] = BooleanValue(false)
	if value, _ := subscription[booleanKey].Boolean(); !value {
		t.Fatal("result mutation changed source layer")
	}
}

func TestResolveFailsClosedOnMissingOrMismatchedDefinition(t *testing.T) {
	key, _ := NewKey("iam.active_users")
	definition, _ := NewDefinition(key, Integer, Counter, Never, Hard, false)
	limit, _ := IntegerValue(10)
	tests := []struct {
		name        string
		definitions map[Key]Definition
		base        map[Key]Value
	}{
		{"nil definitions", nil, map[Key]Value{}},
		{"nil base", map[Key]Definition{key: definition}, nil},
		{"missing definition", map[Key]Definition{}, map[Key]Value{key: limit}},
		{"type mismatch", map[Key]Definition{key: definition},
			map[Key]Value{key: BooleanValue(true)}},
		{"forbidden unlimited", map[Key]Definition{key: definition},
			map[Key]Value{key: UnlimitedValue()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Resolve(test.definitions, test.base, nil, nil); err == nil {
				t.Fatal("invalid resolution was accepted")
			}
		})
	}
}
