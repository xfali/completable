// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package completable

type Completable interface {
	// 给予get的值并正常结束
	Complete(v interface{}) error

	// 发送panic，异常结束
	CompleteExceptionally(v interface{})
}
