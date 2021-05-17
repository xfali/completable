// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package queued

import (
	"container/list"
	"context"
	"fmt"
	"github.com/xfali/completable"
	"github.com/xfali/executor"
	"sync"
	"time"
)

type Type int

const (
	// unknown
	TypeUnknown = iota

	// 当阶段正常完成时执行参数函数
	TypeThenApply

	// 当阶段正常完成时执行参数函数
	TypeThenApplyAsync

	// 当阶段正常完成时执行参数函数：结果消耗
	TypeThenAccept

	// 当阶段正常完成时执行参数函数：结果消耗
	TypeThenAcceptAsync

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	TypeThenRun

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	TypeThenRunAsync

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	TypeThenCombine

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	TypeThenCombineAsync

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	TypeThenAcceptBoth

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	TypeThenAcceptBothAsync

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	TypeRunAfterBoth

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	TypeRunAfterBothAsync

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	TypeApplyToEither

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	TypeApplyToEitherAsync

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	TypeAcceptEither

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	TypeAcceptEitherAsync

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	TypeRunAfterEither

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	TypeRunAfterEitherAsync

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	TypeThenCompose

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	TypeThenComposeAsync

	// 捕获阶段异常，返回补偿结果
	TypeExceptionally

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	TypeWhenComplete

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	TypeWhenCompleteAsync

	// 阶段执行时获得结果或者panic,并转化结果
	TypeHandle

	// 阶段执行时获得结果或者panic,并转化结果
	TypeHandleAsync
)

type stage struct {
	other    completable.CompletionStage
	value    interface{}
	fn       interface{}
	executor executor.Executor
	cfType   Type
}

type queuedCompletableFuture struct {
	origin completable.CompletionStage

	queueLocker sync.Mutex
	queue       *list.List

	once sync.Once
}

func (cf *queuedCompletableFuture) enqueue(stage *stage) {
	cf.queueLocker.Lock()
	defer cf.queueLocker.Unlock()

	cf.queue.PushBack(stage)
}

// 当阶段正常完成时执行参数函数：进行类型变换
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenApply(applyFunc interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeThenApply,
		fn:     applyFunc,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数
// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeThenApplyAsync,
		fn:       applyFunc,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenAccept(acceptFunc interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeThenAccept,
		fn:     acceptFunc,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结果消耗
// Param：参数函数：f func(o TYPE)参数为上阶段结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeThenAcceptAsync,
		fn:       acceptFunc,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenRun(runnable interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeThenRun,
		fn:     runnable,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：不关心上一步结果
// Param：参数函数: f func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenRunAsync(runnable interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeThenRunAsync,
		fn:       runnable,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenCombine(other completable.CompletionStage, combineFunc interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeThenCombine,
		fn:     combineFunc,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
// Param：other，当该CompletionStage也返回后进行结合转化
// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenCombineAsync(other completable.CompletionStage, combineFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeThenCombineAsync,
		fn:       combineFunc,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenAcceptBoth(other completable.CompletionStage, acceptFunc interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeThenAcceptBoth,
		fn:     acceptFunc,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
// Param：other，当该CompletionStage也返回后进行消耗
// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenAcceptBothAsync(
	other completable.CompletionStage, acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeThenAcceptBothAsync,
		fn:       acceptFunc,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) RunAfterBoth(other completable.CompletionStage, runnable interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeRunAfterBoth,
		fn:     runnable,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
// Param：other，当该CompletionStage也完成后执行参数函数
// Param：参数函数 runnable func()
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) RunAfterBothAsync(other completable.CompletionStage, runnable interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeRunAfterBothAsync,
		fn:       runnable,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ApplyToEither(other completable.CompletionStage, applyFunc interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeApplyToEither,
		fn:     applyFunc,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ApplyToEitherAsync(other completable.CompletionStage, applyFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeApplyToEitherAsync,
		fn:       applyFunc,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) AcceptEither(other completable.CompletionStage, acceptFunc interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeAcceptEither,
		fn:     acceptFunc,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) AcceptEitherAsync(other completable.CompletionStage, acceptFunc interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeAcceptEitherAsync,
		fn:       acceptFunc,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) RunAfterEither(other completable.CompletionStage, runnable interface{}) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType: TypeRunAfterEither,
		fn:     runnable,
		other:  other,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
// Param：参数函数
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) RunAfterEitherAsync(other completable.CompletionStage, runnable interface{}, executor ...executor.Executor) completable.CompletionStage {
	cf.checkType(other)
	stage := &stage{
		cfType:   TypeRunAfterEitherAsync,
		fn:       runnable,
		other:    other,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) completable.CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenCompose(f interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeThenCompose,
		fn:     f,
	}
	cf.enqueue(stage)
	return cf
}

// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
// Param：参数函数，f func(o TYPE) completable.CompletionStage 参数：上一阶段结果，返回新的CompletionStage
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) ThenComposeAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeThenComposeAsync,
		fn:       f,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 捕获阶段异常，返回补偿结果
// Param：f func(o interface{}) TYPE参数函数，参数：捕获的panic参数，返回补偿的结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) Exceptionally(f interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeExceptionally,
		fn:     f,
	}
	cf.enqueue(stage)
	return cf
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) WhenComplete(f interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeWhenComplete,
		fn:     f,
	}
	cf.enqueue(stage)
	return cf
}

// 阶段执行时获得结果或者panic,注意会继续传递panic
// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) WhenCompleteAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeWhenCompleteAsync,
		fn:       f,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) Handle(f interface{}) completable.CompletionStage {
	stage := &stage{
		cfType: TypeHandle,
		fn:     f,
	}
	cf.enqueue(stage)
	return cf
}

