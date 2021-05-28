package errors

import (
	"bytes"
	"errors"
)

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

type WrapError struct {
	err []error
}

func (e *WrapError) Error() string {
	buf := &bytes.Buffer{}
	for _, err := range e.err {
		buf.WriteString(err.Error())
		buf.WriteString("\n")
	}
	return buf.String()
}

func NewWrapError(argErrs ...error) error {
	if len(argErrs) == 0 {
		return nil
	}
	var errs []error
	for _, err := range argErrs {
		if err == nil {
			continue
		}
		if val, ok := err.(*WrapError); ok {
			errs = append(errs, val.err...)
		} else {
			errs = append(errs, err)
		}
	}
	return &WrapError{
		err: errs,
	}
}
