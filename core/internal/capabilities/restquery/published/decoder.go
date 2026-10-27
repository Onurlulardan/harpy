package published

type GETDecoder interface {
	Decode(rawQuery string, manifest ParameterManifest) (BoundParameters, error)
}

type POSTDecoder interface {
	Decode(document []byte, manifest ParameterManifest) (BoundParameters, error)
}
