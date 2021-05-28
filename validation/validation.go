package validation

import (
	"bytes"
	"errors"
	"reflect"
	"strings"

	"github.com/oidc-proxy-ecosystem/go-tag"
)

type Valid struct {
	Fn            ValidationFunc
	CustomMessage CustomMessage
}

type ValidationFunc func(field Field) bool
type CustomMessage func(template string, field Field, args ...string) string

var (
	ErrInvalidSpecification       = errors.New("specification must be a struct pointer")
	ErrInvalidMethodFunction      = errors.New("The spec must be a function that returns an error")
	ErrOverrideValidationFunction = errors.New("Set `override` to` true` when registering an already registered validation.")
	validation                    = map[string]Valid{
		required: {
			Fn: requiredValidate,
			CustomMessage: func(template string, field Field, args ...string) string {
				return strings.ReplaceAll(template, "{0}", field.GetStructField().Name)
			},
		},
	}
)

// RegisterValidation is It is possible to add the original validation.
//
// If you want to update the validation that has already been registered, set `orverride` to` true`.
func RegisterValidation(name string, fn Valid, override bool) error {
	if override {
		validation[name] = fn
	} else {
		if _, ok := validation[name]; ok {
			return ErrOverrideValidationFunction
		} else {
			validation[name] = fn
		}
	}
	return nil
}

// MustRegisterValidation is Update and register the already registered validations.
func MustRegisterValidation(name string, fn Valid) {
	RegisterValidation(name, fn, true)
}

func callByMethod(spec reflect.Value, methodName string) (error, bool) {
	if spec.Kind() == reflect.Ptr {
		if spec.IsNil() {
			return nil, false
		} else {
			spec = spec.Elem()
		}
	}
	sTyp := reflect.PtrTo(spec.Type())
	s := reflect.New(sTyp.Elem())
	s.Elem().Set(spec)
	rTyp := s.Type()
	if _, ok := rTyp.MethodByName(methodName); ok {
		method := s.MethodByName(methodName)
		typeOfFunc := method.Type()
		numOut := typeOfFunc.NumOut()
		if numOut != 1 {
			return nil, false
		}
		typeOut := typeOfFunc.Out(0)
		if typeOut.Kind() != reflect.Interface {
			return nil, false
		}
		outs := method.Call(nil)
		out := outs[0]
		err, ok := out.Interface().(error)
		if ok {
			return err, true
		} else {
			return nil, false
		}
	}
	return nil, false
}

type ValidateError struct {
	Errors []error
}

var _ error = new(ValidateError)

func (e *ValidateError) Error() string {
	buff := bytes.NewBufferString("")
	for _, err := range e.Errors {
		buff.WriteString(err.Error())
		buff.WriteString("\n")
	}
	return buff.String()
}

func (e *ValidateError) IsError() bool {
	return len(e.Errors) > 0
}

func (e *ValidateError) ErrorCount() int {
	return len(e.Errors)
}

func newValidateError(errs ...error) *ValidateError {
	if len(errs) == 0 {
		return nil
	}
	return &ValidateError{
		Errors: errs,
	}
}

type Validate interface {
	// StructValidate is spec is pointer struct
	StructValidate(spec interface{}) error
}

type validateImpl struct {
	builder *Builder
}

func (v *validateImpl) getMessage(key string) string {
	if msg, ok := v.builder.Options.Translate[key]; ok {
		return msg
	}
	return ""
}

func (v *validateImpl) StructValidate(spec interface{}) error {
	tags, err := tag.New(spec)
	if err != nil {
		return err
	}
	var errs []error
	s := reflect.ValueOf(spec)
	if s.Kind() != reflect.Ptr {
		return newValidateError(ErrInvalidSpecification)
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return newValidateError(ErrInvalidSpecification)
	}
	for _, structTag := range tags {
		for _, tag := range structTag.Tags {
			if tag.Key == "validate" {
				values := strings.Split(tag.Value, ",")
				for _, value := range values {
					validNames := strings.SplitN(value, "=", 2)
					if len(validNames) < 2 {
						validNames = append(validNames, "")
					}
					funcName := validNames[0]
					if valid, ok := validation[funcName]; ok {
						f := newField(spec, structTag.Field, structTag.TypField, validNames[1:]...)
						if err := valid.Fn(f); err {
							templateMessage := v.getMessage(funcName)
							msg := valid.CustomMessage(templateMessage, f, validNames[1:]...)
							errs = append(errs, errors.New(msg))
						}
					} else {
						// 構造体のメソッド呼び出し
						if err, ok := callByMethod(s, value); ok {
							if err != nil {
								errs = append(errs, err)
							}
						} else {
							// フィールド型のメソッド呼び出し
							err, ok := callByMethod(structTag.Field, value)
							if err != nil {
								if ok {
									errs = append(errs, err)
								} else {
									return newValidateError(err)
								}
							}
						}
					}
				}
			}
		}
	}
	if len(errs) > 0 {
		return newValidateError(errs...)
	}
	return nil
}
