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

func CheckCombineFunction(fn reflect.Type, vType1, vType2 reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 1 {
		return errors.New("Function must be 2 In 1 Out. ")
	}
	inType1 := fn.In(0)
	if inType1 != vType1 {
		return errors.New("Type not match. ")
	}

	inType2 := fn.In(1)
	if inType2 != vType2 {
		return errors.New("Type not match. ")
	}
	return nil
}

func CheckAcceptBothFunction(fn reflect.Type, vType1, vType2 reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 0 {
		return errors.New("Function must be 2 In 0 Out. ")
	}
	inType1 := fn.In(0)
	if inType1 != vType1 {
		return errors.New("Type not match. ")
	}

	inType2 := fn.In(1)
	if inType2 != vType2 {
		return errors.New("Type not match. ")
	}
	return nil
}

func RunApply(fn reflect.Value, v reflect.Value) reflect.Value {
	return fn.Call([]reflect.Value{v})[0]
}

func RunAccept(fn reflect.Value, v reflect.Value) {
	fn.Call([]reflect.Value{v})
}

func RunCombine(fn reflect.Value, v1, v2 reflect.Value) reflect.Value {
	return fn.Call([]reflect.Value{v1, v2})[0]
}

func RunAcceptBoth(fn reflect.Value, v1, v2 reflect.Value) {
	fn.Call([]reflect.Value{v1, v2})
}
