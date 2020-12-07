// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"context"
	"errors"
	"github.com/xfali/completable/functools"
	"github.com/xfali/executor"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultExecutorSize   = 1024
	DefaultExecBufferSize = 256
)

var defaultExecutor executor.Executor

func init() {
	defaultExecutor = executor.NewFixedBufExecutor(DefaultExecutorSize, DefaultExecBufferSize)
}

func SetDefaultExecutor(executor executor.Executor) {
	defaultExecutor = executor
}

const (
	completableFutureNone = iota
	completableFutureDone
	completableFutureCancel
)

// 注意CompletableFuture的修改原则：
// 1、每个返回的CompletableFuture中的ValueHandler都必须有一个Set操作，不论是value、error、panic（目前无error）
// 2、在1的基础上注意程序或者函数参数造的的panic没有被正确步骤，使得Set操作没有被执行，此时会造成死锁；
//    所以在开发时，要么在创建返回CompletableFuture之前就panic，要么就捕捉panic然后ValueHandler SetPanic
type defaultCompletableFuture struct {
	vType      reflect.Type
	v          ValueHandler
	ctx        context.Context
	cancelFunc context.CancelFunc

	status int32

	lock   sync.Mutex
	result ValueOrError
}

func newCf(pCtx context.Context, v *defaultValueHandler) *defaultCompletableFuture {
	ret := &defaultCompletableFuture{
		v:      v,
		result: nil,
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

func newCfWithCancel(cCtx context.Context, cancelFunc context.CancelFunc, v *defaultValueHandler) *defaultCompletableFuture {
	ret := &defaultCompletableFuture{
		v:      v,
		result: nil,
	}
	if v != nil {
		ret.vType = v.Type()
	}
	if cCtx != nil {
		ret.ctx = cCtx
		ret.cancelFunc = cancelFunc
	}
	return ret
}

// 当阶段正常完成时执行参数函数：进行类型变换
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenApply(applyFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := functools.CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	defer handlePanic(vh)

	ve := cf.getValue(ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	err := vh.SetValue(functools.RunApply(fnValue, ve.GetValue()))
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := functools.CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		err := vh.SetValue(functools.RunApply(fnValue, ve.GetValue()))
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenAccept(acceptFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	functools.RunAccept(fnValue, ve.GetValue())
	vh.SetValue(functools.NilValue)
	return
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		functools.RunAccept(fnValue, ve.GetValue())
		vh.SetValue(functools.NilValue)
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenRun(runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	vh := NewSyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	runnable()
	vh.SetValue(functools.NilValue)
	return
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenRunAsync(runnable func(), executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	vh := NewAsyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		runnable()
		vh.SetValue(functools.NilValue)
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenCombine(other CompletionStage, combineFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(combineFunc)
	if err := functools.CheckCombineFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(fnValue.Type().Out(0))
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)
	ve1, ve2 := cf.v.BothValue(ocf.v, cf.ctx)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}

	if !ve2.HaveValue() {
		vh.SetValueOrError(ve2.Clone())
		return
	}

	err := vh.SetValue(functools.RunCombine(fnValue, ve1.GetValue(), ve2.GetValue()))
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenCombineAsync(
	other CompletionStage,
	combineFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(combineFunc)
	if err := functools.CheckCombineFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))

	octx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
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

		err := vh.SetValue(functools.RunCombine(fnValue, ve1.GetValue(), ve2.GetValue()))
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenAcceptBoth(other CompletionStage, acceptFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptBothFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)

	ve1, ve2 := cf.v.BothValue(ocf.v, cf.ctx)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}

	if !ve2.HaveValue() {
		vh.SetValueOrError(ve2.Clone())
		return
	}

	functools.RunAcceptBoth(fnValue, ve1.GetValue(), ve2.GetValue())
	vh.SetValue(functools.NilValue)
	return
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenAcceptBothAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptBothFunction(fnValue.Type(), cf.vType, ocf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewAsyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve1, ve2 := cf.v.BothValue(ocf.v, cf.ctx)
		if !ve1.HaveValue() {
			vh.SetValueOrError(ve1.Clone())
			return
		}

		if !ve2.HaveValue() {
			vh.SetValueOrError(ve2.Clone())
			return
		}

		functools.RunAcceptBoth(fnValue, ve1.GetValue(), ve2.GetValue())
		vh.SetValue(functools.NilValue)
	})
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) RunAfterBoth(other CompletionStage, runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)
	ve1, ve2 := cf.v.BothValue(ocf.v, cf.ctx)
	if !ve1.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}
	if !ve2.HaveValue() {
		vh.SetValueOrError(ve1.Clone())
		return
	}
	runnable()
	vh.SetValue(functools.NilValue)

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) RunAfterBothAsync(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewAsyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve1, ve2 := cf.v.BothValue(ocf.v, cf.ctx)
		if !ve1.HaveValue() {
			vh.SetValueOrError(ve1.Clone())
			return
		}
		if !ve2.HaveValue() {
			vh.SetValueOrError(ve2.Clone())
			return
		}
		runnable()
		vh.SetValue(functools.NilValue)
	})
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ApplyToEither(other CompletionStage, applyFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := functools.CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(fnValue.Type().Out(0))
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)
	ve := cf.v.SelectValue(ocf.v, cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	err := vh.SetValue(functools.RunApply(fnValue, ve.GetValue()))
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ApplyToEitherAsync(
	other CompletionStage,
	applyFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := functools.CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		err := vh.SetValue(functools.RunApply(fnValue, ve.GetValue()))
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) AcceptEither(other CompletionStage, acceptFunc interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)
	ve := cf.v.SelectValue(ocf.v, cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	functools.RunAccept(fnValue, ve.GetValue())
	err := vh.SetValue(functools.NilValue)
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) AcceptEitherAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := functools.CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewAsyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		functools.RunAccept(fnValue, ve.GetValue())
		err := vh.SetValue(functools.NilValue)
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) RunAfterEither(other CompletionStage, runnable func()) (retCf CompletionStage) {
	defer cf.setDone()
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewSyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	defer handlePanic(vh)
	ve := cf.v.SelectValue(ocf.v, cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}

	runnable()
	err := vh.SetValue(functools.NilValue)
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) RunAfterEitherAsync(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) (retCf CompletionStage) {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	octx, _ := context.WithCancel(cf.ctx)

	vh := NewAsyncHandler(functools.NilType)
	retCf = newCfWithCancel(octx, func() {
		cf.cancelFunc()
		ocf.cancelFunc()
	}, vh)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.v.SelectValue(ocf.v, cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}

		runnable()
		err := vh.SetValue(functools.NilValue)
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenCompose(f interface{}) (retCf CompletionStage) {
	defer cf.setDone()
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := checkComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	if !ve.HaveValue() {
		vh.SetValueOrError(ve.Clone())
		return
	}
	newCom := functools.RunCompose(fnValue, ve.GetValue())
	if newCom.IsValid() {
		i := newCom.Interface()
		if i == nil {
			panic("Return CompletionStage is nil. ")
		}
		return i.(CompletionStage)
		//err := ncf.v.SetValue(ve.GetValue())
		//if err != nil {
		//	ncf.v.SetPanic(err)
		//}
		//return ncf
	}
	return
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) ThenComposeAsync(f interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := checkComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(composeCfType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer cf.setDone()
		ve := cf.getValue(cf.ctx)
		if !ve.HaveValue() {
			vh.SetValueOrError(ve.Clone())
			return
		}
		newCom := functools.RunCompose(fnValue, ve.GetValue())
		if newCom.IsValid() {
			i := newCom.Interface()
			if i == nil {
				vh.SetPanic(errors.New("Return CompletionStage is nil. "))
				return
			}
			vh.SetValue(reflect.ValueOf(&composeCf{joinVe: i.(Joinable)}))
		} else {
			vh.SetPanic(errors.New("Return CompletionStage is nil. "))
		}
	})
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

func (cf *defaultCompletableFuture) JoinCompletionStage(ctx context.Context) CompletionStage {
	return cf
}

// 尝试获得ValueOrError
// 此处还负责处理ComposeAsync封装的CompletableFuture，该设计可能不那么“优雅”
func (cf *defaultCompletableFuture) getValue(ctx context.Context) ValueOrError {
	ve := cf.v.Get(ctx)
	if ve.GetError() == nil {
		v := ve.GetValue()
		if v.IsValid() && !v.IsZero() {
			if c, ok := v.Interface().(*composeCf); ok {
				ve = c.joinVe.JoinCompletionStage(ctx).(*defaultCompletableFuture).getValue(ctx)
			}
		}
	}
	return ve
}

func (cf *defaultCompletableFuture) getValueAndCache(ctx context.Context) ValueOrError {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	if cf.result != nil {
		return cf.result
	}

	cf.result = cf.getValue(ctx)
	return cf.result
}

// 捕获阶段异常，返回补偿结果
// Param：f func(o interface{}) TYPE参数函数，参数：捕获的panic参数，返回补偿的结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) Exceptionally(f interface{}) (retCf CompletionStage) {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckPanicFunction(fnValue.Type()); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)
	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	if ve.HaveValue() {
		err := vh.SetValue(ve.GetValue())
		if err != nil {
			vh.SetPanic(err)
		}
		return
	}
	if ve.HavePanic() {
		p := ve.GetPanic()
		if p != nil {
			err := vh.SetValue(functools.RunPanic(fnValue, reflect.ValueOf(p)))
			if err != nil {
				vh.SetPanic(err)
			}
		}
	}

	return
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) WhenComplete(f interface{}) (retCf CompletionStage) {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckWhenCompleteFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	v := ve.GetValue()
	if !v.IsValid() {
		v = reflect.New(cf.vType).Elem()
	}
	p := ve.GetPanic()
	var panicV reflect.Value
	if p == nil {
		panicV = reflect.Zero(functools.InterfaceType)
	} else {
		panicV = reflect.ValueOf(p)
	}
	functools.RunWhenComplete(fnValue, v, panicV)

	vh.SetValue(functools.NilValue)
	return
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) WhenCompleteAsync(f interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckWhenCompleteFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}
	vh := NewAsyncHandler(functools.NilType)
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)

		ve := cf.getValue(cf.ctx)
		v := ve.GetValue()
		if !v.IsValid() {
			v = reflect.New(cf.vType).Elem()
		}
		p := ve.GetPanic()
		var panicV reflect.Value
		if p == nil {
			panicV = reflect.Zero(functools.InterfaceType)
		} else {
			panicV = reflect.ValueOf(p)
		}
		functools.RunWhenComplete(fnValue, v, panicV)
		vh.SetValue(functools.NilValue)
	})
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) Handle(f interface{}) (retCf CompletionStage) {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckHandleFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	defer handlePanic(vh)
	ve := cf.getValue(cf.ctx)
	v := ve.GetValue()
	if !v.IsValid() {
		v = reflect.New(cf.vType).Elem()
	}
	p := ve.GetPanic()

	var panicV reflect.Value
	if p == nil {
		panicV = reflect.Zero(functools.InterfaceType)
	} else {
		panicV = reflect.ValueOf(p)
	}
	err := vh.SetValue(functools.RunHandle(fnValue, v, panicV))
	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *defaultCompletableFuture) HandleAsync(f interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	cf.checkValue()
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckHandleFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	ctx, _ := context.WithCancel(cf.ctx)
	retCf = newCfWithCancel(ctx, cf.cancelFunc, vh)

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		ve := cf.getValue(cf.ctx)
		v := ve.GetValue()
		if !v.IsValid() {
			v = reflect.New(cf.vType).Elem()
		}
		p := ve.GetPanic()

		var panicV reflect.Value
		if p == nil {
			panicV = reflect.Zero(functools.InterfaceType)
		} else {
			panicV = reflect.ValueOf(p)
		}
		err := vh.SetValue(functools.RunHandle(fnValue, v, panicV))
		if err != nil {
			vh.SetPanic(err)
		}
	})

	if err != nil {
		vh.SetPanic(err)
	}

	return
}

