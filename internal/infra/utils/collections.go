package utils

import "reflect"

func SomeHasFieldWithValue[T any, K any](items []T, fieldName string, expected K) bool {
	for _, item := range items {
		if HasFieldWithValue(item, fieldName, expected) {
			return true
		}
	}
	return false
}

func HasFieldWithValue[T any, K any](obj T, fieldName string, expected K) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return false
	}
	return reflect.DeepEqual(field.Interface(), expected)
}

func AllTrue(results ...bool) bool {
	for _, b := range results {
		if !b {
			return false
		}
	}
	return true
}

func Map[T any, K any](items []T, fn func(T) K) []K {
	result := make([]K, len(items))
	for i, v := range items {
		result[i] = fn(v)
	}
	return result
}
