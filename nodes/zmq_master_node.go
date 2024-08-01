package nodes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
	"github.com/OmineDev/neomega-core/utils/waitable_queue"
)

type SlaveNodeInfo struct {
	Ctx              context.Context
	cancelFn         func()
	MsgToPub         *waitable_queue.WaitableQueue[defines.Values]
	SubScribedTopics *sync_wrapper.SyncKVMap[string, struct{}]
	ExposedApis      *sync_wrapper.SyncKVMap[string, struct{}]
	AcquiredLocks    *sync_wrapper.SyncKVMap[string, struct{}]
	Tags             *sync_wrapper.SyncKVMap[string, struct{}]
}

var ErrNotReg = errors.New("need reg node first")

type NewMasterNodeMasterNode struct {
	server defines.NewMasterNodeAPIServer
	*LocalAPINode[defines.Values, defines.Values]
	*LocalLock
	*LocalTags
	*LocalTopicNet[defines.Values]
	slaves                *sync_wrapper.SyncKVMap[string, *SlaveNodeInfo]
	subscribeMu           sync.RWMutex
	slaveSubscribedTopics map[string]map[string]*waitable_queue.WaitableQueue[defines.Values]
	ApiProvider           *sync_wrapper.SyncKVMap[string, string]
	values                *sync_wrapper.SyncKVMap[string, defines.Values]
	can_close.CanCloseWithError
}

func (n *NewMasterNodeMasterNode) IsMaster() bool {
	return true
}

func (n *NewMasterNodeMasterNode) onNewNode(id string) *SlaveNodeInfo {
	// fmt.Println("new client: ", id)
	ctx, cancelFn := context.WithCancel(context.Background())
	nodeInfo := &SlaveNodeInfo{
		Ctx:              ctx,
		cancelFn:         cancelFn,
		MsgToPub:         waitable_queue.NewWaitableQueue[defines.Values](),
		SubScribedTopics: sync_wrapper.NewSyncKVMap[string, struct{}](),
		ExposedApis:      sync_wrapper.NewSyncKVMap[string, struct{}](),
		AcquiredLocks:    sync_wrapper.NewSyncKVMap[string, struct{}](),
		Tags:             sync_wrapper.NewSyncKVMap[string, struct{}](),
	}
	n.slaves.Set(id, nodeInfo)
	return nodeInfo
}

func (n *NewMasterNodeMasterNode) onNodeOffline(id string, info *SlaveNodeInfo) {
	info.cancelFn()
	info.MsgToPub.Put(nil)
	info.SubScribedTopics.Iter(func(topic string, v struct{}) (continueInter bool) {
		n.subscribeMu.Lock()
		defer n.subscribeMu.Unlock()
		if slaves, ok := n.slaveSubscribedTopics[topic]; ok {
			delete(slaves, id)
		}
		return true
	})
	info.ExposedApis.Iter(func(apiName string, v struct{}) (continueInter bool) {
		if providerID, ok := n.ApiProvider.Get(apiName); ok && (providerID == id) {
			n.server.ConcealAPI(apiName)
			n.LocalAPINode.RemoveAPI(apiName)
		}
		return true
	})
	info.AcquiredLocks.Iter(func(k string, v struct{}) (continueInter bool) {
		n.unlock(k, id)
		return true
	})
	// close(info.MsgToPub)
	n.slaves.Delete(id)
	// fmt.Printf("node %v offline\n", string(id))
}

func (n *NewMasterNodeMasterNode) publishMessage(source string, topic string, msg defines.Values) {
	msgWithTopic := defines.FromString(topic).Extend(msg)
	n.LocalTopicNet.publishMessage(topic, msg)
	n.subscribeMu.RLock()
	defer n.subscribeMu.RUnlock()
	subScribers, ok := n.slaveSubscribedTopics[topic]
	if ok {
		for receiver, msgC := range subScribers {
			if receiver == source {
				continue
			}
			msgC.Put(msgWithTopic)
		}
	}
}

func (n *NewMasterNodeMasterNode) PublishMessage(topic string, msg defines.Values) {
	n.publishMessage("", topic, msg)
}

func (n *NewMasterNodeMasterNode) suppressClientApiIfAny(apiName string, newProvider string) {
	provider, found := n.ApiProvider.Get(apiName)
	if !found || provider == newProvider {
		return
	}
	n.server.CallOmitResponse(defines.NewMasterNodeCaller(provider), "/suppress-api", defines.FromString(apiName))
}

