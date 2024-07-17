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

type FrameAPIServer struct {
	nonBlockingApis *sync_wrapper.SyncKVMap[string, func(defines.NewMasterNodeCaller, defines.Values, func(defines.Values, error))]
	conns           *sync_wrapper.SyncKVMap[string, *FrameAPIServerConn]
	onClose         *sync_wrapper.SyncKVMap[string, func()]
	can_close.CanCloseWithError
}

func NewFrameAPIServer(onCloseHook func()) *FrameAPIServer {
	return &FrameAPIServer{
		nonBlockingApis:   sync_wrapper.NewSyncKVMap[string, func(defines.NewMasterNodeCaller, defines.Values, func(defines.Values, error))](),
		conns:             sync_wrapper.NewSyncKVMap[string, *FrameAPIServerConn](),
		onClose:           sync_wrapper.NewSyncKVMap[string, func()](),
		CanCloseWithError: can_close.NewClose(onCloseHook),
	}
}

type FrameAPIServerConn struct {
	identity      string
	identityBytes []byte
	can_close.CanCloseWithError
	*FrameAPIServer
	FrameConn conn_defines.ByteFrameConnBase
	cbs       *sync_wrapper.SyncKVMap[string, func(defines.Values, error)]
}

func (s *FrameAPIServer) NewFrameAPIServer(conn conn_defines.ByteFrameConnBase) *FrameAPIServerConn {
	identity := uuid.New().String()
	c := &FrameAPIServerConn{
		identity:      identity,
		identityBytes: []byte(identity),
		// close underlay conn on err
		CanCloseWithError: can_close.NewClose(conn.Close),
		FrameConn:         conn,
		cbs:               sync_wrapper.NewSyncKVMap[string, func(defines.Values, error)](),
		FrameAPIServer:    s,
	}
	s.conns.Set(identity, c)
	go func() {
		// close when underlay err
		c.CloseWithError(<-conn.WaitClosed())
		onClose, ok := s.onClose.Get(identity)
		if ok {
			onClose()
		}
		s.conns.Delete(identity)
	}()
	return c
}

func (c *FrameAPIServer) SetOnCloseCleanUp(callee defines.NewMasterNodeCaller, cb func()) {
	c.onClose.Set(string(callee), cb)
}

func (s *FrameAPIServer) NewFrameAPIServerWithCtx(conn conn_defines.ByteFrameConn, apis *FrameAPIServer, ctx context.Context) *FrameAPIServerConn {
	c := s.NewFrameAPIServer(conn)
	go func() {
		select {
		case <-c.WaitClosed():
		case <-ctx.Done():
			c.CloseWithError(ctx.Err())
		}
	}()
	return c
}

func (c *FrameAPIServer) ConcealAPI(apiName string) {
	c.nonBlockingApis.Delete(apiName)
}

func (c *FrameAPIServer) ExposeAPI(apiName string, api defines.NewMasterNodeServerAPI, newGoroutine bool) {
	if !strings.HasPrefix(apiName, "/") {
		apiName = "/" + apiName
	}
	c.nonBlockingApis.Set(apiName, func(caller defines.NewMasterNodeCaller, args defines.Values, setResult func(defines.Values, error)) {
		if newGoroutine {
			go func() {
				ret, err := api(caller, args)
				setResult(ret, err)
			}()
		} else {
			ret, err := api(caller, args)
			setResult(ret, err)
		}
	})
}

func (c *FrameAPIServer) CallOmitResponse(callee defines.NewMasterNodeCaller, api string, args defines.Values) {
	conn, ok := c.conns.Get(string(callee))
	if !ok {
		return
	}
	conn.CallOmitResponse(api, args)
}

func (c *FrameAPIServer) CallWithResponse(callee defines.NewMasterNodeCaller, api string, args defines.Values) async_wrapper.AsyncResult[defines.Values] {
	conn, ok := c.conns.Get(string(callee))
	if !ok {
		return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[defines.Values]) {
			ac.SetResult(defines.Empty)
		}, false)
	}
	return conn.CallWithResponse(api, args)
}

func (c *FrameAPIServerConn) Run() {
	c.FrameConn.ReadRoutine(func(data []byte) {
		frames := bytesToBytesSlices(data)
		indexOrApi := string(frames[0])
		if strings.HasPrefix(indexOrApi, "/") {
			index := frames[1]
			if apiFn, ok := c.nonBlockingApis.Get(indexOrApi); ok {
				apiFn(defines.NewMasterNodeCaller(c.identity), frames[2:], func(z defines.Values, err error) {
					if len(index) == 0 {
						return
					}
					frames := append([][]byte{index}, defines.WrapError(z, err)...)
					c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
				})
			}
		} else {
			if cb, ok := c.cbs.GetAndDelete(indexOrApi); ok {
				ret, err := defines.ValueWithErr(frames[1:]).Unwrap()
				cb(ret, err)
			}
		}
	})
}

func (c *FrameAPIServerConn) CallOmitResponse(api string, args defines.Values) {
	if !strings.HasPrefix(api, "/") {
		api = "/" + api
	}
	frames := append([][]byte{[]byte(api), {}}, args...)
	c.FrameConn.WriteBytePacket(byteSlicesToBytes(frames))
}

func (c *FrameAPIServerConn) CallWithResponse(api string, args defines.Values) async_wrapper.AsyncResult[defines.Values] {
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

// type serverVoidRespHandler struct{}

// func (h *serverVoidRespHandler) SetContext(ctx context.Context) defines.NewMasterNodeResultHandler   { return h }
// func (h *serverVoidRespHandler) SetTimeout(timeout time.Duration) defines.NewMasterNodeResultHandler { return h }
// func (h *serverVoidRespHandler) BlockGetResponse() defines.Values                          { return defines.Empty }
// func (h *serverVoidRespHandler) AsyncGetResponse(callback func(defines.Values)) {
// 	go callback(defines.Empty)
// }

// func (c *FrameAPIServer) CallWithResponse(callee defines.NewMasterNodeCaller, api string, args defines.Values) defines.NewMasterNodeResultHandler {
// 	conn, ok := c.conns.Get(string(callee))
// 	if !ok {
// 		return &serverVoidRespHandler{}
// 	}
// 	return conn.CallWithResponse(api, args)
// }

// type serverRespHandler struct {
// 	idx    string
// 	frames [][]byte
// 	c      *FrameAPIServerConn
// 	ctx    context.Context
// }

// func (h *serverRespHandler) doSend() {
// 	h.c.FrameConn.WriteBytePacket(byteSlicesToBytes(h.frames))
// }

// func (h *serverRespHandler) SetContext(ctx context.Context) defines.NewMasterNodeResultHandler {
// 	h.ctx = ctx
// 	return h
// }

// func (h *serverRespHandler) SetTimeout(timeout time.Duration) defines.NewMasterNodeResultHandler {
// 	if h.ctx == nil {
// 		h.ctx = context.Background()
// 	}
// 	h.ctx, _ = context.WithTimeout(h.ctx, timeout)
// 	return h
// }

// func (h *serverRespHandler) BlockGetResponse() defines.Values {
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

// func (h *serverRespHandler) AsyncGetResponse(callback func(defines.Values)) {
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

// func (c *FrameAPIServerConn) CallWithResponse(api string, args defines.Values) defines.NewMasterNodeResultHandler {
// 	if !strings.HasPrefix(api, "/") {
// 		api = "/" + api
// 	}
// 	idx := uuid.New().String()
// 	frames := append([][]byte{[]byte(api), []byte(idx)}, args...)
// 	return &serverRespHandler{
// 		idx, frames, c, nil,
// 	}
// }
