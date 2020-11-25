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

type ValueHandler struct {
	t         reflect.Type
	valueChan chan ValueOrError
	isErr     int32
}

var timeoutError = errors.New("Timeout. ")
var gTimeoutError = newTimeoutError()

type Nil struct{}
var gNil = &Nil{}
var NilType = reflect.TypeOf(gNil)
var NilValue = reflect.ValueOf(gNil)

type ValueOrError interface {
	GetValue() reflect.Value
	GetError() error
	IsTimeout() bool
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

func (ve vOrErr) GetValue() reflect.Value {
	return ve.v
}

func (ve vOrErr) GetError() error {
	return ve.err
}

func (ve vOrErr) IsTimeout() bool {
	return ve.err == timeoutError
}

func NewAsyncHanlder(t reflect.Type) *ValueHandler {
	return &ValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError),
	}
}

func NewSyncValue(t reflect.Type) *ValueHandler {
	return &ValueHandler{
		t:         t,
		valueChan: make(chan ValueOrError, 1),
	}
}

func (vh *ValueHandler) SetValue(v reflect.Value) error {
	if v.Type() != vh.t {
		return errors.New("Type not match. ")
	}
	vh.valueChan <- vOrErr{
		v: v,
	}
	return nil
}

func (vh *ValueHandler) SetError(err error) {
	vh.valueChan <- vOrErr{
		err: err,
	}
}

func (vh *ValueHandler) Type() reflect.Type {
	return vh.t
}

func (vh *ValueHandler) Get(timeout time.Duration) ValueOrError {
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

func (vh *ValueHandler) SelectValue(other *ValueHandler, timeout time.Duration) ValueOrError {
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

func (vh *ValueHandler) BothValue(other *ValueHandler, timeout time.Duration) (v1, v2 ValueOrError) {
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