// 阶段执行时获得结果或者panic,并转化结果
// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
// Return：新的CompletionStage
func (cf *queuedCompletableFuture) HandleAsync(f interface{}, executor ...executor.Executor) completable.CompletionStage {
	stage := &stage{
		cfType:   TypeHandleAsync,
		fn:       f,
		executor: cf.chooseExecutor(executor...),
	}
	cf.enqueue(stage)
	return cf
}

// 给予get的值并正常结束
func (cf *queuedCompletableFuture) Complete(v interface{}) error {
	return cf.join().Complete(v)
}

// 发送panic，异常结束
func (cf *queuedCompletableFuture) CompleteExceptionally(v interface{}) {
	cf.join().CompleteExceptionally(v)
}

// 取消并打断stage链，退出任务
// 如果任务已完成返回false，成功取消返回true
func (cf *queuedCompletableFuture) Cancel() bool {
	return cf.join().Cancel()
}

// 是否在完成前被取消
func (cf *queuedCompletableFuture) IsCancelled() bool {
	return cf.join().Cancel()
}

// 是否任务完成
// 当任务正常完成，被取消，抛出异常都会返回true
func (cf *queuedCompletableFuture) IsDone() bool {
	return cf.join().IsDone()
}

// 等待并获得任务执行结果
// Param： result 目标结果，必须为同类型的指针
// Param： timeout 等待超时时间，如果不传值则一直等待
func (cf *queuedCompletableFuture) Get(result interface{}, timeout ...time.Duration) error {
	return cf.join().Get(result, timeout...)
}

func CompletedFuture(value interface{}) (retCf completable.CompletionStage) {
	ret := &queuedCompletableFuture{
		origin: completable.CompletedFuture(value),
		queue:  list.New(),
	}
	return ret
}

func SupplyAsync(f interface{}, executor ...executor.Executor) (retCf completable.CompletionStage) {
	ret := &queuedCompletableFuture{
		origin: completable.SupplyAsync(f, executor...),
		queue:  list.New(),
	}
	return ret
}

func RunAsync(f func(), executor ...executor.Executor) (retCf completable.CompletionStage) {
	ret := &queuedCompletableFuture{
		origin: completable.RunAsync(f, executor...),
		queue:  list.New(),
	}
	return ret
}

func AllOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	if len(cfs) == 0 {
		return nil
	}

	origins := make([]completable.CompletionStage, len(cfs))
	for i := range cfs {
		if v, ok := cfs[i].(*queuedCompletableFuture); ok {
			origins[i] = v.join()
		} else {
			origins[i] = cfs[i]
		}
	}

	ret := &queuedCompletableFuture{
		origin: completable.AllOf(origins...),
		queue:  list.New(),
	}
	return ret
}

func AnyOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	if len(cfs) == 0 {
		return nil
	}
	cs, _ := completable.GetAny(context.Background(), cfs...)
	return cs
}

func (cf *queuedCompletableFuture) JoinCompletionStage(ctx context.Context) completable.CompletionStage {
	return cf.join()
}

func (cf *queuedCompletableFuture) join() completable.CompletionStage {
	cf.once.Do(func() {
		cf.origin = cf.convertOrigin()
	})
	return cf.origin
}

func (cf *queuedCompletableFuture) chooseExecutor(executor ...executor.Executor) executor.Executor {
	if len(executor) > 0 {
		return executor[0]
	}
	return nil
}

func (cf *queuedCompletableFuture) checkType(completable completable.CompletionStage) {
	checkType(completable)
}

func checkType(completable completable.CompletionStage) {
	if _, ok := completable.(*queuedCompletableFuture); !ok {
		panic("expect queuedCompletableFuture, not match")
	}
}

func toQueued(completable completable.CompletionStage) *queuedCompletableFuture {
	if v, ok := completable.(*queuedCompletableFuture); ok {
		return v
	}
	panic("expect queuedCompletableFuture, not match")
}

func runStage(completable completable.CompletionStage) completable.CompletionStage {
	if v, ok := completable.(*queuedCompletableFuture); ok {
		return v.convertOrigin()
	}
	panic("expect queuedCompletableFuture, not match")
}

func (cf *queuedCompletableFuture) convertOrigin() completable.CompletionStage {
	cf.queueLocker.Lock()
	defer cf.queueLocker.Unlock()

	cur := cf.origin
	for elem := cf.queue.Front(); elem != nil; elem = elem.Next() {
		stage := elem.Value.(*stage)
		switch stage.cfType {
		case TypeThenApply:
			cur = cur.ThenApply(stage.fn)
		case TypeThenApplyAsync:
			cur = cur.ThenApplyAsync(stage.fn, stage.executor)
		case TypeThenAccept:
			cur = cur.ThenAccept(stage.fn)
		case TypeThenAcceptAsync:
			cur = cur.ThenAcceptAsync(stage.fn, stage.executor)
		case TypeThenRun:
			cur = cur.ThenRun(stage.fn)
		case TypeThenRunAsync:
			cur = cur.ThenRunAsync(stage.fn, stage.executor)
		case TypeThenCombine:
			cur = cur.ThenCombine(runStage(stage.other), stage.fn)
		case TypeThenCombineAsync:
			cur = cur.ThenCombineAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeThenAcceptBoth:
			cur = cur.ThenAcceptBoth(runStage(stage.other), stage.fn)
		case TypeThenAcceptBothAsync:
			cur = cur.ThenAcceptBothAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeRunAfterBoth:
			cur = cur.RunAfterBoth(runStage(stage.other), stage.fn)
		case TypeRunAfterBothAsync:
			cur = cur.RunAfterBothAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeApplyToEither:
			cur = cur.ApplyToEither(runStage(stage.other), stage.fn)
		case TypeApplyToEitherAsync:
			cur = cur.ApplyToEitherAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeAcceptEither:
			cur = cur.AcceptEither(runStage(stage.other), stage.fn)
		case TypeAcceptEitherAsync:
			cur = cur.AcceptEitherAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeRunAfterEither:
			cur = cur.RunAfterEither(runStage(stage.other), stage.fn)
		case TypeRunAfterEitherAsync:
			cur = cur.RunAfterEitherAsync(runStage(stage.other), stage.fn, stage.executor)
		case TypeThenCompose:
			cur = cur.ThenCompose(stage.fn)
		case TypeThenComposeAsync:
			cur = cur.ThenComposeAsync(stage.fn, stage.executor)
		case TypeExceptionally:
			cur = cur.Exceptionally(stage.fn)
		case TypeWhenComplete:
			cur = cur.WhenComplete(stage.fn)
		case TypeWhenCompleteAsync:
			cur = cur.WhenCompleteAsync(stage.fn, stage.executor)
		case TypeHandle:
			cur = cur.Handle(stage.fn)
		case TypeHandleAsync:
			cur = cur.HandleAsync(stage.fn, stage.executor)
		default:
			panic(fmt.Sprintln("cannot handle type: ", stage.cfType))
		}
	}
	return cur
}
