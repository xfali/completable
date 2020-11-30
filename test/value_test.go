// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/xfali/completable"
	"reflect"
	"testing"
	"time"
)

func TestSyncValueGet(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value := "Hello world"
		v := reflect.ValueOf(value)
		t := reflect.TypeOf(value)
		vh := completable.NewSyncHandler(t)
		testSet(vh, v, test)
		testGet(vh, nil, test)
	})

	test.Run("ctx Background", func(test *testing.T) {
		value := "Hello world"
		v := reflect.ValueOf(value)
		t := reflect.TypeOf(value)
		vh := completable.NewSyncHandler(t)
		testSet(vh, v, test)
		testGet(vh, context.Background(), test)
	})
}

func TestAsyncValueGet(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value := "Hello world"
		v := reflect.ValueOf(value)
		t := reflect.TypeOf(value)
		vh := completable.NewAsyncHandler(t)
		go func() {
			time.Sleep(time.Second)
			testSet(vh, v, test)
		}()
		testGet(vh, nil, test)
	})
	test.Run("ctx Background", func(test *testing.T) {
		value := "Hello world"
		v := reflect.ValueOf(value)
		t := reflect.TypeOf(value)
		vh := completable.NewAsyncHandler(t)

		go func() {
			time.Sleep(time.Second)
			testSet(vh, v, test)
		}()
		testGet(vh, context.Background(), test)
	})

	test.Run("ctx cancel", func(test *testing.T) {
		value := "Hello world"
		v := reflect.ValueOf(value)
		t := reflect.TypeOf(value)
		vh := completable.NewAsyncHandler(t)

		go func() {
			time.Sleep(time.Second)
			testSet(vh, v, test)
		}()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		testGetDone(vh, ctx, test)
	})
}

func TestSyncValueSelect(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		testSet(vh2, v2, test)
		testSelect(true, vh1, vh2, nil, test)
	})

	test.Run("ctx Background", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		testSet(vh2, v2, test)
		testSelect(true, vh1, vh2, context.Background(), test)
	})
}

func TestAsyncValueSelect(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewAsyncHandler(t1)
		go testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewAsyncHandler(t2)
		go func() {
			time.Sleep(200 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		testSelect(false, vh1, vh2, nil, test)
	})

	test.Run("ctx Background", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewAsyncHandler(t1)
		go testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewAsyncHandler(t2)
		go func() {
			time.Sleep(200 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		testSelect(false, vh1, vh2, context.Background(), test)
	})

	test.Run("ctx cancel", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewAsyncHandler(t1)
		go func() {
			time.Sleep(400 * time.Millisecond)
			testSet(vh1, v1, test)
		}()
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewAsyncHandler(t2)
		go func() {
			time.Sleep(500 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		testSelectDone(false, vh1, vh2, ctx, test)
	})
}

func TestSyncValueBoth(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		testSet(vh2, v2, test)
		testBoth(vh1, vh2, nil, test)
	})

	test.Run("ctx Background", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		testSet(vh2, v2, test)
		testBoth(vh1, vh2, context.Background(), test)
	})
}

func TestAsyncValueBoth(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		go testSet(vh1, v1, test)
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		go func() {
			time.Sleep(200 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		testBoth(vh1, vh2, nil, test)
	})
	test.Run("ctx Background", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		go func() {
			time.Sleep(400 * time.Millisecond)
			testSet(vh1, v1, test)
		}()
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		go func() {
			time.Sleep(500 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		testBoth(vh1, vh2, context.Background(), test)
	})

	test.Run("ctx cancel", func(test *testing.T) {
		value1 := "Hello"
		v1 := reflect.ValueOf(value1)
		t1 := reflect.TypeOf(value1)
		vh1 := completable.NewSyncHandler(t1)
		go func() {
			time.Sleep(400 * time.Millisecond)
			testSet(vh1, v1, test)
		}()
		value2 := "world"
		v2 := reflect.ValueOf(value2)
		t2 := reflect.TypeOf(value2)
		vh2 := completable.NewSyncHandler(t2)
		go func() {
			time.Sleep(500 * time.Millisecond)
			testSet(vh2, v2, test)
		}()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		testBothDone(vh1, vh2, ctx, test)
	})
}

func TestAllOfValue(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		for i := 0; i < 10; i++ {
			vh := completable.NewSyncHandler(reflect.TypeOf(i))
			if i != 2 {
				vh.SetValue(reflect.ValueOf(i))
			} else {
				vh.SetPanic("error")
			}
			vhs[i] = vh
		}
		testAllOfValue(false, nil, test, vhs...)
	})
	test.Run("ctx background", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		for i := 0; i < 10; i++ {
			vh := completable.NewAsyncHandler(reflect.TypeOf(i))
			if i != 2 {
				v := i
				go func() {
					time.Sleep(time.Duration(1+v) * time.Second)
					vh.SetValue(reflect.ValueOf(v))
				}()
			} else {
				go func() {
					time.Sleep(time.Second)
					vh.SetPanic("error")
				}()
			}
			vhs[i] = vh
		}
		now := time.Now()
		testAllOfValue(false, context.Background(), test, vhs...)
		if time.Since(now) < 10*time.Second {
			test.Fatal("not match")
		}
	})

	test.Run("ctx cancel", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		ctx, cancel := context.WithCancel(context.Background())
		for i := 0; i < 10; i++ {
			vh := completable.NewAsyncHandler(reflect.TypeOf(i))
			if i != 2 {
				v := i
				go func() {
					time.Sleep(time.Duration(1+v) * time.Second)
					vh.SetValue(reflect.ValueOf(v))
				}()
			} else {
				go func() {
					time.Sleep(time.Second)
					vh.SetPanic("error")
				}()
			}
			vhs[i] = vh
		}
		go func() {
			time.Sleep(5 * time.Second)
			cancel()
		}()
		now := time.Now()
		testAllOfValue(true, ctx, test, vhs...)
		if time.Since(now) > 6*time.Second {
			test.Fatal("not match")
		}
	})
}

