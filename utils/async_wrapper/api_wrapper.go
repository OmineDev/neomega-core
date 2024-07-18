package async_wrapper

import (
	"errors"

	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
)

// PackInstanceAPI pack an api which can finish in very short time
// to callback format
func PackInstanceAPI[inT any, outT any](
	inAPI func(in inT) (outT, error),
) func(in inT, setResult func(outT, error)) {
	return func(in inT, setResult func(outT, error)) {
		ret, err := inAPI(in)
		setResult(ret, err)
	}
}

// PackBlockingAPI pack an api which could need a very long time to execute
// to callback format with an new goroutine
func PackBlockingAPI[inT any, outT any](
	inAPI func(in inT) (outT, error),
) func(in inT, setResult func(outT, error)) {
	return func(in inT, setResult func(outT, error)) {
		go func() {
			ret, err := inAPI(in)
			setResult(ret, err)
		}()
	}
}

// PackBlockingAPI pack an api which is already a callback format
// to a callback format with double result set check
func PackCallBackAPI[inT any, outT any](
	inAPI func(in inT, setResult func(outT, error)),
) func(in inT, setResult func(outT, error)) {
	set := false
	return func(in inT, setResult func(outT, error)) {
		inAPI(in, func(ot outT, err error) {
			if !set {
				set = true
				setResult(ot, err)
			}
		})
	}
}

type AsyncAPISetHandler[inT any, outT any] interface {
	// set an api which can finish in very short time
	InstanceAPI(func(in inT) (outT, error))
	// set an api which could need a very long time to execute/finish
	BlockingAPI(func(in inT) (outT, error))
	// set an api which is in callback format with double result set check
	CallBackAPI(func(in inT, setResult func(outT, error)))
}

type AsyncAPIGroup[inT any, outT any] interface {
	// Add a new api
	AddAPI(apiName string) AsyncAPISetHandler[inT, outT]
	// call api without blocking
	// if apiName not exist, return api not exist error
	CallAPI(apiName string, in inT, onResult func(ret outT, err error))
}

type baseAPIGroup[inT any, outT any] struct {
	apis *sync_wrapper.SyncKVMap[string, func(in inT, setResult func(outT, error))]
}

var ErrAPINotExist = errors.New("api not exist")

func (g *baseAPIGroup[inT, outT]) CallAPI(apiName string, in inT, onResult func(ret outT, err error)) {
	api, found := g.apis.Get(apiName)
	if !found || api == nil {
		var empty outT
		onResult(empty, ErrAPINotExist)
	} else {
		api(in, onResult)
	}
}

type APISetter[inT any, outT any] struct {
	onSet func(func(in inT, setResult func(outT, error)))
}

// set an api which can finish in very short time
func (s *APISetter[inT, outT]) InstanceAPI(f func(in inT) (outT, error)) {
	s.onSet(PackInstanceAPI(f))
}

// set an api which could need a very long time to execute/finish
func (s *APISetter[inT, outT]) BlockingAPI(f func(in inT) (outT, error)) {
	s.onSet(PackBlockingAPI(f))
}

// set an api which is in callback format with double result set check
func (s *APISetter[inT, outT]) CallBackAPI(f func(in inT, setResult func(outT, error))) {
	s.onSet(PackCallBackAPI(f))
}

func (g *baseAPIGroup[inT, outT]) AddAPI(apiName string) AsyncAPISetHandler[inT, outT] {
	return &APISetter[inT, outT]{
		onSet: func(f func(in inT, setResult func(outT, error))) {
			g.apis.Set(apiName, f)
		},
	}
}

func NewAsyncAPIGroup[inT any, outT any]() AsyncAPIGroup[inT, outT] {
	return &baseAPIGroup[inT, outT]{
		apis: sync_wrapper.NewSyncKVMap[string, func(in inT, setResult func(outT, error))](),
	}
}