func (n *NewMasterNodeMasterNode) ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[defines.Values, defines.Values] {
	n.suppressClientApiIfAny(apiName, "")
	return async_wrapper.NewApiSetter[defines.Values, defines.Values](func(f func(in defines.Values, setResult func(defines.Values, error))) {
		// master to master call
		n.LocalAPINode.ExposeAPI(apiName).CallBackAPI(f)
		// slave to master call
		n.server.ExposeAPI(apiName).CallBackAPI(func(in defines.ArgWithCaller, setResult func(defines.Values, error)) {
			// slave info is omitted
			f(in.Args, setResult)
		})
	}, false)
}

func (c *NewMasterNodeMasterNode) GetValue(key string) (val defines.Values, found bool) {
	v, ok := c.values.Get(key)
	return v, ok
}

func (c *NewMasterNodeMasterNode) SetValue(key string, val defines.Values) {
	c.values.Set(key, val)
}

func (c *NewMasterNodeMasterNode) CheckNetTag(tag string) bool {
	ok := c.LocalTags.CheckLocalTag(tag)
	if ok {
		return true
	}
	found := false
	c.slaves.Iter(func(k string, v *SlaveNodeInfo) bool {
		_, ok := v.Tags.Get(tag)
		if ok {
			found = true
		}
		return !found
	})
	return found
}

func (c *NewMasterNodeMasterNode) tryLock(name string, acquireTime time.Duration, owner string) bool {
	if !c.LocalLock.tryLock(name, acquireTime, owner) {
		return false
	}
	if owner != "" {
		slaveInfo, ok := c.slaves.Get(owner)
		if ok {
			slaveInfo.AcquiredLocks.Set(name, struct{}{})
		}
	}
	return true
}

func (c *NewMasterNodeMasterNode) unlock(name string, owner string) {
	if c.LocalLock.unlock(name, owner) {
		if owner != "" {
			slaveInfo, ok := c.slaves.Get(owner)
			if ok {
				slaveInfo.AcquiredLocks.Delete(name)
			}
		}
	}
}

func (master *NewMasterNodeMasterNode) exposePingFunc() {
	master.server.ExposeAPI("/ping").InstantAPI(func(argsWithCaller defines.ArgWithCaller) (defines.Values, error) {
		return defines.Values{[]byte("pong")}, nil
	})
}

func (master *NewMasterNodeMasterNode) exposeNewClientFunc() {
	master.server.ExposeAPI("/new_client").InstantAPI(func(argsWithCaller defines.ArgWithCaller) (defines.Values, error) {
		caller := argsWithCaller.Caller
		nodeInfo := master.onNewNode(string(caller))
		master.server.SetOnCloseCleanUp(caller, func() {
			master.onNodeOffline(string(caller), nodeInfo)
		})
		go func() {
			for {
				msg := nodeInfo.MsgToPub.Get()
				if nodeInfo.Ctx.Err() != nil {
					return
				}
				master.server.CallOmitResponse(caller, "/on_new_msg", msg)
			}
		}()
		return defines.Empty, nil
	})
}

func (master *NewMasterNodeMasterNode) exposeTopicFunc() {
	master.server.ExposeAPI("/subscribe").InstantAPI(func(argsWithCaller defines.ArgWithCaller) (defines.Values, error) {
		caller := argsWithCaller.Caller
		args := argsWithCaller.Args
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return defines.Empty, fmt.Errorf("must call new client first")
		}
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty, fmt.Errorf("cannot get topic name")
		}
		slaveInfo.SubScribedTopics.Set(topic, struct{}{})
		master.subscribeMu.Lock()
		subscribers, found := master.slaveSubscribedTopics[topic]
		if !found {
			subscribers = map[string]*waitable_queue.WaitableQueue[defines.Values]{}
			master.slaveSubscribedTopics[topic] = subscribers
		}
		subscribers[string(caller)] = slaveInfo.MsgToPub

		master.subscribeMu.Unlock()

		return defines.Empty, nil
	})
	master.server.ExposeAPI("/publish").InstantAPI(func(argsWithCaller defines.ArgWithCaller) (defines.Values, error) {
		caller := argsWithCaller.Caller
		args := argsWithCaller.Args
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty, fmt.Errorf("cannot get topic name")
		}
		msg := args.ConsumeHead()
		master.publishMessage(string(caller), topic, msg)
		return defines.Empty, nil
	})
}

