package async_wrapper

import (
	"context"
	"time"
)

type NoRetAsyncWrapper struct {
	controller       *NoRetAsyncController
	doFunc           func(*NoRetAsyncController)
	doInNewGoRoutine bool
	launched         bool
}

func (w *NoRetAsyncWrapper) SetContext(ctx context.Context) *NoRetAsyncWrapper {
	if w.launched {
		panic("set ctx after async func launched")
	} else {
		w.controller.c = ctx
	}
	return w
}

func (w *NoRetAsyncWrapper) SetTimeout(timeout time.Duration) *NoRetAsyncWrapper {
	if w.launched {
		panic("set ctx after async func launched")
	} else {
		ctx, _ := context.WithTimeout(w.controller.c, timeout)
		w.controller.c = ctx
	}
	return w
}

func (w *NoRetAsyncWrapper) AsyncOmitResult() {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	go w.do()
}

func (w *NoRetAsyncWrapper) AsyncGetResult(callback func(error)) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	go func() {
		w.do()
		callback(w.controller.err)
	}()
}

func (w *NoRetAsyncWrapper) BlockGetResult() (err error) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	w.do()
	return w.controller.err
}

func (w *NoRetAsyncWrapper) RedirectResult(reciver chan error, block bool) {
	if w.launched {
		panic("double launch async func")
	}
	w.launched = true
	launch := func() {
		w.do()
		reciver <- w.controller.err
	}
	if block {
		launch()
	} else {
		go launch()
	}
}

// will block
func (w *NoRetAsyncWrapper) do() {
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

type NoRetAsyncController struct {
	c context.Context
	w chan struct{}
	// 0: waiting
	// 1: ret set
	// 2: err set
	status     int
	err        error
	cancelHook func()
}

// add cancel hook
func (a *NoRetAsyncController) SetCancelHook(hook func()) {
	a.cancelHook = hook
}

// make context readonly
func (a *NoRetAsyncController) Context() context.Context {
	return a.c
}

func (a *NoRetAsyncController) SetOK() {
	// when ret not set, record
	if a.status == 0 {
		a.status = 1
		close(a.w)
	}
}

func (a *NoRetAsyncController) SetErr(err error) {
	if err == nil {
		a.SetOK()
	} else {
		a.setErr(err)
	}
}

func (a *NoRetAsyncController) setErr(err error) {
	// when ret not set/an error is set, record
	if a.status == 0 {
		a.err = err
		a.status = 2
		close(a.w)
	}
}

func NewNoRetAsyncWrapper[T any](doFunc func(*NoRetAsyncController), runInGotoutine bool) *NoRetAsyncWrapper {
	ctx := context.Background()
	controller := &NoRetAsyncController{
		w:      make(chan struct{}),
		c:      ctx,
		status: 0,
	}
	wrapper := &NoRetAsyncWrapper{
		controller:       controller,
		doFunc:           doFunc,
		doInNewGoRoutine: runInGotoutine,
		launched:         false,
	}
	// no guarantee when it will run
	// runtime.SetFinalizer(wrapper, func(*NoRetAsyncWrapper) {
	// 	wrapper.onGCCheck()
	// })
	return wrapper
}
