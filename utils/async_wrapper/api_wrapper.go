package async_wrapper

import (
	"errors"

	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
)

// PackInstantAPI pack an api which can finish in very short time
// to callback format
func PackInstantAPI[inT any, outT any](
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
	return func(in inT, setResult func(outT, error)) {
		set := false
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
	InstantAPI(func(in inT) (outT, error))
	// set an api which could need a very long time to execute/finish
	BlockingAPI(func(in inT) (outT, error))
	// set an api which is in callback format with double result set check
	CallBackAPI(func(in inT, setResult func(outT, error)))
}

type AsyncAPIGroup[inT any, outT any] interface {
	// check api exist
	Exist(apiName string) bool
	// Add a new api
	AddAPI(apiName string) AsyncAPISetHandler[inT, outT]
	// Remove an api
	RemoveAPI(apiName string)
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

func (g *baseAPIGroup[inT, outT]) Exist(apiName string) bool {
	api, found := g.apis.Get(apiName)
	if !found || api == nil {
		return false
	} else {
		return true
	}
}

type apiSetter[inT any, outT any] struct {
	onSet          func(func(in inT, setResult func(outT, error)))
	checkDoubleSet bool
}

func NewApiSetter[inT any, outT any](cb func(func(in inT, setResult func(outT, error))), checkDoubleSet bool) AsyncAPISetHandler[inT, outT] {
	return &apiSetter[inT, outT]{
		onSet:          cb,
		checkDoubleSet: checkDoubleSet,
	}
}

// set an api which can finish in very short time
func (s *apiSetter[inT, outT]) InstantAPI(f func(in inT) (outT, error)) {
	s.onSet(PackInstantAPI(f))
}

// set an api which could need a very long time to execute/finish
func (s *apiSetter[inT, outT]) BlockingAPI(f func(in inT) (outT, error)) {
	s.onSet(PackBlockingAPI(f))
}

// set an api which is in callback format with double result set check
func (s *apiSetter[inT, outT]) CallBackAPI(f func(in inT, setResult func(outT, error))) {
	if s.checkDoubleSet {
		s.onSet(PackCallBackAPI(f))
	} else {
		s.onSet(f)
	}

}

func (g *baseAPIGroup[inT, outT]) AddAPI(apiName string) AsyncAPISetHandler[inT, outT] {
	return &apiSetter[inT, outT]{
		onSet: func(f func(in inT, setResult func(outT, error))) {
			g.apis.Set(apiName, f)
		},
		checkDoubleSet: true,
	}
}

func (g *baseAPIGroup[inT, outT]) RemoveAPI(apiName string) {
	g.apis.Delete(apiName)
}

func NewAsyncAPIGroup[inT any, outT any]() AsyncAPIGroup[inT, outT] {
	return &baseAPIGroup[inT, outT]{
		apis: sync_wrapper.NewSyncKVMap[string, func(in inT, setResult func(outT, error))](),
	}
}
