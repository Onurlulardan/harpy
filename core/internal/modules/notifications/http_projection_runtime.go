package notifications

func (runtime *HTTPRuntime) NewRealtimeProjectionRuntime(
	transport HTTPRealtimeTransport,
) (*RealtimeProjectionRuntime, error) {
	if runtime == nil {
		return nil, ErrHTTPUnavailable
	}
	return NewRealtimeProjectionRuntime(runtime.resolver, runtime.coordinator,
		runtime.projection, transport)
}
