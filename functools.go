// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"errors"
	"reflect"
)

func CheckApplyFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 1 {
		return errors.New("Function must be 1 In 1 Out. ")
	}
	inType := fn.In(0)
	if inType != vType {
		return errors.New("Type not match. ")
	}
	return nil
}

func CheckAcceptFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 0 {
		return errors.New("Function must be 1 In 0 Out. ")
	}
	inType := fn.In(0)
	if inType != vType {
		return errors.New("Type not match. ")
	}
	return nil
}

func RunApply(fn reflect.Value, v reflect.Value) reflect.Value {
	return fn.Call([]reflect.Value{v})[0]
}
