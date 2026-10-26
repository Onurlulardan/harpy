package entitlement

type Value struct {
	kind      ValueKind
	boolean   bool
	integer   int64
	unlimited bool
}

func BooleanValue(value bool) Value {
	return Value{kind: Boolean, boolean: value}
}

func IntegerValue(value int64) (Value, error) {
	if value < 0 {
		return Value{}, ErrValue
	}
	return Value{kind: Integer, integer: value}, nil
}

func UnlimitedValue() Value {
	return Value{kind: Integer, unlimited: true}
}

func (value Value) Kind() ValueKind { return value.kind }

func (value Value) Boolean() (bool, bool) {
	return value.boolean, value.kind == Boolean
}

func (value Value) Integer() (int64, bool) {
	return value.integer, value.kind == Integer && !value.unlimited
}

func (value Value) Unlimited() bool {
	return value.kind == Integer && value.unlimited
}

func (definition Definition) Accepts(value Value) bool {
	return value.validFor(definition)
}

func (value Value) validFor(definition Definition) bool {
	if !definition.valid() || value.kind != definition.valueKind {
		return false
	}
	if value.kind == Boolean {
		return !value.unlimited && value.integer == 0
	}
	if value.boolean || value.integer < 0 {
		return false
	}
	return !value.unlimited || definition.allowUnlimited
}
