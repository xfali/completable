// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
)

type Nil struct{}

const (
	vOrErrNone = iota
	vOrErrNormal
	vOrErrError
	vOrErrPanic
	vOrErrDone
)

const (
	valueHandlerNone = iota
	valueHandlerNormal
	valueHandlerError
	valueHandlerPanic
	valueHandlerUnknown
)

var (
	gNil          = &Nil{}
	NilType       = reflect.TypeOf(gNil)
	NilValue      = reflect.ValueOf(gNil)
	InterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

	doneError  = errors.New("Done. ")
	gDoneError = newDone()
)

type ValueOrError interface {
	// 获得值
	GetValue() reflect.Value

	// 获得错误
	GetError() error

	// 获得panic
	GetPanic() interface{}

	// 返回clone对象
	Clone() ValueOrError

	// 是否正常返回值
	HaveValue() bool

	// 是否存在错误
	HaveError() bool

	// 是否panic
	HavePanic() bool

	// 是否被context控制提前结束
	IsDone() bool
}

type ValueHandler interface {
	// 设置ValueOrError，如果已经存在值或者错误则返回失败
	SetValueOrError(v ValueOrError) error

	// 设置值，如果已经存在值或者错误则返回失败
	SetValue(v reflect.Value) error

	// 设置错误，如果已经存在值或者错误则panic
	SetError(err error)

	// 设置panic，如果已经存在值或者错误则panic
	SetPanic(o interface{})

	// 获得value的type
	Type() reflect.Type

	// 等待并获得ValueOrError（线程安全）
	// ctx：控制context
	Get(ctx context.Context) ValueOrError

	// 在两个ValueHandler中选择最先返回的ValueOrError（线程安全）
	// ctx：控制context
	SelectValue(other ValueHandler, ctx context.Context) ValueOrError

	// 同时等待并返回两个ValueHandler返回的ValueOrError（线程安全）
	// ctx：控制context
	BothValue(other ValueHandler, ctx context.Context) (v1, v2 ValueOrError)
}

type vOrErr struct {
	v      interface{}
	status int32
}

func newDone() *vOrErr {
	return &vOrErr{
		status: vOrErrDone,
	}
}

func newPanic(o interface{}) *vOrErr {
	return &vOrErr{
		v:      o,
		status: vOrErrPanic,
	}
}

type defaultValueHandler struct {
	t         reflect.Type
	valueChan chan ValueOrError
	status    int32
}

func (ve vOrErr) GetValue() reflect.Value {
	if ve.status != vOrErrNormal {
		return reflect.Value{}
	}
	return ve.v.(reflect.Value)
}

func (ve vOrErr) GetError() error {
	if ve.status != vOrErrError {
		return nil
	}
	return ve.v.(error)
}

func (ve vOrErr) GetPanic() interface{} {
	if ve.status != vOrErrPanic {
		return nil
	}
	return ve.v
}

func (ve vOrErr) Clone() ValueOrError {
	return vOrErr{
		v:      ve.v,
		status: ve.status,
	}
}

func (ve vOrErr) HaveValue() bool {
	return ve.status == vOrErrNormal
}

// 是否存在错误
func (ve vOrErr) HaveError() bool {
	return ve.status == vOrErrError
}

// 是否panic
func (ve vOrErr) HavePanic() bool {
	return ve.status == vOrErrPanic
}

func (ve vOrErr) IsDone() bool {
	return ve.status == vOrErrDone
}

func NewAsyncHandler(t reflect.Type) *defaultValueHandler {
	return &defaultValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError),
		status:    valueHandlerNone,
	}
}

func NewSyncHandler(t reflect.Type) *defaultValueHandler {
	return &defaultValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError, 1),
		status:    valueHandlerNone,
	}
}

func (vh *defaultValueHandler) SetValueOrError(v ValueOrError) error {
	ve := v.(vOrErr)
	if atomic.CompareAndSwapInt32(&vh.status, valueHandlerNone, convertStatus(ve.status)) {
		if len(vh.valueChan) == 0 {
			vh.valueChan <- v
			return nil
		} else {
			return errors.New("Already have a value. ")
		}
	} else {
		return errors.New("Already have a value. ")
	}
}

