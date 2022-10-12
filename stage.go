/*
 * Copyright 2022 Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package completable

import "github.com/xfali/executor"

type CompletionStage interface {
	// 当阶段正常完成时执行参数函数：进行类型变换
	// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
	// Return：新的CompletionStage
	ThenApply(applyFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数
	// Param：参数函数：f func(o TYPE1) TYPE2参数为上阶段结果，返回为处理后的返回值
	// Return：新的CompletionStage
	ThenApplyAsync(applyFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结果消耗
	// Param：参数函数：f func(o TYPE)参数为上阶段结果
	// Return：新的CompletionStage
	ThenAccept(acceptFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：结果消耗
	// Param：参数函数：f func(o TYPE)参数为上阶段结果
	// Return：新的CompletionStage
	ThenAcceptAsync(acceptFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	// Param：参数函数: f func()
	// Return：新的CompletionStage
	ThenRun(runnable interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：不关心上一步结果
	// Param：参数函数: f func()
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	ThenRunAsync(runnable interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	// Param：other，当该CompletionStage也返回后进行结合转化
	// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ThenCombine(other CompletionStage, combineFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，转化后返回
	// Param：other，当该CompletionStage也返回后进行结合转化
	// Param：参数函数，combineFunc func(TYPE1, TYPE2) TYPE3参数为两个CompletionStage的结果，返回转化结果
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	ThenCombineAsync(other CompletionStage, combineFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	// Param：other，当该CompletionStage也返回后进行消耗
	// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
	// Return：新的CompletionStage
	ThenAcceptBoth(other CompletionStage, acceptFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：结合两个CompletionStage的结果，进行消耗
	// Param：other，当该CompletionStage也返回后进行消耗
	// Param：参数函数，acceptFunc func(TYPE1, TYPE2) 参数为两个CompletionStage的结果
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	ThenAcceptBothAsync(other CompletionStage, acceptFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	// Param：other，当该CompletionStage也完成后执行参数函数
	// Param：参数函数 runnable func()
	// Return：新的CompletionStage
	RunAfterBoth(other CompletionStage, runnable interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage都完成后执行
	// Param：other，当该CompletionStage也完成后执行参数函数
	// Param：参数函数 runnable func()
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	RunAfterBothAsync(other CompletionStage, runnable interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
	// Return：新的CompletionStage
	ApplyToEither(other CompletionStage, applyFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行转化
	// Param：other，与该CompletionStage比较，用先完成的结果进行转化，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数 f func(o Type1) Type2参数为先完成的CompletionStage的结果，返回转化结果
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	ApplyToEitherAsync(other CompletionStage, applyFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数  f func(o Type)参数为先完成的CompletionStage的结果
	// Return：新的CompletionStage
	AcceptEither(other CompletionStage, acceptFunc interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage使用先完成的结果进行消耗
	// Param：other，与该CompletionStage比较，用先完成的结果进行消耗，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数 f func(o Type)参数为先完成的CompletionStage的结果
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	AcceptEitherAsync(other CompletionStage, acceptFunc interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数
	// Return：新的CompletionStage
	RunAfterEither(other CompletionStage, f interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：两个CompletionStage任意一个完成则执行操作
	// Param：other，与该CompletionStage比较，任意一个完成则执行操作，注意两个CompletionStage的返回结果类型必须相同
	// Param：参数函数
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	RunAfterEitherAsync(other CompletionStage, f interface{}, executor ...executor.Executor) CompletionStage

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
	// Return：新的CompletionStage
	ThenCompose(f interface{}) CompletionStage

	// 当阶段正常完成时执行参数函数：使用上一阶段结果转化为新的CompletionStage
	// Param：参数函数，f func(o TYPE) CompletionStage 参数：上一阶段结果，返回新的CompletionStage
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	ThenComposeAsync(f interface{}, executor ...executor.Executor) CompletionStage

	// 捕获阶段异常，返回补偿结果
	// Param：f func(o interface{}) TYPE参数函数，参数：捕获的panic参数，返回补偿的结果
	// Return：新的CompletionStage
	Exceptionally(f interface{}) CompletionStage

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
	// Return：新的CompletionStage
	WhenComplete(f interface{}) CompletionStage

	// 阶段执行时获得结果或者panic,注意会继续传递panic
	// Param：参数函数，f func(result Type, panic interface{}) 参数result：结果，参数panic：异常
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	WhenCompleteAsync(f interface{}, executor ...executor.Executor) CompletionStage

	// 阶段执行时获得结果或者panic,并转化结果
	// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
	// Return：新的CompletionStage
	Handle(f interface{}) CompletionStage

	// 阶段执行时获得结果或者panic,并转化结果
	// Param：参数函数，f func(result TYPE1, panic interface{}) TYPE2 参数result：结果，参数panic：异常，返回：转化的结果
	// Param：Executor: 异步执行的协程池，如果不填则使用内置默认协程池
	// Return：新的CompletionStage
	HandleAsync(f interface{}, executor ...executor.Executor) CompletionStage

	Completable

	Future
}
