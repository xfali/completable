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
