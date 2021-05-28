package validation

import "reflect"

type Field interface {
	GetSpec() interface{}
	GetField() reflect.Value
	GetStructField() reflect.StructField
	GetArgs() []string
}

type fieldImpl struct {
	spec        interface{}
	field       reflect.Value
	structField reflect.StructField
	args        []string
}

func (f *fieldImpl) GetSpec() interface{} {
	return f.spec
}

func (f *fieldImpl) GetField() reflect.Value {
	return f.field
}

func (f *fieldImpl) GetStructField() reflect.StructField {
	return f.structField
}

func (f *fieldImpl) GetArgs() []string {
	return f.args
}

var _ Field = new(fieldImpl)

func newField(spec interface{}, field reflect.Value, strcutField reflect.StructField, args ...string) Field {
	return &fieldImpl{
		spec:        spec,
		field:       field,
		structField: strcutField,
		args:        args,
	}
}