func TestAnyOfValue(test *testing.T) {
	test.Run("ctx nil", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		for i := 0; i < 10; i++ {
			vh := completable.NewSyncHandler(reflect.TypeOf(i))
			if i != 2 {
				vh.SetValue(reflect.ValueOf(i))
			} else {
				vh.SetPanic("error")
			}
			vhs[i] = vh
		}
		testAnyOfValue(false, nil, test, vhs...)
	})
	test.Run("ctx background", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		for i := 0; i < 10; i++ {
			vh := completable.NewAsyncHandler(reflect.TypeOf(i))
			if i != 2 {
				v := i
				go func() {
					time.Sleep(time.Duration(1+v) * time.Second)
					vh.SetValue(reflect.ValueOf(v))
				}()
			} else {
				go func() {
					time.Sleep(time.Second)
					vh.SetPanic("error")
				}()
			}
			vhs[i] = vh
		}
		now := time.Now()
		testAnyOfValue(false, context.Background(), test, vhs...)
		if time.Since(now) > 2*time.Second {
			test.Fatal("not match")
		}
	})

	test.Run("ctx cancel", func(test *testing.T) {
		vhs := make([]completable.ValueHandler, 10)
		ctx, cancel := context.WithCancel(context.Background())
		for i := 0; i < 10; i++ {
			vh := completable.NewAsyncHandler(reflect.TypeOf(i))
			if i != 2 {
				v := i
				go func() {
					time.Sleep(time.Duration(1+v) * time.Second)
					vh.SetValue(reflect.ValueOf(v))
				}()
			} else {
				go func() {
					time.Sleep(time.Second)
					vh.SetPanic("error")
				}()
			}
			vhs[i] = vh
		}

		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		now := time.Now()
		testAnyOfValue(false, ctx, test, vhs...)
		if time.Since(now) > 300*time.Millisecond {
			test.Fatal("not match")
		}
	})
}

func testGet(vh completable.ValueHandler, ctx context.Context, test *testing.T) {
	ret := vh.Get(ctx)
	if !ret.HaveValue() {
		test.Fatal("not match")
	}
	if ret.HaveError() {
		test.Fatal("not match")
	}
	if ret.HavePanic() {
		test.Fatal("not match")
	}

	retV := ret.GetValue()
	if !retV.IsValid() {
		test.Fatal("not match")
	}

	if retV.Interface().(string) != "Hello world" {
		test.Fatal("not match")
	}
}

func testGetDone(vh completable.ValueHandler, ctx context.Context, test *testing.T) {
	ret := vh.Get(ctx)
	if !ret.IsDone() {
		test.Fatal("ret is not done")
	}
}

func testSet(vh completable.ValueHandler, v reflect.Value, test *testing.T) {
	err := vh.SetValue(v)
	if err != nil {
		test.Fatal(err)
	}
	err = vh.SetValue(v)
	if err != nil {
		test.Fatal(err)
	}

	vh.SetPanic("panic")
	if err != nil {
		test.Fatal(err)
	}

	vh.SetPanic("panic")
	if err != nil {
		test.Fatal(err)
	}

	vh.SetError(errors.New("error"))
	if err != nil {
		test.Fatal(err)
	}

	vh.SetError(errors.New("error"))
	if err != nil {
		test.Fatal(err)
	}
}