// 给予get的值并正常结束
func (cf *defaultCompletableFuture) Complete(v interface{}) error {
	return cf.v.SetValue(reflect.ValueOf(v))
}

// 发送panic，异常结束
func (cf *defaultCompletableFuture) CompleteExceptionally(v interface{}) {
	cf.v.SetPanic(v)
}

// 取消并打断stage链，退出任务
// 如果任务已完成返回false，成功取消返回true
func (cf *defaultCompletableFuture) Cancel() bool {
	if cf.cancelFunc != nil {
		cf.setCancel()
		cf.cancelFunc()
		return cf.IsCancelled()
	}
	return false
}

// 是否在完成前被取消
func (cf *defaultCompletableFuture) IsCancelled() bool {
	if atomic.LoadInt32(&cf.status) == completableFutureCancel {
		return true
	} else {
		select {
		case <-cf.ctx.Done():
			return true
		default:
			return false
		}
	}
}

// 是否任务完成
// 当任务正常完成，被取消，抛出异常都会返回true
func (cf *defaultCompletableFuture) IsDone() bool {
	return atomic.LoadInt32(&cf.status) != completableFutureNone
}

// 等待并获得任务执行结果
// Param： result 目标结果，必须为同类型的指针
// Param： timeout 等待超时时间，如果不传值则一直等待
func (cf *defaultCompletableFuture) Get(result interface{}, timeout ...time.Duration) error {
	defer cf.setDone()

	cf.checkValue()
	var ve ValueOrError
	if len(timeout) > 0 {
		ctx, _ := context.WithTimeout(cf.ctx, timeout[0])
		ve = cf.getValueAndCache(ctx)
	} else {
		ve = cf.getValueAndCache(cf.ctx)
	}
	if ve.HavePanic() {
		panic(ve.GetPanic())
	}
	if ve.IsDone() {
		return errors.New("cancelled. ")
	}
	if result == nil {
		return nil
	}
	retValue := reflect.ValueOf(result)
	if err := functools.CheckPtr(retValue.Type()); err != nil {
		return err
	}
	err := ve.GetError()
	if err != nil {
		return err
	}
	v := ve.GetValue()
	if v.IsValid() {
		if v.Type() == functools.NilType {
			return errors.New("Nil type cannot be set! ")
		}
		retValue = retValue.Elem()
		if !retValue.CanSet() {
			return errors.New("Cannot set. ")
		}
		retValue.Set(v)
	}

	return nil
}

