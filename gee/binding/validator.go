package binding

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

type StructValidator interface {
	// ValidateStruct 结构体验证，如果错误返回对应的错误信息
	ValidateStruct(any) error
	// Engine 返回对应使用的验证器
	Engine() any
}

var Validator StructValidator = &defaultValidator{}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}

func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

func (d *defaultValidator) Validate(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return d.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return d.Validate(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.Validate(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		} else {
			return validateRet
		}
	default:
		return nil
	}
}

func validate(obj any) error {
	return Validator.ValidateStruct(obj)
}
