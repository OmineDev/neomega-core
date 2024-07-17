package nodes

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
)

type SlaveNodeInfo struct {
	Ctx              context.Context
	cancelFn         func()
	MsgToPub         chan defines.Values
	SubScribedTopics *sync_wrapper.SyncKVMap[string, struct{}]
	ExposedApis      *sync_wrapper.SyncKVMap[string, struct{}]
	AcquiredLocks    *sync_wrapper.SyncKVMap[string, struct{}]
	Tags             *sync_wrapper.SyncKVMap[string, struct{}]
}

var ErrNotReg = errors.New("need reg node first")

type NewMasterNodeMasterNode struct {
	server defines.NewMasterNodeAPIServer
	*LocalAPINode
	*LocalLock
	*LocalTags
	*LocalTopicNet
	slaves                *sync_wrapper.SyncKVMap[string, *SlaveNodeInfo]
	subscribeMu           sync.RWMutex
	slaveSubscribedTopics map[string]map[string]chan defines.Values
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
		MsgToPub:         make(chan defines.Values, 128),
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
	msgWithTopic := n.LocalTopicNet.publishMessage(topic, msg)
	n.subscribeMu.RLock()
	defer n.subscribeMu.RUnlock()
	subScribers, ok := n.slaveSubscribedTopics[topic]
	if ok {
		for receiver, msgC := range subScribers {
			if receiver == source {
				continue
			}
			select {
			case msgC <- msgWithTopic:
			default:
				fmt.Println("communication between nodes too slow, msg queued!")
				go func() {
					msgC <- msgWithTopic
				}()
			}
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

func (n *NewMasterNodeMasterNode) ExposeAPI(apiName string, api defines.API, newGoroutine bool) error {
	n.suppressClientApiIfAny(apiName, "")
	// master to master call
	n.LocalAPINode.ExposeAPI(apiName, api, newGoroutine)
	// slave to master call
	n.server.ExposeAPI(apiName, func(caller defines.NewMasterNodeCaller, args defines.Values) (ret defines.Values, err error) {
		// slave info is omitted
		return api(args)
	}, newGoroutine)
	return nil
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
	master.server.ExposeAPI("/ping", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		return defines.Values{[]byte("pong")}, nil
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeNewClientFunc() {
	master.server.ExposeAPI("/new_client", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		nodeInfo := master.onNewNode(string(caller))
		master.server.SetOnCloseCleanUp(caller, func() {
			master.onNodeOffline(string(caller), nodeInfo)
		})
		go func() {
			for {
				select {
				case <-nodeInfo.Ctx.Done():
					return
				case msg := <-nodeInfo.MsgToPub:
					master.server.CallOmitResponse(caller, "/on_new_msg", msg)
				}
			}
		}()
		return defines.Empty, nil
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeTopicFunc() {
	master.server.ExposeAPI("/subscribe", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
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
			subscribers = make(map[string]chan defines.Values)
			master.slaveSubscribedTopics[topic] = subscribers
		}
		subscribers[string(caller)] = slaveInfo.MsgToPub

		master.subscribeMu.Unlock()

		return defines.Empty, nil
	}, false)
	master.server.ExposeAPI("/publish", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty, fmt.Errorf("cannot get topic name")
		}
		msg := args.ConsumeHead()
		master.publishMessage(string(caller), topic, msg)
		return defines.Empty, nil
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeRegApiFunc() {
	master.server.ExposeAPI("/reg_api", func(provider defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		apiName, err := args.ToString()
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
		master.LocalAPINode.ExposeAPI(apiName, func(args defines.Values) (result defines.Values, err error) {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return defines.Empty, fmt.Errorf("not found")
			} else {
				return master.server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResult()
			}
		}, true)
		// slave to salve call
		master.server.ExposeAPI(apiName, func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
			callerInfo, ok := master.slaves.Get(string(provider))
			if !ok {
				return defines.Empty, fmt.Errorf("not found")
			}
			return master.server.CallWithResponse(provider, apiName, args).SetContext(callerInfo.Ctx).BlockGetResult()
		}, true)
		return defines.Empty, nil
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeNetValueFunc() {
	master.ExposeAPI("/set-value", func(args defines.Values) (result defines.Values, err error) {
		key, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		master.values.Set(key, args.ConsumeHead())
		return defines.Empty, nil
	}, false)
	master.ExposeAPI("/get-value", func(args defines.Values) (result defines.Values, err error) {
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
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeNetTagFunc() {
	master.server.ExposeAPI("/set-tags", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		tags := args.ToStrings()
		slaveInfo, ok := master.slaves.Get(string(caller))
		if !ok {
			return defines.Empty, fmt.Errorf("need reg first")
		}
		for _, tag := range tags {
			slaveInfo.Tags.Set(tag, struct{}{})
		}
		return defines.Empty, nil
	}, false)
	master.server.ExposeAPI("/check-tag", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		tag, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		has := master.CheckNetTag(tag)
		return defines.FromBool(has), nil
	}, false)
}

func (master *NewMasterNodeMasterNode) exposeNetLockFunc() {
	master.server.ExposeAPI("/try-lock", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		lockName, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.Empty, err
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.tryLock(lockName, lockTime, string(caller))
		return defines.FromBool(has), nil
	}, false)
	master.server.ExposeAPI("/reset-lock-time", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		lockName, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		ms, err := args.ConsumeHead().ToInt64()
		if err != nil {
			return defines.Empty, err
		}
		lockTime := time.Duration(ms * int64(time.Millisecond))
		has := master.resetLockTime(lockName, lockTime, string(caller))
		return defines.FromBool(has), nil
	}, false)
	master.server.ExposeAPI("/unlock", func(caller defines.NewMasterNodeCaller, args defines.Values) (defines.Values, error) {
		lockName, err := args.ToString()
		if err != nil {
			return defines.Empty, err
		}
		master.unlock(lockName, string(caller))
		return defines.Empty, nil
	}, false)
}

func NewMasterNode(server defines.NewMasterNodeAPIServer) defines.Node {
	master := &NewMasterNodeMasterNode{
		server:                server,
		LocalAPINode:          NewLocalAPINode(),
		LocalLock:             NewLocalLock(),
		LocalTags:             NewLocalTags(),
		LocalTopicNet:         NewLocalTopicNet(),
		slaves:                sync_wrapper.NewSyncKVMap[string, *SlaveNodeInfo](),
		subscribeMu:           sync.RWMutex{},
		slaveSubscribedTopics: map[string]map[string]chan defines.Values{},
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
