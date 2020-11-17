// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package executor

import "github.com/xfali/executor"

type CompletionStage interface {
	// 当阶段正常完成时执行参数函数：进行类型变换
	// Param：参数函数：参数为上阶段结果，返回为处理后的返回值
	// Return：新的CompletionStage
	ThenApply(f func(o interface{}) interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数
	// Param：参数函数：参数为上阶段结果，返回为处理后的返回值
	// Return：新的CompletionStage
	ThenApplyAsync(f func(o interface{}) interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结果消耗
	// Param：参数函数：参数为上阶段结果
	// Return：新的CompletionStage
	ThenAccept(f func(o interface{})) CompletionStage

	// 当阶段正常完成时执行参数函数：结果消耗
	// Param：参数函数：参数为上阶段结果
	// Return：新的CompletionStage
	ThenAcceptAsync(f func(o interface{}), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	// Param：参数函数
	// Return：新的CompletionStage
	ThenRun(f func()) CompletionStage

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	// Param：参数函数
	// Return：新的CompletionStage
	ThenRunAsync(f func(), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	// Param：other，当该CompletionStage也返回后进行结合转化
	// Param：参数函数，参数为两个CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ThenCombine(other CompletionStage, combineFunc func(a, b interface{}) interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	// Param：other，当该CompletionStage也返回后进行结合转化
	// Param：参数函数，参数为两个CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ThenCombineAsync(other CompletionStage, combineFunc func(a, b interface{}) interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	// Param：other，当该CompletionStage也返回后进行消耗
	// Param：参数函数，参数为两个CompletionStage的结果
	// Return：新的CompletionStage
	ThenAcceptBoth(other CompletionStage, combineFunc func(a, b interface{})) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	// Param：other，当该CompletionStage也返回后进行消耗
	// Param：参数函数，参数为两个CompletionStage的结果
	// Return：新的CompletionStage
	ThenAcceptBothAsync(other CompletionStage, combineFunc func(a, b interface{}), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	// Param：other，当该CompletionStage也完成后执行参数函数
	// Param：参数函数
	// Return：新的CompletionStage
	RunAfterBoth(other CompletionStage, f func()) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	// Param：other，当该CompletionStage也完成后执行参数函数
	// Param：参数函数
	// Return：新的CompletionStage
	RunAfterBothAsync(other CompletionStage, f func(), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	// Param：other，与该CompletionStage比较，用先完成的结果进行转化
	// Param：参数函数 参数为先完成的CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ApplyToEither(other CompletionStage, f func(o interface{}) interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	// Param：other，与该CompletionStage比较，用先完成的结果进行转化
	// Param：参数函数 参数为先完成的CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ApplyToEitherAsync(other CompletionStage, f func(o interface{}) interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	// Param：other，与该CompletionStage比较，用先完成的结果进行消耗
	// Param：参数函数 参数为先完成的CompletionStage的结果
	// Return：新的CompletionStage
	AcceptEither(f func(o interface{})) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	// Param：other，与该CompletionStage比较，用先完成的结果进行消耗
	// Param：参数函数 参数为先完成的CompletionStage的结果
	// Return：新的CompletionStage
	AcceptEitherAsync(f func(o interface{}), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	// Param：other，与该CompletionStage比较，任意一个完成则执行操作
	// Param：参数函数
	// Return：新的CompletionStage
	RunAfterEither(other CompletionStage, f func()) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	// Param：other，与该CompletionStage比较，任意一个完成则执行操作
	// Param：参数函数
	// Return：新的CompletionStage
	RunAfterAsyncEither(other CompletionStage, f func(), executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	// Param：参数函数，参数：上一阶段结果，返回新的CompletionStage
	// Return：新的CompletionStage
	ThenCompose(f func(o interface{}) CompletionStage) CompletionStage

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	// Param：参数函数，参数：上一阶段结果，返回新的CompletionStage
	// Return：新的CompletionStage
	ThenComposeAsync(f func(o interface{}) CompletionStage) CompletionStage

	// 捕获阶段异常，返回补偿结果
	// Param：参数函数，参数：捕获的panic参数，返回补偿的结果
	// Return：新的CompletionStage
	Panic(f func(o interface{}) interface{}) CompletionStage

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	// Param：参数函数，参数result：结果，参数panic：异常
	// Return：新的CompletionStage
	WhenComplete(f func(result, panic interface{})) CompletionStage

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	// Param：参数函数，参数result：结果，参数panic：异常
	// Return：新的CompletionStage
	WhenCompleteAsync(f func(result, panic interface{}), executor ...executor.Executor) CompletionStage

	// 阶段执行时获得结果或者panic,并转化结果
	// Param：参数函数，参数result：结果，参数panic：异常，返回：转化的结果
	// Return：新的CompletionStage
	Handle(f func(result, panic interface{}) interface{}) CompletionStage

	// 阶段执行时获得结果或者panic,并转化结果
	// Param：参数函数，参数result：结果，参数panic：异常，返回：转化的结果
	// Return：新的CompletionStage
	HandleAsync(f func(result, panic interface{}) interface{}, executor ...executor.Executor) CompletionStage
}
