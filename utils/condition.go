package utils

import (
	"reflect"
	"strings"
)

func IfAssign[T any](cond bool, first, second T) T {
	if cond {
		return first
	}
	return second
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func IsEmail(i string) bool {
	return strings.Contains(i, "@")
}

func IsEmpty[T string | []interface{}](value *T) bool {
	if IsNil(value) {
		return true
	}
	return len(*value) > 0
}
