package api

type Store interface {
	Uploader
	Reader
	Deleter
	Reconciler
}
