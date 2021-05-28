package validation

import (
	"fmt"
	"reflect"
)

type ErrRequiredField struct {
	key     string
	message string
}

var _ error = new(ErrRequiredField)

func (e *ErrRequiredField) Error() string {
	return fmt.Sprintf("%s %s", e.key, e.message)
}

func newErrorRequired(key, message string) error {
	return fmt.Errorf("%s %s", key, message)
}

func requiredValidate(f Field) bool {
	var (
		spec        interface{}         = f.GetSpec()
		field       reflect.Value       = f.GetField()
		structField reflect.StructField = f.GetStructField()
	)

	switch field.Kind() {
	case reflect.String:
		if field.String() == "" {
			return true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() == 0 {
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() == 0 {
			return true
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() == 0 {
			return true
		}
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		if field.IsNil() {
			return true
		}
		length := field.Len()
		if length == 0 {
			return true
		} else {
			args := f.GetArgs()
			if args[0] == "deep" {
				if field.Kind() == reflect.Map {
					iter := field.MapRange()
					for iter.Next() {
						v := iter.Value()
						newF := newField(spec, v, structField)
						if err := requiredValidate(newF); err {
							return err
						}
					}
				} else {
					for i := 0; i < length; i++ {
						f := field.Index(i)
						newF := newField(spec, f, structField)
						if err := requiredValidate(newF); err {
							return err
						}
					}
				}
			}
		}
	case reflect.Ptr:
		if field.IsNil() {
			return true
		} else {
			field = field.Elem()
			newF := newField(spec, field, structField)
			return requiredValidate(newF)
		}
	}
	return false
}
