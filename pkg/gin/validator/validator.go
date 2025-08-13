package validator

import (
	"reflect"
	"sync"

	valid "github.com/go-playground/validator/v10"
)

func Init() *CustomValidator {
	v := NewCustomValidator()
	v.Engine()
	return v
}

type CustomValidator struct {
	once     sync.Once
	Validate *valid.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{}
}

func (v *CustomValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		if err := v.Validate.Struct(obj); err != nil {
			return err
		}

	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return v.ValidateStruct(val.Elem().Interface())

	case reflect.Slice, reflect.Array:

		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if err := v.ValidateStruct(elem.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *CustomValidator) Engine() interface{} {
	v.lazyInit()
	return v.Validate
}

func (v *CustomValidator) lazyInit() {
	v.once.Do(func() {
		v.Validate = valid.New()
		v.Validate.SetTagName("binding")
	})
}