func (vh *defaultValueHandler) SetValue(v reflect.Value) error {
	if v.Type() != vh.t {
		return errors.New("Type not match. ")
	}
	if atomic.CompareAndSwapInt32(&vh.status, valueHandlerNone, valueHandlerNormal) {
		if len(vh.valueChan) == 0 {
			vh.valueChan <- vOrErr{
				v:      v,
				status: vOrErrNormal,
			}
			return nil
		} else {
			return errors.New("Already have a value. ")
		}
	} else {
		// do nothing
		return nil
	}
}

func (vh *defaultValueHandler) SetError(err error) {
	if atomic.CompareAndSwapInt32(&vh.status, valueHandlerNone, valueHandlerError) {
		if len(vh.valueChan) == 0 {
			vh.valueChan <- vOrErr{
				v:      err,
				status: vOrErrError,
			}
		} else {
			panic("Already have a value")
		}
	} else {
		// do nothing
	}
}

func (vh *defaultValueHandler) SetPanic(o interface{}) {
	if atomic.CompareAndSwapInt32(&vh.status, valueHandlerNone, valueHandlerPanic) {
		if len(vh.valueChan) == 0 {
			vh.valueChan <- vOrErr{
				v:      o,
				status: vOrErrPanic,
			}
		} else {
			panic("Already have a value")
		}
	} else {
		// do nothing
	}
}

func (vh *defaultValueHandler) Type() reflect.Type {
	return vh.t
}

func (vh *defaultValueHandler) Get(ctx context.Context) ValueOrError {
	if ctx == nil {
		return <-vh.valueChan
	} else {
		select {
		case v := <-vh.valueChan:
			return v
		case <-ctx.Done():
			return newDone()
		}
	}
}

func (vh *defaultValueHandler) SelectValue(ovh ValueHandler, ctx context.Context) ValueOrError {
	other := ovh.(*defaultValueHandler)
	if ctx == nil {
		select {
		case v := <-vh.valueChan:
			return v
		case v := <-other.valueChan:
			return v
		}
	} else {
		select {
		case v := <-vh.valueChan:
			return v
		case v := <-other.valueChan:
			return v
		case <-ctx.Done():
			return newDone()
		}
	}
}

func (vh *defaultValueHandler) BothValue(ovh ValueHandler, ctx context.Context) (v1, v2 ValueOrError) {
	other := ovh.(*defaultValueHandler)
	if ctx == nil {
		v1 = <-vh.valueChan
		v2 = <-other.valueChan
		return
	} else {
		b1, b2 := false, false
		for {
			select {
			case v1 = <-vh.valueChan:
				b1 = true
			case v2 = <-other.valueChan:
				b2 = true
			case <-ctx.Done():
				if !b1 {
					v1 = newDone()
				}
				if !b2 {
					v2 = newDone()
				}
				return
			}
			if b1 && b2 {
				return
			}
		}
	}
}

func AllOfValue(ctx context.Context, vhs ...ValueHandler) []ValueOrError {
	ret := make([]ValueOrError, len(vhs))
	if ctx == nil {
		ctx = context.Background()
	}
	for i, vh := range vhs {
		select {
		case v := <-vh.(*defaultValueHandler).valueChan:
			ret[i] = v
		case <-ctx.Done():
			ret[i] = newDone()
		}
	}
	return ret
}

func AnyOfValue(ctx context.Context, vhs ...ValueHandler) (int, ValueOrError) {
	size := len(vhs)
	if ctx != nil {
		size++
	}
	selectCases := make([]reflect.SelectCase, size)
	if ctx != nil {
		selectCases[len(vhs)] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
	}
	for i, vh := range vhs {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(vh.(*defaultValueHandler).valueChan),
		}
	}
	index, value, _ := reflect.Select(selectCases)
	if index == len(vhs) {
		return index, newDone()
	}
	return index, value.Interface().(ValueOrError)
}

var vOrErrMap = map[int32]int32{
	vOrErrNone:   valueHandlerNone,
	vOrErrNormal: valueHandlerNormal,
	vOrErrError:  valueHandlerError,
	vOrErrPanic:  valueHandlerPanic,
	vOrErrDone:   valueHandlerUnknown,
}

func convertStatus(vOrErrStatus int32) int32 {
	return vOrErrMap[vOrErrStatus]
}
