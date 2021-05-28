package validation_test

import (
	"errors"
	"log"
	"testing"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
	"github.com/stretchr/testify/assert"
)

var (
	ErrStrcutValid   = errors.New("StrcutValid")
	ErrFieldValidate = errors.New("FieldValidate")
)

type FieldValidate struct {
}

func (f *FieldValidate) FieldValidate() error {
	return ErrFieldValidate
}

type TestStruct struct {
	Test  string            `validate:"required"`
	Test2 int               `validate:"required,StrcutValid"`
	Test3 *FieldValidate    `validate:"required,FieldValidate"`
	Test4 *FieldValidate    `validate:"required"`
	Test5 map[string]string `validate:"required=deep"`
	Test6 chan int          `validate:"required"`
	Test7 map[string]string `validate:"required"`
}

func (t *TestStruct) StrcutValid() error {
	return ErrStrcutValid
}

func TestValidate(t *testing.T) {
	test := &TestStruct{
		Test:  "a", // no required
		Test3: &FieldValidate{},
		Test5: map[string]string{
			"test": "",
		},
		Test7: map[string]string{
			"test": "",
		},
	}
	expecteds := []struct {
		fn       func(t assert.TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
		expected interface{}
		idx      int
	}{
		{
			fn:       assert.Equal,
			expected: "Test2は必須フィールドです。",
			idx:      0,
		},
		{
			fn:       assert.Equal,
			expected: ErrStrcutValid.Error(),
			idx:      1,
		},
		{
			fn:       assert.Equal,
			expected: ErrFieldValidate.Error(),
			idx:      2,
		},
		{
			fn:       assert.Equal,
			expected: "Test4は必須フィールドです。",
			idx:      3,
		},
		{
			fn:       assert.Equal,
			expected: "Test5は必須フィールドです。",
			idx:      4,
		},
		{
			fn:       assert.Equal,
			expected: "Test6は必須フィールドです。",
			idx:      5,
		},
	}
	valid := validation.New()
	validError := valid.StructValidate(test)
	if val, ok := validError.(*validation.ValidateError); ok {
		if assert.Equal(t, true, val.IsError()) && assert.Equal(t, len(expecteds), val.ErrorCount()) {
			for _, expected := range expecteds {
				expected.fn(t, expected.expected, val.Errors[expected.idx].Error())
			}
		} else {
			log.Println(val.Error())
		}

	} else {
		assert.NoError(t, validError)
	}
}
