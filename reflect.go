package mptt

import (
	"reflect"
)

var (
	nullableKinds = map[reflect.Kind]struct{}{
		reflect.Chan:      {},
		reflect.Func:      {},
		reflect.Interface: {},
		reflect.Map:       {},
		reflect.Ptr:       {},
		reflect.Slice:     {},
	}
)

func reflectNew(target interface{}) interface{} {
	if target == nil {
		return nil
	}
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	newStruc := reflect.New(t)
	return newStruc.Interface()
}

// isEmpty only for mptt key field
func isEmpty(object interface{}) bool {
	if object == nil {
		return true
	}
	fieldType := reflect.TypeOf(object)
	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(object).Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(object).Uint() == 0
	case reflect.String:
		return reflect.ValueOf(object).String() == ""
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(object).Float() == 0
	case reflect.Struct:
		return reflect.DeepEqual(object, reflect.New(fieldType).Elem().Interface())
	default:
		return false
	}
}

// IsNil if the object is nil
func IsNil(object interface{}) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	if _, isNullableKind := nullableKinds[value.Kind()]; isNullableKind && value.IsNil() {
		return true
	}
	return false
}

func (t *tree) validateType(n interface{}) error {
	kind := reflect.TypeOf(n).Kind()
	if kind != reflect.Ptr {
		return ModelTypeError
	}
	return nil
}

func (t *tree) equalIDValue(ida, idb interface{}) bool {
	typeA := reflect.TypeOf(ida)
	typeB := reflect.TypeOf(idb)
	// 如果类型不同，返回 false
	if typeA != typeB {
		return false
	}
	return reflect.DeepEqual(ida, idb)
}
