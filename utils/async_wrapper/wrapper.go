package async_wrapper

import (
	"context"
	"time"
)

type AsyncWrapper[T any] struct {
	controller       *AsyncController[T]
	doFunc           func(*AsyncController[T])
	doInNewGoRoutine bool
	launched         bool
}

func (w *AsyncWrapper[T]) SetContext(ctx context.Context) AsyncResult[T] {
	if w.launched {
		panic("set ctx after async func launched")
	} else {
		w.controller.c = ctx
	}
	return w
}

func (w *AsyncWrapper[T]) SetTimeout(timeout time.Duration) AsyncResult[T] {
	if w.launched {
		panic("set ctx after async func launched")
	} else {
		ctx, _ := context.WithTimeout(w.controller.c, timeout)
		w.controller.c = ctx
	}
	return w
}

func (w *AsyncWrapper[T]) AsyncGetResult(callback func(ret T, err error)) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	go func() {
		w.do()
		callback(w.controller.ret, w.controller.err)
	}()
}

func (w *AsyncWrapper[T]) BlockGetResult() (ret T, err error) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	w.do()
	return w.controller.ret, w.controller.err
}

func (w *AsyncWrapper[T]) RedirectResult(reciver chan struct {
	ret T
	err error
}, block bool) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	launch := func() {
		w.do()
		reciver <- struct {
			ret T
			err error
		}{
			ret: w.controller.ret,
			err: w.controller.err,
		}
	}
	if block {
		launch()
	} else {
		go launch()
	}
}

// func (w *AsyncWrapper[T]) onGCCheck() {
// 	if !w.launched {
// 		panic("async function created but not launched!")
// 	}
// }

// will block
func (w *AsyncWrapper[T]) do() {
	if w.doInNewGoRoutine {
		go w.doFunc(w.controller)
	} else {
		w.doFunc(w.controller)
	}
	select {
	case <-w.controller.w:
		break
	case <-w.controller.c.Done():
		w.controller.SetErr(w.controller.Context().Err())
		if w.controller.cancelHook != nil {
			w.controller.cancelHook()
		}
		break
	}
}

type AsyncController[T any] struct {
	c context.Context
	w chan struct{}
	// 0: waiting
	// 1: ret set
	// 2: err set
	status     int
	ret        T
	err        error
	cancelHook func()
}

// add cancel hook
func (a *AsyncController[T]) SetCancelHook(hook func()) {
	a.cancelHook = hook
}

// make context readonly
func (a *AsyncController[T]) Context() context.Context {
	return a.c
}

func (a *AsyncController[T]) SetResult(r T) {
	// when ret not set, record
	if a.status == 0 {
		a.ret = r
		a.status = 1
		close(a.w)
	}
}

func (a *AsyncController[T]) SetUnfinishedResult(r T) {
	// will not make async progress in finish status, but can set an result with error
	a.ret = r
}

func (a *AsyncController[T]) SetResultAndErr(r T, err error) {
	if err == nil {
		a.SetResult(r)
	} else {
		a.ret = r
		a.SetErr(err)
	}
}

func (a *AsyncController[T]) SetErr(err error) {
	// when ret not set/an error is set, record
	if a.status == 0 {
		a.err = err
		a.status = 2
		close(a.w)
	}
}

func NewAsyncWrapper[T any](doFunc func(*AsyncController[T]), runInGotoutine bool) AsyncResult[T] {
	ctx := context.Background()
	controller := &AsyncController[T]{
		w:      make(chan struct{}),
		c:      ctx,
		status: 0,
	}
	wrapper := &AsyncWrapper[T]{
		controller:       controller,
		doFunc:           doFunc,
		doInNewGoRoutine: runInGotoutine,
		launched:         false,
	}
	// no guarantee when it will run
	// runtime.SetFinalizer(wrapper, func(*AsyncWrapper[T]) {
	// 	wrapper.onGCCheck()
	// })
	return wrapper
}
