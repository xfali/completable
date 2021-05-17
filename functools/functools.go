// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package functools

import (
	"errors"
	"reflect"
)

type Nil struct{}

var (
	gNil          = (*Nil)(nil)
	NilType       = reflect.TypeOf(gNil)
	NilValue      = reflect.ValueOf(gNil)
	InterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
)

func CheckSupplyFunction(fn reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 0 || fn.NumOut() != 1 {
		return errors.New("Type must be f func() TYPE . in[0] Function must be 0 In 1 Out. ")
	}
	return nil
}

func CheckApplyFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 1 {
		return errors.New("Type must be f func( TYPE) Type2 . in[0] Function must be 1 In 1 Out. ")
	}
	inType := fn.In(0)
	if vType != NilType && inType != vType {
		return errors.New("Type must be f func( TYPE) Type2 . in[0] not match. ")
	}
	return nil
}

func CheckAcceptFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 0 {
		return errors.New("Type must be f func( TYPE) . Function must be 1 In 0 Out. ")
	}
	inType := fn.In(0)
	if vType != NilType && inType != vType {
		return errors.New("Type must be f func( TYPE) . in[0] not match. ")
	}
	return nil
}

func CheckRunnableFunction(fn reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 0 || fn.NumOut() != 0 {
		return errors.New("Type must be f func() . Function must be 0 In 0 Out. ")
	}
	return nil
}

func CheckCombineFunction(fn reflect.Type, vType1, vType2 reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 1 {
		return errors.New("Type must be f func( TYPE,  Type2) Type3 . in[1] Function must be 2 In 1 Out. ")
	}
	inType1 := fn.In(0)
	if vType1 != NilType && inType1 != vType1 {
		return errors.New("Type must be f func( TYPE,  Type2) Type3 . in[0] not match. ")
	}

	inType2 := fn.In(1)
	if vType2 != NilType && inType2 != vType2 {
		return errors.New("Type must be f func( TYPE,  Type2) Type3 . in[1] not match. ")
	}
	return nil
}

func CheckAcceptBothFunction(fn reflect.Type, vType1, vType2 reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 0 {
		return errors.New("Type must be f func( TYPE,  Type2) . number not match. ")
	}
	inType1 := fn.In(0)
	if vType1 != NilType && inType1 != vType1 {
		return errors.New("Type must be f func( TYPE,  Type2) . in[0] not match. ")
	}

	inType2 := fn.In(1)
	if vType2 != NilType && inType2 != vType2 {
		return errors.New("Type must be f func( TYPE,  Type2) . in[1] not match. ")
	}
	return nil
}

func CheckHandleFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 1 {
		return errors.New("Type must be f func(o TYPE1, err interface{}) TYPE2. number not match. ")
	}
	inType := fn.In(0)
	if vType != NilType && inType != vType {
		return errors.New("Type must be f func(o TYPE1, err interface{}) TYPE2. in[0] not match. ")
	}

	return nil
}

func CheckWhenCompleteFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 2 || fn.NumOut() != 0 {
		return errors.New("Type must be f func(o TYPE1, err interface{}). number not match. ")
	}
	inType := fn.In(0)
	if vType != NilType && inType != vType {
		return errors.New("Type must be f func(o TYPE1, err interface{}) . in[0] not match. ")
	}

	return nil
}

func CheckPanicFunction(fn reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 1 {
		return errors.New("Type must be f func(o interface{}) TYPE. number not match. ")
	}

	return nil
}

func CheckPtr(v reflect.Type) error {
	if v.Kind() != reflect.Ptr {
		return errors.New("Not a pointer. ")
	}
	return nil
}

func RunSupply(fn reflect.Value) reflect.Value {
	return fn.Call(nil)[0]
}

func RunApply(fn reflect.Value, v reflect.Value) reflect.Value {
	if v == NilValue {
		v = reflect.Zero(fn.Type().In(0))
	}
	return fn.Call([]reflect.Value{v})[0]
}

func RunAccept(fn reflect.Value, v reflect.Value) {
	if v == NilValue {
		v = reflect.Zero(fn.Type().In(0))
	}
	fn.Call([]reflect.Value{v})
}

func RunRunnable(fn reflect.Value) {
	fn.Call(nil)
}

func RunCombine(fn reflect.Value, v1, v2 reflect.Value) reflect.Value {
	if v1 == NilValue {
		v1 = reflect.Zero(fn.Type().In(0))
	}
	if v2 == NilValue {
		v2 = reflect.Zero(fn.Type().In(1))
	}
	return fn.Call([]reflect.Value{v1, v2})[0]
}

func RunAcceptBoth(fn reflect.Value, v1, v2 reflect.Value) {
	if v1 == NilValue {
		v1 = reflect.Zero(fn.Type().In(0))
	}
	if v2 == NilValue {
		v2 = reflect.Zero(fn.Type().In(1))
	}
	fn.Call([]reflect.Value{v1, v2})
}

func RunCompose(fn reflect.Value, v reflect.Value) reflect.Value {
	if v == NilValue {
		v = reflect.Zero(fn.Type().In(0))
	}
	return fn.Call([]reflect.Value{v})[0]
}

func RunHandle(fn reflect.Value, v1, v2 reflect.Value) reflect.Value {
	if v1 == NilValue {
		v1 = reflect.Zero(fn.Type().In(0))
	}
	if v2 == NilValue {
		v2 = reflect.Zero(fn.Type().In(1))
	}
	return fn.Call([]reflect.Value{v1, v2})[0]
}

func RunWhenComplete(fn reflect.Value, v1, v2 reflect.Value) {
	if v1 == NilValue {
		v1 = reflect.Zero(fn.Type().In(0))
	}
	if v2 == NilValue {
		v2 = reflect.Zero(fn.Type().In(1))
	}
	fn.Call([]reflect.Value{v1, v2})
}

func RunPanic(fn reflect.Value, v reflect.Value) reflect.Value {
	if v == NilValue {
		v = reflect.Zero(fn.Type().In(0))
	}
	return fn.Call([]reflect.Value{v})[0]
}
