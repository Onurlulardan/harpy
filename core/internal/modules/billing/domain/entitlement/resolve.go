package entitlement

import "slices"

func Resolve(
	definitions map[Key]Definition,
	base map[Key]Value,
	subscription map[Key]Value,
	overrides map[Key]Value,
) (map[Key]Value, error) {
	if definitions == nil || base == nil {
		return nil, ErrResolution
	}
	resolved := make(map[Key]Value, len(base)+len(subscription)+len(overrides))
	for _, layer := range []map[Key]Value{base, subscription, overrides} {
		keys := make([]Key, 0, len(layer))
		for key := range layer {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			definition, exists := definitions[key]
			if !exists || definition.key != key || !layer[key].validFor(definition) {
				return nil, ErrResolution
			}
			resolved[key] = layer[key]
		}
	}
	return resolved, nil
}
