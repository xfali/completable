// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"context"
	"errors"
	"github.com/xfali/executor"
	"reflect"
	"sync/atomic"
	"time"
)

const (
	DefaultExecutorSize   = 1024
	DefaultExecBufferSize = 64
)

var defaultExecutor executor.Executor

func init() {
	defaultExecutor = executor.NewFixedBufExecutor(DefaultExecutorSize, DefaultExecBufferSize)
}

const (
	completableFutureNone = iota
	completableFutureDone
	completableFutureCancel
)

type CompletableFuture struct {
	vType      reflect.Type
	v          ValueHandler
	ctx        context.Context
	cancelFunc context.CancelFunc

	status int32
}

func newCf(pCtx context.Context, v *defaultValueHandler) *CompletableFuture {
	ret := &CompletableFuture{
		v: v,
	}
	if v != nil {
		ret.vType = v.Type()
	}
	if pCtx != nil {
		ctx, cancel := context.WithCancel(pCtx)
		ret.ctx = ctx
		ret.cancelFunc = cancel
	}
	return ret
}

func newCfWithCancel(pCtx context.Context, cancelFunc context.CancelFunc, v *defaultValueHandler) *CompletableFuture {
	ret := &CompletableFuture{
		v: v,
	}
	if v != nil {
		ret.vType = v.Type()
	}
	if pCtx != nil {
		ret.ctx = pCtx
		ret.cancelFunc = cancelFunc
	}
	return ret
}

