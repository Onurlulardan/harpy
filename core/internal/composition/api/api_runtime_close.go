package api

func (runtime apiRuntime) Close() {
	runtime.iam.Close()
	if runtime.customerProvisioning != nil {
		runtime.customerProvisioning.Close()
	}
}
