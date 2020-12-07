// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package CompletableFuture

import (
	"github.com/xfali/completable"
	"github.com/xfali/executor"
)

func CompletedFuture(value interface{}) (retCf completable.CompletionStage) {
	return completable.CompletedFuture(value)
}

func SupplyAsync(f interface{}, executor ...executor.Executor) (retCf completable.CompletionStage) {
	return completable.SupplyAsync(f, executor...)
}

func RunAsync(f func(), executor ...executor.Executor) (retCf completable.CompletionStage) {
	return completable.RunAsync(f, executor...)
}

func AllOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	return completable.AllOf(cfs...)
}

func AnyOf(cfs ...completable.CompletionStage) (retCf completable.CompletionStage) {
	return completable.AnyOf(cfs...)
}
