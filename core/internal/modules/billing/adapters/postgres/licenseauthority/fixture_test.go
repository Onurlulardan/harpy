package licenseauthority

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}
