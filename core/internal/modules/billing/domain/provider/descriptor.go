package provider

type Capabilities struct {
	Commands       bool
	Webhooks       bool
	Reconciliation bool
}

type Descriptor struct {
	key          Key
	capabilities Capabilities
}

func NewDescriptor(key Key, capabilities Capabilities) (Descriptor, error) {
	if !key.Valid() || (!capabilities.Commands && !capabilities.Webhooks &&
		!capabilities.Reconciliation) {
		return Descriptor{}, ErrDescriptor
	}
	if key == Manual && (capabilities.Webhooks || !capabilities.Commands ||
		!capabilities.Reconciliation) {
		return Descriptor{}, ErrDescriptor
	}
	return Descriptor{key: key, capabilities: capabilities}, nil
}

func ManualDescriptor() Descriptor {
	return Descriptor{
		key:          Manual,
		capabilities: Capabilities{Commands: true, Reconciliation: true},
	}
}

func (descriptor Descriptor) Key() Key                   { return descriptor.key }
func (descriptor Descriptor) Capabilities() Capabilities { return descriptor.capabilities }

func (descriptor Descriptor) valid() bool {
	validated, err := NewDescriptor(descriptor.key, descriptor.capabilities)
	return err == nil && validated == descriptor
}
