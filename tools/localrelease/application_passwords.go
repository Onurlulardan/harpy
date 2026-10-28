package main

type applicationPasswords struct {
	runtime  string
	migrator string
	backup   string
	restore  string
	observer string
}

func readApplicationPasswords(options prepareOptions) (applicationPasswords, error) {
	inputs := []struct {
		path  string
		label string
	}{
		{options.runtimePassword, "runtime password"},
		{options.migratorPassword, "migrator password"},
		{options.backupPassword, "backup password"},
		{options.restorePassword, "restore password"},
		{options.observerPassword, "reactivation observer password"},
	}
	values := make([]string, len(inputs))
	for index, input := range inputs {
		value, err := readSecret(input.path, input.label)
		if err != nil {
			return applicationPasswords{}, err
		}
		values[index] = value
	}
	return applicationPasswords{
		runtime: values[0], migrator: values[1],
		backup: values[2], restore: values[3],
		observer: values[4],
	}, nil
}
