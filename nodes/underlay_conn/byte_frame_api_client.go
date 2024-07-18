package underlay_conn

import (
	"context"
	"strings"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	conn_defines "github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/defines"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"

	"github.com/google/uuid"
)

type FrameAPIClient struct {
	can_close.CanCloseWithError
	FrameConn       conn_defines.ByteFrameConnBase
	cbs             *sync_wrapper.SyncKVMap[string, func(defines.Values, error)]
	nonBlockingApis async_wrapper.AsyncAPIGroup[defines.Values, defines.Values]
}

func NewFrameAPIClient(conn conn_defines.ByteFrameConnBase) *FrameAPIClient {
	c := &FrameAPIClient{
		// close underlay conn on err
		CanCloseWithError: can_close.NewClose(conn.Close),
		FrameConn:         conn,
		cbs:               sync_wrapper.NewSyncKVMap[string, func(defines.Values, error)](),
		nonBlockingApis:   async_wrapper.NewAsyncAPIGroup[defines.Values, defines.Values](),
	}
	go func() {
		// close when underlay err
		c.CloseWithError(<-conn.WaitClosed())
	}()
	return c
}

func NewFrameAPIClientWithCtx(conn conn_defines.ByteFrameConnBase, ctx context.Context) *FrameAPIClient {
	c := NewFrameAPIClient(conn)
	go func() {
		select {
		case <-c.WaitClosed():
		case <-ctx.Done():
			c.CloseWithError(ctx.Err())
		}
	}()
	return c
}

func (c *FrameAPIClient) ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[defines.Values, defines.Values] {
	if !strings.HasPrefix(apiName, "/") {
		apiName = "/" + apiName
	}
	return c.nonBlockingApis.AddAPI(apiName)
}

func (c *FrameAPIClient) Run() (err error) {
	go c.FrameConn.ReadRoutine(func(data []byte) {
		frames := bytesToBytesSlices(data)
		indexOrApi := string(frames[0])
		if strings.HasPrefix(indexOrApi, "/") {
			index := frames[1]
			c.nonBlockingApis.CallAPI(indexOrApi, frames[2:], func(z defines.Values, err error) {
				if len(index) == 0 {
					return
				}
				frames := append([][]byte{index}, defines.WrapError(z, err)...)
				c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
			})
		} else {
			if cb, ok := c.cbs.GetAndDelete(indexOrApi); ok {
				ret, err := defines.ValueWithErr(frames[1:]).Unwrap()
				cb(ret, err)
			}
		}
	})
	return <-c.WaitClosed()
}

func (c *FrameAPIClient) CallOmitResponse(api string, args defines.Values) {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	frames := append([][]byte{[]byte(api), {}}, args...)
	c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
}

func (c *FrameAPIClient) CallWithResponse(api string, args defines.Values) async_wrapper.AsyncResult[defines.Values] {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	idx := uuid.New().String()
	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[defines.Values]) {
		c.cbs.Set(idx, func(v defines.Values, err error) {
			ac.SetResultAndErr(v, err)
		})
		ac.SetCancelHook(func() {
			c.cbs.Delete(idx)
		})
		c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
	}, false)
}

// type clientRespHandler struct {
// 	idx    string
// 	frames [][]byte
// 	c      *FrameAPIClient
// 	ctx    context.Context
// }

// func (h *clientRespHandler) doSend() {
// 	h.c.FrameConn.WriteBytePacket(byteSlicesToBytes(h.frames))
// }

// func (h *clientRespHandler) SetContext(ctx context.Context) defines.NewMasterNodeResultHandler {
// 	h.ctx = ctx
// 	return h
// }

// func (h *clientRespHandler) SetTimeout(timeout time.Duration) defines.NewMasterNodeResultHandler {
// 	if h.ctx == nil {
// 		h.ctx = context.Background()
// 	}
// 	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
// 	return h
// }

// func (h *clientRespHandler) BlockGetResponse() defines.Values {
// 	resolver := make(chan defines.Values, 1)
// 	h.c.cbs.Set(h.idx, func(ret defines.Values) {
// 		resolver <- ret
// 	})
// 	h.doSend()
// 	if h.ctx == nil {
// 		return <-resolver
// 	}
// 	select {
// 	case ret := <-resolver:
// 		return ret
// 	case <-h.ctx.Done():
// 		h.c.cbs.Delete(h.idx)
// 		return defines.Empty
// 	}
// }

// func (h *clientRespHandler) AsyncGetResponse(callback func(defines.Values)) {
// 	if h.ctx == nil {
// 		h.c.cbs.Set(h.idx, callback)
// 	} else {
// 		resolver := make(chan defines.Values, 1)
// 		h.c.cbs.Set(h.idx, func(ret defines.Values) {
// 			resolver <- ret
// 		})
// 		go func() {
// 			select {
// 			case ret := <-resolver:
// 				callback(ret)
// 			case <-h.ctx.Done():
// 				h.c.cbs.Delete(h.idx)
// 				callback(defines.Empty)
// 				return
// 			}
// 		}()
// 	}
// 	h.doSend()
// }

// func (c *FrameAPIClient) CallWithResponse(api string, args defines.Values) defines.NewMasterNodeResultHandler {
// 	if !strings.HasPrefix(api, "/") {
// 		api = "/" + api
// 	}
// 	idx := uuid.New().String()
// 	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
// 	return &clientRespHandler{
// 		idx, frames, c, nil,
// 	}
// }
