// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package lazycompletable

import (
	"context"
	"errors"
	"github.com/xfali/completable"
	"github.com/xfali/executor"
	"sync"
	"sync/atomic"
	"time"
)

type lazyCompletableFuture struct {
	origin atomic.Value
	fn     func(completable.CompletionStage) completable.CompletionStage
	header *lazyCompletableFuture
	next   *lazyCompletableFuture

	lock sync.Mutex
}

// 当阶段正常完成时执行参数函数：进行类型变换
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenApply(applyFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenApply(applyFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenApplyAsync(applyFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenAccept(acceptFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenAccept(acceptFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenAcceptAsync(acceptFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenRun(runnable func()) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenRun(runnable)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenRunAsync(runnable func(), executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenRunAsync(runnable, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenCombine(other completable.CompletionStage, combineFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ThenCombine(other, combineFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenCombineAsync(other completable.CompletionStage, combineFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ThenCombineAsync(other, combineFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenAcceptBoth(other completable.CompletionStage, acceptFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ThenAcceptBoth(other, acceptFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenAcceptBothAsync(
	other completable.CompletionStage, acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ThenAcceptBothAsync(other, acceptFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) RunAfterBoth(other completable.CompletionStage, runnable func()) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.RunAfterBoth(other, runnable)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) RunAfterBothAsync(other completable.CompletionStage, runnable func(), executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.RunAfterBothAsync(other, runnable, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ApplyToEither(other completable.CompletionStage, applyFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ApplyToEither(other, applyFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ApplyToEitherAsync(other completable.CompletionStage, applyFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.ApplyToEitherAsync(other, applyFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) AcceptEither(other completable.CompletionStage, acceptFunc interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.AcceptEither(other, acceptFunc)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) AcceptEitherAsync(other completable.CompletionStage, acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.AcceptEitherAsync(other, acceptFunc, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) RunAfterEither(other completable.CompletionStage, f func()) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.RunAfterEither(other, f)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) RunAfterEitherAsync(other completable.CompletionStage, f func(), executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			other = joinOriginCompletableStage(other)
			return origin.RunAfterEitherAsync(other, f, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenCompose(f interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenCompose(f)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) ThenComposeAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.ThenComposeAsync(f, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 捕获阶段异常，返回补偿结果
// Param：f func(o interface{}) TYPE参数函数，参数：捕获的panic参数，返回补偿的结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) Exceptionally(f interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.Exceptionally(f)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) WhenComplete(f interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.WhenComplete(f)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) WhenCompleteAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.WhenCompleteAsync(f, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) Handle(f interface{}) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.Handle(f)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *lazyCompletableFuture) HandleAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			return origin.HandleAsync(f, executor...)
		},
	}
	ret.header = cf.header
	cf.next = ret
	return ret
}

// 给予get的值并正常结束
func (cf *lazyCompletableFuture) Complete(v interface{}) error {
	for h := cf.header; h != nil; h = h.next {
		o := h.getOrigin()
		if o != nil {
			o.Complete(v)
		}
	}
	return nil
}

// 发送panic，异常结束
func (cf *lazyCompletableFuture) CompleteExceptionally(v interface{}) {
	for h := cf.header; h != nil; h = h.next {
		o := h.getOrigin()
		if o != nil {
			o.CompleteExceptionally(v)
			return
		}
	}
}

// 取消并打断stage链，退出任务
// 如果任务已完成返回false，成功取消返回true
func (cf *lazyCompletableFuture) Cancel() bool {
	for h := cf.header; h != nil; h = h.next {
		o := h.getOrigin()
		if o != nil {
			o.Cancel()
		}
	}
	return true
}

// 是否在完成前被取消
func (cf *lazyCompletableFuture) IsCancelled() bool {
	o := cf.getOrigin()
	if o == nil {
		return false
	} else {
		return o.IsCancelled()
	}
}

// 是否任务完成
// 当任务正常完成，被取消，抛出异常都会返回true
func (cf *lazyCompletableFuture) IsDone() bool {
	o := cf.getOrigin()
	if o == nil {
		return false
	} else {
		return o.IsDone()
	}
}

func (cf *lazyCompletableFuture) getOrigin() completable.CompletionStage {
	o := cf.origin.Load()
	if o == nil {
		return nil
	} else {
		return o.(completable.CompletionStage)
	}
}

func (cf *lazyCompletableFuture) resetOrigin(o completable.CompletionStage) {
	cf.origin.Store(o)
}

func (cf *lazyCompletableFuture) join() completable.CompletionStage {
	cf.lock.Lock()
	defer cf.lock.Unlock()

	if v := cf.getOrigin(); v != nil {
		return v
	}

	h := cf.header
	if h == nil {
		return nil
	}
	o := cf.header.getOrigin()
	for {
		o = getOriginCompletableStage(o)
		o = joinOriginCompletableStage(h.fn(o))
		h.resetOrigin(o)

		if h.next == nil {
			return o
		}
		h = h.next
	}
}

// 等待并获得任务执行结果
// Param： result 目标结果，必须为同类型的指针
// Param： timeout 等待超时时间，如果不传值则一直等待
func (cf *lazyCompletableFuture) Get(result interface{}, timeout ...time.Duration) error {
	o := cf.join()
	if o == nil {
		return errors.New("No origin CompletableFuture found. ")
	}
	return o.Get(result, timeout...)
}

func joinOriginCompletableStage(o completable.CompletionStage) completable.CompletionStage {
	if o == nil {
		return nil
	}
	for {
		if v, ok := o.(*lazyCompletableFuture); ok {
			o = v.join()
		} else {
			return o
		}
	}
}

func getOriginCompletableStage(o completable.CompletionStage) completable.CompletionStage {
	if o == nil {
		return nil
	}
	for {
		if v, ok := o.(*lazyCompletableFuture); ok {
			o = v.getOrigin()
		} else {
			return o
		}
	}
}

func CompletedFuture(value interface{}) (retCf completable.CompletionStage) {
	ret := &lazyCompletableFuture{
		fn: func(o completable.CompletionStage) completable.CompletionStage {
			return completable.CompletedFuture(value)
		},
	}
	ret.header = ret
	return ret
}

func SupplyAsync(f interface{}, executor ...executor.Executor) (retCf completable.CompletionStage) {
	ret := &lazyCompletableFuture{
		fn: func(o completable.CompletionStage) completable.CompletionStage {
			return completable.SupplyAsync(f, executor...)
		},
	}
	ret.header = ret
	return ret
}

func RunAsync(f func(), executor ...executor.Executor) (retCf completable.CompletionStage) {
	ret := &lazyCompletableFuture{
		fn: func(o completable.CompletionStage) completable.CompletionStage {
			return completable.RunAsync(f, executor...)
		},
	}
	ret.header = ret
	return ret
}

func AllOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			origins := make([]completable.CompletionStage, len(cfs))
			for i := range cfs {
				if v, ok := cfs[i].(*lazyCompletableFuture); ok {
					origins[i] = v.join()
				} else {
					origins[i] = cfs[i]
				}
			}
			return completable.AllOf(origins...)
		},
	}
	ret.header = ret
	return ret
}

func AnyOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	ret := &lazyCompletableFuture{
		fn: func(origin completable.CompletionStage) completable.CompletionStage {
			origins := make([]completable.CompletionStage, len(cfs))
			for i := range cfs {
				if v, ok := cfs[i].(*lazyCompletableFuture); ok {
					origins[i] = v.join()
				} else {
					origins[i] = cfs[i]
				}
			}
			return completable.AnyOf(origins...)
		},
	}
	ret.header = ret
	return ret
}

func (cf *lazyCompletableFuture) JoinCompletionStage(ctx context.Context) completable.CompletionStage {
	return cf.join()
}
