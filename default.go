// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import (
	"context"
	"github.com/xfali/executor"
	"reflect"
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

type CompletableFuture struct {
	vType      reflect.Type
	v          ValueHandler
	ctx        context.Context
	cancelFunc context.CancelFunc
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
func (cf *CompletableFuture) ThenApply(applyFunc interface{}) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	ve := cf.v.Get(-1)
	if ve.GetError() != nil {
		vh.SetError(ve.GetError())
		return newCf(cf.ctx, vh)
	}

	err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev := cf.v.Get(-1)
		err := ev.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		err = vh.SetValue(RunApply(fnValue, ev.GetValue()))
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}

	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAccept(acceptFunc interface{}) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	ev := cf.v.Get(-1)
	err := ev.GetError()
	if err != nil {
		vh.SetError(err)
		return newCf(cf.ctx, vh)
	}

	RunAccept(fnValue, ev.GetValue())
	vh.SetValue(NilValue)
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev := cf.v.Get(-1)
		err := ev.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		RunAccept(fnValue, ev.GetValue())
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, nil)
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenRun(runnable func()) CompletionStage {
	cf.checkValue()

	vh := NewSyncHandler(NilType)
	ev := cf.v.Get(-1)
	err := ev.GetError()
	if err != nil {
		vh.SetError(err)
		return newCf(cf.ctx, vh)
	}

	runnable()
	vh.SetValue(NilValue)
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenRunAsync(runnable func(), executor ...executor.Executor) CompletionStage {
	cf.checkValue()

	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev := cf.v.Get(-1)
		err := ev.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		runnable()
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCombine(other CompletionStage, combineFunc interface{}) CompletionStage {
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
	ev1, ev2 := cf.v.BothValue(ocf.v, -1)
	err := ev1.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}

	err = ev2.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}

	err = vh.SetValue(RunCombine(fnValue, ev1.GetValue(), ev2.GetValue()))
	if err != nil {
		panic(err)
	}
	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCombineAsync(
	other CompletionStage,
	combineFunc interface{},
	executor ...executor.Executor) CompletionStage {
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

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev1, ev2 := cf.v.BothValue(ocf.v, -1)
		err := ev1.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}

		err = ev2.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}

		err = vh.SetValue(RunCombine(fnValue, ev1.GetValue(), ev2.GetValue()))
		if err != nil {
			vh.SetError(err)
		}
	})
	if err != nil {
		panic(err)
	}

	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptBoth(other CompletionStage, acceptFunc interface{}) CompletionStage {
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
	ev1, ev2 := cf.v.BothValue(ocf.v, -1)
	err := ev1.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}

	err = ev2.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}

	RunAcceptBoth(fnValue, ev1.GetValue(), ev2.GetValue())
	vh.SetValue(NilValue)
	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenAcceptBothAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) CompletionStage {
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
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev1, ev2 := cf.v.BothValue(ocf.v, -1)
		err := ev1.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}

		err = ev2.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}

		RunAcceptBoth(fnValue, ev1.GetValue(), ev2.GetValue())
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}

	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterBoth(other CompletionStage, runnable func()) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)

	vh := NewSyncHandler(NilType)
	ev1, ev2 := cf.v.BothValue(ocf.v, -1)
	err := ev1.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}
	err = ev2.GetError()
	if err != nil {
		vh.SetError(err)
		return newCfWithCancel(octx, func() {
			cancel()
			ocancel()
		}, vh)
	}
	runnable()
	vh.SetValue(NilValue)

	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterBothAsync(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()

	ctx, cancel := context.WithCancel(cf.ctx)
	octx, ocancel := context.WithCancel(ctx)
	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ev1, ev2 := cf.v.BothValue(ocf.v, -1)
		err := ev1.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		err = ev2.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		runnable()
		vh.SetValue(NilValue)
	})
	if err != nil {
		panic(err)
	}

	return newCfWithCancel(octx, func() {
		cancel()
		ocancel()
	}, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *CompletableFuture) ApplyToEither(other CompletionStage, applyFunc interface{}) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(fnValue.Type().Out(0))
	ve := cf.v.SelectValue(ocf.v, -1)
	if ve.GetError() != nil {
		vh.SetError(ve.GetError())
		return newCf(cf.ctx, vh)
	}

	err := vh.SetValue(RunApply(fnValue, ve.GetValue()))
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) ApplyToEitherAsync(
	other CompletionStage,
	applyFunc interface{},
	executor ...executor.Executor) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(applyFunc)
	if err := CheckApplyFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(fnValue.Type().Out(0))

	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ve := cf.v.SelectValue(ocf.v, -1)
		if ve.GetError() != nil {
			vh.SetError(ve.GetError())
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

	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
// Return：新的CompletionStage
func (cf *CompletableFuture) AcceptEither(other CompletionStage, acceptFunc interface{}) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	ve := cf.v.SelectValue(ocf.v, -1)
	if ve.GetError() != nil {
		vh.SetError(ve.GetError())
		return newCf(cf.ctx, vh)
	}

	RunAccept(fnValue, ve.GetValue())
	err := vh.SetValue(NilValue)
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) AcceptEitherAsync(
	other CompletionStage,
	acceptFunc interface{},
	executor ...executor.Executor) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	fnValue := reflect.ValueOf(acceptFunc)
	if err := CheckAcceptFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ve := cf.v.SelectValue(ocf.v, -1)
		if ve.GetError() != nil {
			vh.SetError(ve.GetError())
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
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterEither(other CompletionStage, runnable func()) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	vh := NewSyncHandler(NilType)
	ve := cf.v.SelectValue(ocf.v, -1)
	if ve.GetError() != nil {
		vh.SetError(ve.GetError())
		return newCf(cf.ctx, vh)
	}

	runnable()
	err := vh.SetValue(NilValue)
	if err != nil {
		panic(err)
	}
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) RunAfterAsyncEither(
	other CompletionStage,
	runnable func(),
	executor ...executor.Executor) CompletionStage {
	ocf := convert(other)
	cf.checkValue()
	ocf.checkValue()
	cf.checkSameType(ocf)

	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ve := cf.v.SelectValue(ocf.v, -1)
		if ve.GetError() != nil {
			vh.SetError(ve.GetError())
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
	return newCf(cf.ctx, vh)
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Return：新的CompletionStage
func (cf *CompletableFuture) ThenCompose(f interface{}) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := CheckComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewSyncHandler(NilType)
	ve := cf.v.Get(-1)
	err := ve.GetError()
	if err != nil {
		vh.SetError(err)
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
func (cf *CompletableFuture) ThenComposeAsync(f interface{}, executor ...executor.Executor) CompletionStage {
	cf.checkValue()

	fnValue := reflect.ValueOf(f)
	if err := CheckComposeFunction(fnValue.Type(), cf.vType); err != nil {
		panic(err)
	}

	vh := NewAsyncHandler(NilType)
	exec := cf.chooseExecutor(executor...)
	err := exec.Run(func() {
		ve := cf.v.Get(-1)
		err := ve.GetError()
		if err != nil {
			vh.SetError(err)
			return
		}
		newCom := RunCompose(fnValue, ve.GetValue())
		if newCom.IsValid() {
			i := newCom.Interface()
			if i == nil {
				panic("Return CompletionStage is nil. ")
			}
			return i.(*CompletableFuture)
		}
	})
	if err != nil {
		panic(err)
	}

	return nil
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
	return nil
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *CompletableFuture) HandleAsync(f interface{}, executor ...executor.Executor) CompletionStage {
	return nil
}

// 取消并打断stage链，退出任务
// 如果任务已完成返回false，成功取消返回true
func (cf *CompletableFuture) Cancel() bool {
	if cf.cancelFunc != nil {
		cf.cancelFunc()
	}
	return false
}

// 是否在完成前被取消
func (cf *CompletableFuture) IsCancelled() bool {
	return false
}

// 是否任务完成
// 当任务正常完成，被取消，抛出异常都会返回true
func (cf *CompletableFuture) IsDone() bool {
	return false
}

// 等待并获得任务执行结果
// Param： result 目标结果，必须为同类型的指针
// Param： timeout 等待超时时间，如果不传值则一直等待
func (cf *CompletableFuture) Get(result interface{}, timeout ...time.Duration) error {
	return nil
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