// 当阶段正常完成时执行参数函数：进行类型变换
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenApply(applyFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	retCf = newCf(cf.ctx, vh)
	defer handlePanic(vh)

	ve := cf.getValue(nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAccept(acceptFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	ve := cf.getValue(nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	RunAccept(fnValue, ve.GetValue())
	vh.SetValue(NilValue)
	return
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		RunAccept(fnValue, ve.GetValue())
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenRun(runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	vh := NewSyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	ve := cf.getValue(nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	runnable()
	vh.SetValue(NilValue)
	return
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenRunAsync(runnable func(), executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	vh := NewAsyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		runnable()
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCombine(other CompletionStage, combineFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(combineFunc)
	if err := CheckCombineFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)

	vh := NewSyncHandler(fnValue.Type().Out(0))
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
	ve1, ve2 := cf.v.BothValue(ocf.v, nil)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}

	if !ve2.HaveValue() {
		vh.SetValueOrError(ve2.Clone())
		return
	}

	err := vh.SetValue(RunCombine(fnValue, ve1.GetValue(), ve2.GetValue()))
	if err != nil {
		vh.SetError(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCombineAsync(
	other CompletionStage,
	combineFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(combineFunc)
	if err := CheckCombineFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer cf.setDone()
		ve1, ve2 := cf.v.BothValue(ocf.v, nil)
		if !ve1.HaveValue() {
			vh.SetValueOrError(ve1.Clone())
			return
		}

		if !ve2.HaveValue() {
			vh.SetValueOrError(ve2.Clone())
			return
		}

		err := vh.SetValue(RunCombine(fnValue, ve1.GetValue(), ve2.GetValue()))
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptBoth(other CompletionStage, acceptFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptBothFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)

	vh := NewSyncHandler(NilType)
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)

	ve1, ve2 := cf.v.BothValue(ocf.v, nil)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}

	if !ve2.HaveValue() {
		vh.SetValueOrError(ve2.Clone())
		return
	}

	RunAcceptBoth(fnValue, ve1.GetValue(), ve2.GetValue())
	vh.SetValue(NilValue)
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptBothAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptBothFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)

	vh := NewAsyncHandler(NilType)
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer cf.setDone()
		ve1, ve2 := cf.v.BothValue(ocf.v, nil)
		if !ve1.HaveValue() {
			vh.SetValueOrError(ve1.Clone())
			return
		}

		if !ve2.HaveValue() {
			vh.SetValueOrError(ve2.Clone())
			return
		}

		RunAcceptBoth(fnValue, ve1.GetValue(), ve2.GetValue())
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterBoth(other CompletionStage, runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)

	vh := NewSyncHandler(NilType)
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)

	ve1, ve2 := cf.v.BothValue(ocf.v, nil)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}
	if !ve2.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}
	runnable()
	vh.SetValue(NilValue)

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterBothAsync(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)
	vh := NewAsyncHandler(NilType)
	retCf = newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve1, ve2 := cf.v.BothValue(ocf.v, nil)
		if !ve1.HaveValue() {
			vh.SetValueOrError(ve1.Clone())
			return
		}
		if !ve2.HaveValue() {
			vh.SetValueOrError(ve2.Clone())
			return
		}
		runnable()
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ApplyToEither(other CompletionStage, applyFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	retCf = newCf(cf.ctx, vh)
	ve := cf.v.SelectValue(ocf.v, nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
	if err != nil {
		vh.SetError(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ApplyToEitherAsync(
	other CompletionStage,
	applyFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) AcceptEither(other CompletionStage, acceptFunc interface{}) (retCf CompletionStage)  {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	ve := cf.v.SelectValue(ocf.v, nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	RunAccept(fnValue, ve.GetValue())
	err := vh.SetValue(NilValue)
	if err != nil {
		vh.SetError(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) AcceptEitherAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		RunAccept(fnValue, ve.GetValue())
		err := vh.SetValue(NilValue)
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterEither(other CompletionStage, runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	vh := NewSyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	ve := cf.v.SelectValue(ocf.v, nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	runnable()
	err := vh.SetValue(NilValue)
	if err != nil {
		vh.SetError(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterAsyncEither(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	vh := NewAsyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		runnable()
		err := vh.SetValue(NilValue)
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCompose(f interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := CheckComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	ve := cf.getValue(nil)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return newCf(cf.ctx, vh)
	}
	newCom := RunCompose(fnValue, ve.GetValue())
	if newCom.IsValid() {
		i := newCom.Interface()
		if i == nil {
			panic("Return CompletionStage is nil. ")
		}
		return i.(*CompletableFuture)
		//err := ncf.v.SetValue(ve.GetValue())
		//if err != nil {
		//	ncf.v.SetError(err)
		//}
		//return ncf
	}
	return nil
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenComposeAsync(f interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := CheckComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	retCf = newCf(cf.ctx, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(nil)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		newCom := RunCompose(fnValue, ve.GetValue())
		if newCom.IsValid() {
			i := newCom.Interface()
			if i == nil {
				vh.SetError(errors.New("Return CompletionStage is nil. "))
				return
			}
			vh.SetValue(reflect.ValueOf(&composeCf{cf: i.(*CompletableFuture)}))
		} else {
			vh.SetError(errors.New("Return CompletionStage is nil. "))
		}
	})
	if err != nil {
		panic(err)
	}

	return
}

// 尝试获得ValueOrError
// 此处还负责处理ComposeAsync封装的CompletableFuture，该设计可能不那么“优雅”
func (cf *CompletableFuture) getValue(ctx context.Context) ValueOrError {
	ve := cf.v.Get(ctx)
	if ve.GetError() == nil {
		v := ve.GetValue()
		if v.IsValid() {
			if c, ok := v.Interface().(*composeCf); ok {
				return c.cf.getValue(ctx)
			}
		}
	}
	return ve
}

// 捕获阶段异常，返回补偿结果
// Param：参数函数，参数：捕获的panic参数，返回补偿的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) Panic(f func(o interface{}) interface{}) CompletionStage {
	return nil
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Return：新的CompletionStage
func (cf *CompletableFuture) WhenComplete(f interface{}) CompletionStage {
	return nil
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) WhenCompleteAsync(f interface{}, executor ...executor.Executor) CompletionStage {
	return nil
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) Handle(f interface{}) CompletionStage {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := CheckHandleFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	ve := cf.getValue(nil)
	v := ve.GetValue()
	if v.IsValid() {
		v = reflect.New(cf.vType).Elem()
	}
	p := ve.GetPanic()

	err := vh.SetValue(RunHandle(fnValue, v, reflect.ValueOf(p)))
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) HandleAsync(f interface{}, executor ...executor.Executor) CompletionStage {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := CheckHandleFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ve := cf.getValue(nil)
		v := ve.GetValue()
		if v.IsValid() {
			v = reflect.New(cf.vType).Elem()
		}
		p := ve.GetPanic()

		err := vh.SetValue(RunHandle(fnValue, v, reflect.ValueOf(p)))
		if err != nil {
			vh.SetError(err)
		}
	})

	if err != nil {
		panic(err)
	}

	return newCf(cf.ctx, vh)
}

// 取消并打断stage链，退出任务
// 如果任务已完成返回false，成功取消返回true
func (cf *CompletableFuture) Cancel() bool {
	if cf.cancelFunc != nil {
		cf.cancelFunc()
		cf.setCancel()
	}
	return false
}

// 是否在完成前被取消
func (cf *CompletableFuture) IsCancelled() bool {
	return atomic.LoadInt32(&cf.status) == completableFutureCancel
}

// 是否任务完成
// 当任务正常完成，被取消，抛出异常都会返回true
func (cf *CompletableFuture) IsDone() bool {
	return atomic.LoadInt32(&cf.status) != completableFutureNone
}

// 等待并获得任务执行结果
// Param： result 目标结果，必须为同类型的指针
// Param： timeout 等待超时时间，如果不传值则一直等待
func (cf *CompletableFuture) Get(result interface{}, timeout ...time.Duration) error {
	cf.checkValue()
	var ve ValueOrError
	if len(timeout) > 0 {
		ctx, _ := context.WithTimeout(cf.ctx, timeout[0])
		ve = cf.getValue(ctx)
	} else {
		ve = cf.getValue(nil)
	}
	if ve.HavePanic() {
		panic(ve.GetPanic())
	}
	if result == nil {
		return nil
	}
	retValue := reflect.ValueOf(result)
	if err := CheckPtr(retValue.Type()); err != nil {
		return err
	}
	err := ve.GetError()
	if err != nil {
		return err
	}
	v := ve.GetValue()
	if v.IsValid() {
		retValue = retValue.Elem()
		if !retValue.CanSet() {
			return errors.New("Cannot set. ")
		}
		retValue.Set(v)
	}

	return nil
}

func (cf *CompletableFuture) setDone() bool {
	return atomic.CompareAndSwapInt32(&cf.status, completableFutureNone, completableFutureDone)
}

func (cf *CompletableFuture) setCancel() bool {
	return atomic.CompareAndSwapInt32(&cf.status, completableFutureNone, completableFutureCancel)
}

func (cf *CompletableFuture) checkValue() {
	if cf.v == nil {
		panic("Without value, cannot be here")
	}
}

func (cf *CompletableFuture) checkSameType(other *CompletableFuture) {
	if cf.vType != other.vType {
		panic("Not same type!")
	}
}

func handlePanic(handler *defaultValueHandler) {
	if r := recover(); r != nil {
		handler.SetPanic(r)
	}
}

func convert(stage CompletionStage) *CompletableFuture {
	if v, ok := stage.(*CompletableFuture); ok {
		return v
	}
	panic("CompletionStage is not *CompletableFuture")
}

func (cf *CompletableFuture) chooseExecutor(executor ...executor.Executor) executor.Executor {
	if len(executor) == 0 {
		return defaultExecutor
	} else {
		return executor[0]
	}
}

type composeCf struct {
	cf *CompletableFuture
}
