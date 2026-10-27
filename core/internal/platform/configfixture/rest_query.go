package configfixture

import "fmt"

const restQueryCursorKeyringEnvironmentKey = "HARPY_REST_QUERY_CURSOR_KEYRING"

func setRestQueryCursorKeyring(setter EnvironmentSetter) {
	setter.Helper()
	setter.Setenv(restQueryCursorKeyringEnvironmentKey, restQueryCursorKeyring())
}

func restQueryCursorKeyringEnvironment() string {
	return restQueryCursorKeyringEnvironmentKey + "=" + restQueryCursorKeyring()
}

func restQueryCursorKeyring() string {
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},`+
			`"previous":{"kid":"previous-1","key":"%s"}}`,
		material(1), material(2),
	)
}
