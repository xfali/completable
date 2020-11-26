// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"errors"
	"reflect"
	"runtime"
	"time"
)

type Nil struct{}

var (
	timeoutError  = errors.New("Timeout. ")
	gTimeoutError = newTimeoutError()

	gNil     = &Nil{}
	NilType  = reflect.TypeOf(gNil)
	NilValue = reflect.ValueOf(gNil)
)

type ValueOrError interface {
	// 获得值
	GetValue() reflect.Value

	// 获得错误
	GetError() error

	// 是否操作
	IsTimeout() bool
}

type ValueHandler interface {
	// 设置值，如果已经存在值或者错误则返回失败
	SetValue(v reflect.Value) error

	// 设置错误，如果已经存在值或者错误则panic
	SetError(err error)

	// 获得value的type
	Type() reflect.Type

	// 等待并获得ValueOrError（线程安全）
	// timeout：等待超时时间
	Get(timeout time.Duration) ValueOrError

	// 在两个ValueHandler中选择最先返回的ValueOrError（线程安全）
	// timeout：等待超时时间
	SelectValue(other ValueHandler, timeout time.Duration) ValueOrError

	// 同时等待并返回两个ValueHandler返回的ValueOrError（线程安全）
	// timeout：等待超时时间
	BothValue(other ValueHandler, timeout time.Duration) (v1, v2 ValueOrError)
}

type vOrErr struct {
	v   reflect.Value
	err error
}

func newTimeoutError() *vOrErr {
	return &vOrErr{
		err: timeoutError,
	}
}

type defaultValueHandler struct {
	t         reflect.Type
	valueChan chan ValueOrError
	isErr     int32
}

func (ve vOrErr) GetValue() reflect.Value {
	return ve.v
}

func (ve vOrErr) GetError() error {
	return ve.err
}

func (ve vOrErr) IsTimeout() bool {
	return ve.err == timeoutError
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
			v: v,
		}
		return nil
	} else {
		return errors.New("Already have a value. ")
	}
}

func (vh *defaultValueHandler) SetError(err error) {
	if len(vh.valueChan) == 0 {
		vh.valueChan <- vOrErr{
			err: err,
		}
	} else {
		panic("Already have a value")
	}
}

func (vh *defaultValueHandler) Type() reflect.Type {
	return vh.t
}

func (vh *defaultValueHandler) Get(timeout time.Duration) ValueOrError {
	if timeout <= 0 {
		return <-vh.valueChan
	} else {
		select {
		case v := <-vh.valueChan:
			return v
		case <-time.After(timeout):
			return newTimeoutError()
		}
	}
}

func (vh *defaultValueHandler) SelectValue(ovh ValueHandler, timeout time.Duration) ValueOrError {
	other := ovh.(*defaultValueHandler)
	if timeout <= 0 {
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
		case <-time.After(timeout):
			return newTimeoutError()
		}
	}
}

func (vh *defaultValueHandler) BothValue(ovh ValueHandler, timeout time.Duration) (v1, v2 ValueOrError) {
	other := ovh.(*defaultValueHandler)
	if timeout <= 0 {
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
		timer := time.NewTimer(timeout)
		s1, s2 := false, false
		for {
			if !s1 {
				select {
				case v1 = <-vh.valueChan:
					s1 = true
				case <-timer.C:
					v1 = newTimeoutError()
					return
				default:
				}
			}
			if !s2 {
				select {
				case v2 = <-other.valueChan:
					s2 = true
				case <-timer.C:
					v2 = newTimeoutError()
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
