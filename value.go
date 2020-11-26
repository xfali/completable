// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"context"
	"errors"
	"reflect"
	"runtime"
)

type Nil struct{}

var (
	gNil     = &Nil{}
	NilType  = reflect.TypeOf(gNil)
	NilValue = reflect.ValueOf(gNil)

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

	// 是否操作
	IsDone() bool
}

type ValueHandler interface {
	// 设置值，如果已经存在值或者错误则返回失败
	SetValue(v reflect.Value) error

	// 设置错误，如果已经存在值或者错误则panic
	SetError(err error)

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

const (
	vOrErrNormal = iota
	vOrErrError
	vOrErrPanic
	vOrErrDone
)

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

func (ve vOrErr) IsDone() bool {
	return ve.status == vOrErrDone
}

func NewAsyncHandler(t reflect.Type) *defaultValueHandler {
	return &defaultValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError),
	}
}

func NewSyncHandler(t reflect.Type) *defaultValueHandler {
	return &defaultValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError, 1),
	}
}

func (vh *defaultValueHandler) SetValue(v reflect.Value) error {
	if v.Type() != vh.t {
		return errors.New("Type not match. ")
	}
	if len(vh.valueChan) == 0 {
		vh.valueChan <- vOrErr{
			v:      v,
			status: vOrErrNormal,
		}
		return nil
	} else {
		return errors.New("Already have a value. ")
	}
}

func (vh *defaultValueHandler) SetError(err error) {
	if len(vh.valueChan) == 0 {
		vh.valueChan <- vOrErr{
			v:      err,
			status: vOrErrError,
		}
	} else {
		panic("Already have a value")
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
		s1, s2 := false, false
		for {
			if !s1 {
				select {
				case v1 = <-vh.valueChan:
					s1 = true
				default:
					if s2 {
						v1 = <-vh.valueChan
						return
					}
				}
			}
			if !s2 {
				select {
				case v2 = <-other.valueChan:
					s2 = true
				default:
					if s1 {
						v2 = <-other.valueChan
						return
					}
				}
			}
			if s1 && s2 {
				return
			} else {
				runtime.Gosched()
			}
		}
	} else {
		s1, s2 := false, false
		for {
			if !s1 {
				select {
				case v1 = <-vh.valueChan:
					s1 = true
				case <-ctx.Done():
					v1 = newDone()
					return
				default:
				}
			}
			if !s2 {
				select {
				case v2 = <-other.valueChan:
					s2 = true
				case <-ctx.Done():
					v2 = newDone()
					return
				default:
				}
			}
			if s1 && s2 {
				return
			} else {
				runtime.Gosched()
			}
		}
	}
}
