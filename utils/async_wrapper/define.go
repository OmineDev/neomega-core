package async_wrapper

import (
	"context"
	"time"
)

type AsyncResult[T any] interface {
	SetContext(ctx context.Context) AsyncResult[T]
	SetTimeout(timeout time.Duration) AsyncResult[T]
	AsyncGetResult(callback func(ret T, err error))
	BlockGetResult() (ret T, err error)
	RedirectResult(reciver chan struct {
		ret T
		err error
	}, block bool)
}
