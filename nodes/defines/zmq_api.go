package defines

import (
	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type NewMasterNodeCaller string
type ArgWithCaller struct {
	Caller NewMasterNodeCaller
	Args   Values
}
type NewMasterNodeAPIClient interface {
	CallOmitResponse(api string, args Values)
	CallWithResponse(api string, args Values) async_wrapper.AsyncResult[Values]
	ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[Values, Values]
	can_close.CanClose
}

type NewMasterNodeAPIServer interface {
	ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[ArgWithCaller, Values]
	ConcealAPI(apiName string)
	CallOmitResponse(callee NewMasterNodeCaller, api string, args Values)
	CallWithResponse(callee NewMasterNodeCaller, api string, args Values) async_wrapper.AsyncResult[Values]
	SetOnCloseCleanUp(callee NewMasterNodeCaller, cb func())
	can_close.CanClose
}