func (master *NewMasterNodeMasterNode) exposeRegApiFunc() {
	master.server.ExposeAPI("/reg_api").InstantAPI(func(argsWithProvider defines.ArgWithCaller) (defines.Values, error) {
		provider := argsWithProvider.Caller
		apiArgs := argsWithProvider.Args
		apiName, err := apiArgs.ToString()
		if err != nil {
			return defines.Empty, fmt.Errorf("cannot get api name")
		}
		slaveInfo, ok := master.slaves.Get(string(provider))
		if !ok {
			return defines.Empty, fmt.Errorf("must call new client first")
		}
		// surpress old api provider
		master.suppressClientApiIfAny(apiName, string(provider))
		slaveInfo.ExposedApis.Set(apiName, struct{}{})
		master.ApiProvider.Set(apiName, string(provider))
		// master to slave call
		master.LocalAPINode.ExposeAPI(apiName).CallBackAPI(func(in defines.Values, onResult func(defines.Values, error)) {
			providerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				onResult(defines.Empty, fmt.Errorf("not found"))
			} else {
				master.server.CallWithResponse(provider, apiName, in).SetContext(providerInfo.Ctx).AsyncGetResult(onResult)
			}
		})
		// slave to salve call
		master.server.ExposeAPI(apiName).CallBackAPI(func(inAndCaller defines.ArgWithCaller, onResult func(defines.Values, error)) {
			providerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				onResult(defines.Empty, fmt.Errorf("not found"))
			}
			master.server.CallWithResponse(provider, apiName, inAndCaller.Args).SetContext(providerInfo.Ctx).AsyncGetResult(onResult)
		})
		return defines.Empty, nil
	})
}

func (master *NewMasterNodeMasterNode) exposeNetValueFunc() {
	master.ExposeAPI("/set-value").InstantAPI(func(args defines.Values) (defines.Values, error) {
		key, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		master.values.Set(key, args.ConsumeHead())
		return defines.Empty, nil
	})
	master.ExposeAPI("/get-value").InstantAPI(func(args defines.Values) (defines.Values, error) {
		key, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		head, ok := master.values.Get(key)
		if ok {
			return head, nil
		} else {
			return defines.Empty, nil
		}
	})
}

func (master *NewMasterNodeMasterNode) exposeNetTagFunc() {
	master.server.ExposeAPI("/set-tags").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		tags := callerAndArgs.Args.ToStrings()
		slaveInfo, ok := master.slaves.Get(string(callerAndArgs.Caller))
		if !ok {
			return defines.Empty, fmt.Errorf("need reg first")
		}
		for _, tag := range tags {
			slaveInfo.Tags.Set(tag, struct{}{})
		}
		return defines.Empty, nil
	})
	master.server.ExposeAPI("/check-tag").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		tag, err := callerAndArgs.Args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		has := master.CheckNetTag(tag)
		return defines.FromBool(has), nil
	})
}

func (master *NewMasterNodeMasterNode) exposeNetLockFunc() {
	master.server.ExposeAPI("/try-lock").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		lockName, err := callerAndArgs.Args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		ms, err := callerAndArgs.Args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.Empty, err
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.tryLock(lockName, lockTime, string(callerAndArgs.Caller))
		return defines.FromBool(has), nil
	})
	master.server.ExposeAPI("/reset-lock-time").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		lockName, err := callerAndArgs.Args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		ms, err := callerAndArgs.Args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.Empty, err
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.resetLockTime(lockName, lockTime, string(callerAndArgs.Caller))
		return defines.FromBool(has), nil
	})
	master.server.ExposeAPI("/unlock").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		lockName, err := callerAndArgs.Args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		master.unlock(lockName, string(callerAndArgs.Caller))
		return defines.Empty, nil
	})
}

func NewMasterNode(server defines.NewMasterNodeAPIServer) defines.Node {
	master := &NewMasterNodeMasterNode{
		server:                server,
		LocalAPINode:          NewLocalAPINode[defines.Values, defines.Values](),
		LocalLock:             NewLocalLock(),
		LocalTags:             NewLocalTags(),
		LocalTopicNet:         NewLocalTopicNet[defines.Values](),
		slaves:                sync_wrapper.NewSyncKVMap[string, *SlaveNodeInfo](),
		subscribeMu:           sync.RWMutex{},
		slaveSubscribedTopics: map[string]map[string]*waitable_queue.WaitableQueue[defines.Values]{},
		ApiProvider:           sync_wrapper.NewSyncKVMap[string, string](),
		values:                sync_wrapper.NewSyncKVMap[string, defines.Values](),
		CanCloseWithError:     can_close.NewClose(server.Close),
	}
	go func() {
		master.CloseWithError(<-server.WaitClosed())
	}()
	master.exposePingFunc()
	master.exposeNewClientFunc()
	master.exposeTopicFunc()
	master.exposeRegApiFunc()
	master.exposeNetValueFunc()
	master.exposeNetTagFunc()
	master.exposeNetLockFunc()
	return master
}
