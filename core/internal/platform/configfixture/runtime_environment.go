package configfixture

import "time"

type APISetter interface {
	IAMAPISetter
}

func SetAPI(setter APISetter) {
	setter.Helper()
	SetIAMAPI(setter)
	setPhase14API(setter)
	setDocuments(setter, false)
	setNotifications(setter, false)
}

type WorkerSetter interface {
	IAMWorkerSetter
}

func SetWorker(setter WorkerSetter) {
	setter.Helper()
	setter.Setenv("GOMEMLIMIT", "512MiB")
	SetIAMWorker(setter)
	setPhase14Worker(setter)
	setRestQueryCursorKeyring(setter)
	setDocuments(setter, true)
	setNotifications(setter, true)
}

func APIEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{}, IAMAPIEnvironment(fixture, now)...)
	values = append(values, phase14APIEnvironment(fixture)...)
	values = append(values, documentsEnvironment(fixture, false)...)
	values = append(values, notificationsEnvironment(fixture, false))
	return values
}

func WorkerEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{
		"GOMEMLIMIT=512MiB",
	}, IAMWorkerEnvironment(fixture, now)...)
	values = append(values, phase14WorkerEnvironment(fixture)...)
	values = append(values, restQueryCursorKeyringEnvironment())
	values = append(values, documentsEnvironment(fixture, true)...)
	values = append(values, notificationsEnvironment(fixture, true))
	return values
}
