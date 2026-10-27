package configfixture

import "time"

type APISetter interface {
	IAMAPISetter
}

func SetAPI(setter APISetter) {
	setter.Helper()
	setCatalogPackageKeyring(setter)
	SetIAMAPI(setter)
	setPhase14API(setter)
	setAnalyticsDashboardExport(setter, false)
	setIAMSystemTransfer(setter)
	setDocuments(setter, false)
	setNotifications(setter, false)
}

type WorkerSetter interface {
	IAMWorkerSetter
}

func SetWorker(setter WorkerSetter) {
	setter.Helper()
	setter.Setenv("GOMEMLIMIT", "512MiB")
	setCatalogPackageKeyring(setter)
	SetIAMWorker(setter)
	setPhase14Worker(setter)
	setRestQueryCursorKeyring(setter)
	setAnalyticsDashboardExport(setter, true)
	setIAMSystemTransfer(setter)
	setDocuments(setter, true)
	setNotifications(setter, true)
}

func APIEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{catalogPackageKeyringEnvironment()},
		IAMAPIEnvironment(fixture, now)...)
	values = append(values, phase14APIEnvironment(fixture)...)
	values = append(values, analyticsDashboardExportEnvironment(fixture, false)...)
	values = append(values, iamSystemTransferEnvironment(fixture)...)
	values = append(values, documentsEnvironment(fixture, false)...)
	values = append(values, notificationsEnvironment(fixture, false))
	return values
}

func WorkerEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{
		"GOMEMLIMIT=512MiB", catalogPackageKeyringEnvironment(),
	}, IAMWorkerEnvironment(fixture, now)...)
	values = append(values, phase14WorkerEnvironment(fixture)...)
	values = append(values, restQueryCursorKeyringEnvironment())
	values = append(values, analyticsDashboardExportEnvironment(fixture, true)...)
	values = append(values, iamSystemTransferEnvironment(fixture)...)
	values = append(values, documentsEnvironment(fixture, true)...)
	values = append(values, notificationsEnvironment(fixture, true))
	return values
}
