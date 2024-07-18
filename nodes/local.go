package nodes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
)

type LocalAPINode[inT any, outT any] struct {
	RegednonBlockingApi async_wrapper.AsyncAPIGroup[inT, outT]
}

var ErrAPIExist = errors.New("defines.API already exposed")
var ErrAPINotExist = errors.New("defines.API not exist")

func (n *LocalAPINode[inT, outT]) ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[inT, outT] {
	if n.RegednonBlockingApi.Exist(apiName) {
		fmt.Printf("an new api shadow an exist api with same name: %v\n", apiName)
	}
	return n.RegednonBlockingApi.AddAPI(apiName)
}

func (n *LocalAPINode[inT, outT]) RemoveAPI(apiName string) {
	n.RegednonBlockingApi.RemoveAPI(apiName)
}

func (n *LocalAPINode[inT, outT]) HasAPI(apiName string) bool {
	return n.RegednonBlockingApi.Exist(apiName)
}

func (c *LocalAPINode[inT, outT]) CallWithResponse(api string, args inT) async_wrapper.AsyncResult[outT] {
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[outT]) {
		c.RegednonBlockingApi.CallAPI(api, args, func(ret outT, err error) {
			ac.SetResultAndErr(ret, err)
		})
	}, false)
}

func (c *LocalAPINode[inT, outT]) CallOmitResponse(api string, args inT) {
	c.RegednonBlockingApi.CallAPI(api, args, func(ret outT, err error) {})
}

func NewLocalAPINode[inT any, outT any]() *LocalAPINode[inT, outT] {
	return &LocalAPINode[inT, outT]{
		RegednonBlockingApi: async_wrapper.NewAsyncAPIGroup[inT, outT](),
	}
}

type timeLock struct {
	ctx      context.Context
	reset    chan struct{}
	cancelFN func()
	owner    string
	unlocked bool
}

type LocalLock struct {
	Locks *sync_wrapper.SyncKVMap[string, *timeLock]
}

func (c *LocalLock) tryLock(name string, acquireTime time.Duration, owner string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	if acquireTime > 0 {
		ctx, _ = context.WithTimeout(ctx, acquireTime)
	}
	l := &timeLock{
		ctx:      ctx,
		reset:    make(chan struct{}),
		cancelFN: cancel,
		owner:    owner,
	}
	_, locked := c.Locks.GetOrSet(name, l)
	if locked {
		return false
	} else {
		go func() {
			select {
			case <-l.ctx.Done():
				if !l.unlocked {
					c.unlock(name, owner)
				}
			case <-l.reset:
				return
			}
		}()
	}
	return true
}

func (c *LocalLock) unlock(name string, owner string) bool {
	l, ok := c.Locks.GetAndDelete(name)
	if !ok {
		return false
	}
	if l.owner == owner {
		l.unlocked = true
		l.cancelFN()
		return true
	}
	return false
}

func (c *LocalLock) resetLockTime(name string, acquireTime time.Duration, owner string) bool {
	l, ok := c.Locks.Get(name)
	if !ok {
		return false
	}
	if l.owner == owner {
		// remove previous ctx
		previousReset := l.reset

		// make a new ctx
		ctx, cancel := context.WithTimeout(context.Background(), acquireTime)
		if acquireTime > 0 {

		} else {
			if !l.unlocked {
				c.unlock(name, owner)
			}
			c.Locks.Delete(name)
			cancel()
			return true
		}
		l.ctx = ctx
		l.cancelFN = cancel
		l.reset = make(chan struct{})
		go func() {
			select {
			case <-l.ctx.Done():
				if !l.unlocked {
					c.unlock(name, owner)
				}
			case <-l.reset:
				return
			}
		}()
		close(previousReset)
		return true
	}
	return false
}

func (c *LocalLock) TryLock(name string, acquireTime time.Duration) bool {
	return c.tryLock(name, acquireTime, "")
}

func (c *LocalLock) ResetLockTime(name string, acquireTime time.Duration) bool {
	return c.resetLockTime(name, acquireTime, "")
}

func (c *LocalLock) Unlock(name string) {
	c.unlock(name, "")
}

func NewLocalLock() *LocalLock {
	return &LocalLock{
		Locks: sync_wrapper.NewSyncKVMap[string, *timeLock](),
	}
}

type LocalTags struct {
	tags *sync_wrapper.SyncKVMap[string, struct{}]
}

func (n *LocalTags) SetTags(tags ...string) {
	for _, tag := range tags {
		n.tags.Set(tag, struct{}{})
	}

}
func (n *LocalTags) CheckNetTag(tag string) bool {
	_, ok := n.tags.Get(tag)
	return ok
}

func (n *LocalTags) CheckLocalTag(tag string) bool {
	return n.CheckNetTag(tag)
}

func NewLocalTags() *LocalTags {
	return &LocalTags{
		tags: sync_wrapper.NewSyncKVMap[string, struct{}](),
	}
}

type LocalTopicNet[dataT any] struct {
	listenedTopics *sync_wrapper.SyncKVMap[string, []func(dataT)]
}

func NewLocalTopicNet[dataT any]() *LocalTopicNet[dataT] {
	return &LocalTopicNet[dataT]{
		listenedTopics: sync_wrapper.NewSyncKVMap[string, []func(dataT)](),
	}
}

func (n *LocalTopicNet[dataT]) ListenMessage(topic string, listener func(dataT), newGoroutine bool) {
	wrappedListener := func(msg dataT) {
		if newGoroutine {
			go listener(msg)
		} else {
			listener(msg)
		}
	}
	n.listenedTopics.UnsafeGetAndUpdate(topic, func(currentListeners []func(dataT)) []func(dataT) {
		if currentListeners == nil {
			return []func(dataT){wrappedListener}
		}
		currentListeners = append(currentListeners, wrappedListener)
		return currentListeners
	})
}

func (n *LocalTopicNet[dataT]) publishMessage(topic string, msg dataT) {
	listeners, _ := n.listenedTopics.Get(topic)
	for _, listener := range listeners {
		listener(msg)
	}
}

func (n *LocalTopicNet[dataT]) PublishMessage(topic string, msg dataT) {
	n.publishMessage(topic, msg)
}

type LocalNode[inT any, outT any, dataT any] struct {
	*LocalAPINode[inT, outT]
	*LocalLock
	*LocalTags
	*LocalTopicNet[dataT]
	can_close.CanCloseWithError
	values *sync_wrapper.SyncKVMap[string, dataT]
}

func (n *LocalNode[inT, outT, dataT]) GetValue(key string) (val dataT, found bool) {
	return n.values.Get(key)
}
func (n *LocalNode[inT, outT, dataT]) SetValue(key string, val dataT) {
	n.values.Set(key, val)
}

func NewLocalNode(ctx context.Context) defines.Node {
	n := &LocalNode[defines.Values, defines.Values, defines.Values]{
		LocalAPINode:      NewLocalAPINode[defines.Values, defines.Values](),
		LocalLock:         NewLocalLock(),
		LocalTags:         NewLocalTags(),
		LocalTopicNet:     NewLocalTopicNet[defines.Values](),
		CanCloseWithError: can_close.NewClose(func() {}),
		values:            sync_wrapper.NewSyncKVMap[string, defines.Values](),
	}
	go func() {
		<-ctx.Done()
		n.CloseWithError(ctx.Err())
	}()
	return n
}
