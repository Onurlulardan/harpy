package period

type Cadence string

const (
	Monthly Cadence = "monthly"
	Yearly  Cadence = "yearly"
)

func (cadence Cadence) valid() bool {
	return cadence == Monthly || cadence == Yearly
}
