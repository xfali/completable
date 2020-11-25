// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

import "time"

type Future interface {
	// 取消并打断stage链，退出任务
	// 如果任务已完成返回false，成功取消返回true
	Cancel() bool

	// 是否在完成前被取消
	IsCancelled() bool

	// 是否任务完成
	// 当任务正常完成，被取消，抛出异常都会返回true
	IsDone() bool

	// 等待并获得任务执行结果
	// Param： result 目标结果，必须为同类型的指针
	// Param： timeout 等待超时时间，如果不传值则一直等待
	Get(result interface{}, timeout ...time.Duration) error
}
