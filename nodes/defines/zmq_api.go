package defines

import (
	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type NewMasterNodeCaller string

type NewMasterNodeClientAPI func(args Values) (ret Values, err error)
type NewMasterNodeServerAPI func(caller NewMasterNodeCaller, args Values) (ret Values, err error)

type NewMasterNodeAPIClient interface {
	CallOmitResponse(api string, args Values)
	CallWithResponse(api string, args Values) *async_wrapper.AsyncWrapper[Values]
	ExposeAPI(apiName string, api NewMasterNodeClientAPI, newGoroutine bool)
	can_close.CanClose
}

type NewMasterNodeAPIServer interface {
	ExposeAPI(apiName string, api NewMasterNodeServerAPI, newGoroutine bool)
	ConcealAPI(apiName string)
	CallOmitResponse(callee NewMasterNodeCaller, api string, args Values)
	CallWithResponse(callee NewMasterNodeCaller, api string, args Values) *async_wrapper.AsyncWrapper[Values]
	SetOnCloseCleanUp(callee NewMasterNodeCaller, cb func())
	can_close.CanClose
}