func (cf *defaultCompletableFuture) setDone() bool {
	return atomic.CompareAndSwapInt32(&cf.status, completableFutureNone, completableFutureDone)
}

func (cf *defaultCompletableFuture) setCancel() bool {
	return atomic.CompareAndSwapInt32(&cf.status, completableFutureNone, completableFutureCancel)
}

func (cf *defaultCompletableFuture) checkValue() {
	if cf.v == nil {
		panic("Without value, cannot be here")
	}
}

func (cf *defaultCompletableFuture) checkSameType(other *defaultCompletableFuture) {
	if cf.vType != other.vType {
		panic("Not same type!")
	}
}

func handlePanic(handler *defaultValueHandler) {
	if r := recover(); r != nil {
		handler.SetPanic(r)
	}
}

func convert(stage CompletionStage) *defaultCompletableFuture {
	if v, ok := stage.(*defaultCompletableFuture); ok {
		return v
	}
	panic("CompletionStage is not *defaultCompletableFuture")
}

func (cf *defaultCompletableFuture) chooseExecutor(executor ...executor.Executor) executor.Executor {
	if len(executor) == 0 {
		return defaultExecutor
	} else {
		return executor[0]
	}
}

func chooseExecutor(executor ...executor.Executor) executor.Executor {
	if len(executor) == 0 {
		return defaultExecutor
	} else {
		return executor[0]
	}
}

