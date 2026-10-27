package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func writeField(writer *identityWriter, field api.FieldDefinition) {
	identifier, err := field.ID()
	if err != nil {
		writer.err = fmt.Errorf("resolved REST Query field is invalid")
		return
	}
	writer.writeBytes(identifier[:])
}

func writeOutputKey(writer *identityWriter, key OutputKey) {
	value, err := key.Value()
	if err != nil {
		writer.err = err
		return
	}
	writer.writeString(value)
}

func writeRequestPath(writer *identityWriter, path RequestPath) {
	value, err := path.Value()
	if err != nil {
		writer.err = err
		return
	}
	writer.writeString(value)
}

func writeValue(writer *identityWriter, value Value) {
	kind, err := value.Kind()
	if err != nil {
		writer.err = err
		return
	}
	writer.writeByte(byte(kind))
	if scalar, found, scalarErr := value.Scalar(); scalarErr != nil {
		writer.err = scalarErr
	} else if found {
		writer.writeByte(1)
		writer.writeString(scalar)
		return
	}
	if boolean, found, booleanErr := value.Boolean(); booleanErr != nil {
		writer.err = booleanErr
	} else if found {
		writer.writeByte(2)
		writer.writeBool(boolean)
		return
	}
	document, found, documentErr := value.JSON()
	if documentErr != nil || !found {
		writer.err = fmt.Errorf("resolved REST Query value is invalid")
		return
	}
	writer.writeByte(3)
	writer.writeBytes(document)
}

func writeProjection(writer *identityWriter, projection Projection) {
	writeField(writer, projection.field)
	writeOutputKey(writer, projection.key)
	writeRequestPath(writer, projection.path)
	writer.writeBool(projection.hidden)
}

func writeSort(writer *identityWriter, sort Sort) {
	writeField(writer, sort.field)
	writer.writeByte(byte(sort.direction))
	writeRequestPath(writer, sort.path)
	writer.writeBool(sort.implicit)
}