func testBoth(vh1, vh2 completable.ValueHandler, ctx context.Context, test *testing.T) {
	v1, v2 := vh1.BothValue(vh2, ctx)
	if !v1.HaveValue() {
		test.Fatal("not match")
	}
	if v1.HaveError() {
		test.Fatal("not match")
	}
	if v1.HavePanic() {
		test.Fatal("not match")
	}

	retV1 := v1.GetValue()
	if !retV1.IsValid() {
		test.Fatal("not match")
	}

	if retV1.Interface().(string) != "Hello" {
		test.Fatal("not match")
	}

	if !v2.HaveValue() {
		test.Fatal("not match")
	}
	if v2.HaveError() {
		test.Fatal("not match")
	}
	if v2.HavePanic() {
		test.Fatal("not match")
	}

	retV2 := v2.GetValue()
	if !retV2.IsValid() {
		test.Fatal("not match")
	}

	if retV2.Interface().(string) != "world" {
		test.Fatal("not match")
	}
}

func testBothDone(vh1, vh2 completable.ValueHandler, ctx context.Context, test *testing.T) {
	v1, v2 := vh1.BothValue(vh2, ctx)
	if !v1.IsDone() || !v2.IsDone() {
		test.Fatal("v1, v2 must be done")
	}
}

func testSelect(sync bool, vh1, vh2 completable.ValueHandler, ctx context.Context, test *testing.T) {
	ret := vh1.SelectValue(vh2, ctx)
	if !ret.HaveValue() {
		test.Fatal("not match")
	}
	if ret.HaveError() {
		test.Fatal("not match")
	}
	if ret.HavePanic() {
		test.Fatal("not match")
	}

	retV := ret.GetValue()
	if !retV.IsValid() {
		test.Fatal("not match")
	}

	if !sync {
		if retV.Interface().(string) != "Hello" {
			test.Fatal("not match")
		}
	} else {
		retStr := retV.Interface().(string)
		if retStr != "Hello" && retStr != "world" {
			test.Fatal("not match")
		}
		test.Log(retStr)
	}
}

func testSelectDone(sync bool, vh1, vh2 completable.ValueHandler, ctx context.Context, test *testing.T) {
	ret := vh1.SelectValue(vh2, ctx)
	if !ret.IsDone() {
		test.Fatal("ret must be done")
	}
}

func testAllOfValue(cancel bool, ctx context.Context, test *testing.T, vhs ...completable.ValueHandler) {
	vs := completable.AllOfValue(ctx, vhs...)
	for i, v := range vs {
		if i != 2 {
			if cancel {
				if i < 4 {
					if !v.HaveValue() {
						test.Fatal("not match")
					}
					ret := v.GetValue().Interface().(int)
					if ret != i {
						test.Fatal("not match", ret, i)
					}
				}
			} else {
				if !v.HaveValue() {
					test.Fatal("not match")
				}
				ret := v.GetValue().Interface().(int)
				if ret != i {
					test.Fatal("not match", ret, i)
				}
			}
		} else {
			if !v.HavePanic() {
				test.Fatal("not match")
			}
		}
	}
}

func testAnyOfValue(cancel bool, ctx context.Context, test *testing.T, vhs ...completable.ValueHandler) {
	i, v := completable.AnyOfValue(ctx, vhs...)
	if v.HaveValue() {
		ret := v.GetValue().Interface().(int)
		test.Log(ret)
		if ret != i {
			test.Fatal("not match")
		}
	}
	if v.HavePanic() {

	}
}

func testAllOfValueDone(sync bool, vh1, vh2 completable.ValueHandler, ctx context.Context, test *testing.T) {
	ret := vh1.SelectValue(vh2, ctx)
	if !ret.IsDone() {
		test.Fatal("ret must be done")
	}
}

func TestContext(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	ctx3, cancel3 := context.WithCancel(ctx1)
	ctx4, cancel4 := context.WithCancel(ctx2)

	go func() {
		select {
		case <-ctx3.Done():
			fmt.Println("ctx3 done")
		}
	}()

	go func() {
		select {
		case <-ctx4.Done():
			fmt.Println("ctx4 done")
		}
	}()

	cancel12 := func() {
		cancel1()
		cancel2()
	}
	cancel34 := func() {
		cancel3()
		cancel4()
	}
	time.Sleep(time.Second)
	if true {
		cancel12()
	} else {
		cancel34()
	}
	time.Sleep(time.Second)
}

func TestSelect(t *testing.T) {
	var selectCases []reflect.SelectCase
	var selectVals []chan int
	for i := 0; i < 10; i++ {
		v := make(chan int)
		selectVals = append(selectVals, v)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(v),
		})
	}
	t.Run("3 ok", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second)
			selectVals[3] <- 3
		}()
		i, value, ok := reflect.Select(selectCases)
		if !ok {
			t.Fatal("not match")
		}
		if i != 3 {
			t.Fatal("not match")
		}
		if value.Interface().(int) != 3 {
			t.Fatal("not match")
		}
	})
	t.Run("3 close", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second)
			close(selectVals[3])
		}()
		i, value, ok := reflect.Select(selectCases)
		if ok {
			t.Fatal("not match")
		}
		if i != 3 {
			t.Fatal("not match")
		}
		if !value.IsZero() {
			t.Fatal("not match")
		}
	})
}
