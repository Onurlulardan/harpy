package profile

import "reflect"

type Commands struct {
	port       CommandPort
	references ReferencePort
}

func NewCommands(port CommandPort, references ReferencePort) (*Commands, error) {
	if interfaceNil(port) || interfaceNil(references) {
		return nil, ErrEvidenceMissing
	}
	return &Commands{port: port, references: references}, nil
}

func interfaceNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