type composeCf struct {
	joinVe Joinable
}

var composeCfType = reflect.TypeOf((*composeCf)(nil))

func CompletedFuture(value interface{}) (retCf CompletionStage) {
	var v reflect.Value
	var t reflect.Type
	if value == nil {
		v = functools.NilValue
		t = functools.NilType
	} else {
		v = reflect.ValueOf(value)
		t = v.Type()
	}

	vh := NewSyncHandler(t)
	ctx, cancel := context.WithCancel(context.Background())
	retCf = newCfWithCancel(ctx, cancel, vh)

	defer retCf.(*defaultCompletableFuture).setDone()

	err := vh.SetValue(v)
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

func SupplyAsync(f interface{}, executor ...executor.Executor) (retCf CompletionStage) {
	fnValue := reflect.ValueOf(f)
	if err := functools.CheckSupplyFunction(fnValue.Type()); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	ctx, cancel := context.WithCancel(context.Background())
	retCf = newCfWithCancel(ctx, cancel, vh)

	exec := chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer retCf.(*defaultCompletableFuture).setDone()

		v := functools.RunSupply(fnValue)
		err := vh.SetValue(v)
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		panic(err)
	}
	return
}

func RunAsync(f func(), executor ...executor.Executor) (retCf CompletionStage) {
	vh := NewAsyncHandler(functools.NilType)
	ctx, cancel := context.WithCancel(context.Background())
	retCf = newCfWithCancel(ctx, cancel, vh)

	exec := chooseExecutor(executor...)
	err := exec.Run(func() {
		defer handlePanic(vh)
		defer retCf.(*defaultCompletableFuture).setDone()
		f()
		err := vh.SetValue(functools.NilValue)
		if err != nil {
			vh.SetPanic(err)
		}
	})
	if err != nil {
		panic(err)
	}
	return
}

