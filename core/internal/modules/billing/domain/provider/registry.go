package provider

type Registry struct {
	descriptors map[Key]Descriptor
}

func NewRegistry(descriptors ...Descriptor) (Registry, error) {
	registered := make(map[Key]Descriptor, len(descriptors))
	for _, descriptor := range descriptors {
		if !descriptor.valid() {
			return Registry{}, ErrRegistry
		}
		if _, duplicate := registered[descriptor.key]; duplicate {
			return Registry{}, ErrRegistry
		}
		registered[descriptor.key] = descriptor
	}
	manual, exists := registered[Manual]
	if !exists || manual != ManualDescriptor() {
		return Registry{}, ErrRegistry
	}
	return Registry{descriptors: registered}, nil
}

func (registry Registry) Resolve(key Key) (Descriptor, error) {
	if registry.descriptors == nil || !key.Valid() {
		return Descriptor{}, ErrUnknown
	}
	descriptor, exists := registry.descriptors[key]
	if !exists || !descriptor.valid() {
		return Descriptor{}, ErrUnknown
	}
	return descriptor, nil
}

func (registry Registry) Contains(key Key) bool {
	_, err := registry.Resolve(key)
	return err == nil
}
