package recoveryanchor

type Store[S any] interface {
	Reader[S]
	OptionsStore[S]
	RegistrationStore[S]
	DeleteStore[S]
	AuthorityStore[S]
}
