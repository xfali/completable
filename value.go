// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"errors"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
)

type ValueHandler struct {
	t         reflect.Type
	valueChan chan reflect.Value
	isErr     int32
}

type vhError struct {
	err error
}

func newVhError(err error) reflect.Value {
	ret := vhError{
		err: err,
	}
	return reflect.ValueOf(&ret)
}

var vhErrorType = reflect.TypeOf((*vhError)())

func NewAsyncHanlder(t reflect.Type) *ValueHandler {
	return &ValueHandler{
		t:         t,
		valueChan: make(chan reflect.Value),
	}
}

func NewSyncValue(t reflect.Type) *ValueHandler {
	return &ValueHandler{
		t:         t,
		valueChan: make(chan reflect.Value, 1),
	}
}

func (vh *ValueHandler) SetValue(v reflect.Value) error {
	if v.Type() != vh.t {
		return errors.New("Type not match. ")
	}
	vh.valueChan <- v
	return nil
}

func (vh *ValueHandler) IsError() bool {
	return atomic.LoadInt32(&vh.isErr) == 1
}

func (vh *ValueHandler) SetError(err error) {
	atomic.StoreInt32(&vh.isErr, 1)
	vh.valueChan <- newVhError(err)
}

func (vh *ValueHandler) GetError() error {
	if !vh.IsError() {
		return nil
	}
	v := <-vh.valueChan
	return v.Interface().(*vhError).err
}

func (vh *ValueHandler) Type() reflect.Type {
	return vh.t
}

func (vh *ValueHandler) GetValue(timeout time.Duration) reflect.Value {
	if timeout <= 0 {
		return <-vh.valueChan
	} else {
		select {
		case v := <-vh.valueChan:
			return v
		case <-time.After(timeout):
			return reflect.Value{}
		}
	}
}

func (vh *ValueHandler) SelectValue(other *ValueHandler, timeout time.Duration) reflect.Value {
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
			return reflect.Value{}
		}
	}
}

func (vh *ValueHandler) BothValue(other *ValueHandler, timeout time.Duration) (v1, v2 reflect.Value) {
	if timeout <= 0 {
		s1, s2 := false, false
		for {
			if !s1 {
				select {
				case v1 = <-vh.valueChan:
					s1 = true
				default:
				}
			}
			if !s2 {
				select {
				case v2 = <-other.valueChan:
					s2 = true
				default:
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
					return
				default:
				}
			}
			if !s2 {
				select {
				case v2 = <-other.valueChan:
					s2 = true
				case <-timer.C:
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