func AllOf(cfs ...CompletionStage) (retCf CompletionStage) {
	vh := NewSyncHandler(functools.NilType)
	cancellers := make([]context.CancelFunc, 0, len(cfs))
	vhs := make([]ValueHandler, 0, len(cfs))
	for _, cf := range cfs {
		vhs = append(vhs, cf.(*defaultCompletableFuture).v)
		cancellers = append(cancellers, cf.(*defaultCompletableFuture).cancelFunc)
	}
	ctx, cancel := context.WithCancel(context.Background())
	retCf = newCfWithCancel(ctx, func() {
		cancel()
		for _, cancelFunc := range cancellers {
			cancelFunc()
		}
	}, vh)

	rets := AllOfValue(ctx, vhs...)
	for _, v := range rets {
		if v.HavePanic() {
			panic(v.GetPanic())
		}
	}
	err := vh.SetValue(functools.NilValue)
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

func AnyOf(cfs ...CompletionStage) (retCf CompletionStage) {
	vh := NewSyncHandler(functools.NilType)
	vhs := make([]ValueHandler, 0, len(cfs))
	for _, cf := range cfs {
		vhs = append(vhs, cf.(*defaultCompletableFuture).v)
	}
	ctx, cancel := context.WithCancel(context.Background())
	retCf = newCfWithCancel(ctx, func() {
		cancel()
	}, vh)

	_, ve := AnyOfValue(ctx, vhs...)
	if ve.HavePanic() {
		panic(ve.GetPanic())
	}
	err := vh.SetValue(functools.NilValue)
	if err != nil {
		vh.SetPanic(err)
	}
	return
}

var completionStageType = reflect.TypeOf((*CompletionStage)(nil)).Elem()

type Joinable interface {
	JoinCompletionStage(ctx context.Context) CompletionStage
}

func checkComposeFunction(fn reflect.Type, vType reflect.Type) error {
	if fn.Kind() != reflect.Func {
		return errors.New("Param is not a function. ")
	}
	if fn.NumIn() != 1 || fn.NumOut() != 1 {
		return errors.New("Type must be f func(o TYPE) CompletionStage. number not match. ")
	}
	inType := fn.In(0)
	if inType != vType {
		return errors.New("Type must be f func(o TYPE) CompletionStage. in[0] not match. ")
	}

	outType := fn.Out(0)
	if outType != completionStageType && !outType.Implements(completionStageType) {
		return errors.New("Type must be f func(o TYPE) CompletionStage. out[0] not match. ")
	}
	return nil
}

type UnlimitedExecutor struct{}

func (ue UnlimitedExecutor) Run(task executor.Task) error {
	go task()
	return nil
}

func (ue UnlimitedExecutor) Stop() {

}
