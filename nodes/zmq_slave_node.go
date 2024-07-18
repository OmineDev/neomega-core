package nodes

import (
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type NewMasterNodeSlaveNode struct {
	client        defines.NewMasterNodeAPIClient
	localAPI      *LocalAPINode
	localTags     *LocalTags
	localTopicNet *LocalTopicNet
	can_close.CanCloseWithError
}

func (n *NewMasterNodeSlaveNode) IsMaster() bool {
	return false
}

func (n *NewMasterNodeSlaveNode) ListenMessage(topic string, listener defines.MsgListener, newGoroutine bool) {
	n.client.CallWithResponse("/subscribe", defines.FromString(topic)).BlockGetResult()
	n.localTopicNet.ListenMessage(topic, listener, newGoroutine)
}

func (n *NewMasterNodeSlaveNode) PublishMessage(topic string, msg defines.Values) {
	n.client.CallOmitResponse("/publish", defines.FromString(topic).Extend(msg))
	n.localTopicNet.publishMessage(topic, msg)
}

func (n *NewMasterNodeSlaveNode) ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[defines.Values, defines.Values] {
	n.client.CallOmitResponse("/reg_api", defines.FromString(apiName))
	return async_wrapper.NewApiSetter(func(f func(in defines.Values, setResult func(defines.Values, error))) {
		// salve to master & salve (other call)
		n.client.ExposeAPI(apiName).CallBackAPI(f)
		// salve to salve (self) call
		n.localAPI.ExposeAPI(apiName).CallBackAPI(f)
	}, false)
}

func (c *NewMasterNodeSlaveNode) CallOmitResponse(api string, args defines.Values) {
	if c.localAPI.HasAPI(api) {
		c.localAPI.CallOmitResponse(api, args)
	} else {
		c.client.CallOmitResponse(api, args)
	}
}

func (c *NewMasterNodeSlaveNode) CallWithResponse(api string, args defines.Values) async_wrapper.AsyncResult[defines.Values] {
	if c.localAPI.HasAPI(api) {
		return c.localAPI.CallWithResponse(api, args)
	} else {
		return c.client.CallWithResponse(api, args)
	}
}

func (c *NewMasterNodeSlaveNode) GetValue(key string) (val defines.Values, found bool) {
	v, err := c.CallWithResponse("/get-value", defines.FromString(key)).BlockGetResult()
	if err != nil || v.IsEmpty() {
		return nil, false
	} else {
		return v, true
	}
}

func (c *NewMasterNodeSlaveNode) SetValue(key string, val defines.Values) {
	c.CallOmitResponse("/set-value", defines.FromString(key).Extend(val))
}

func (c *NewMasterNodeSlaveNode) SetTags(tags ...string) {
	c.CallOmitResponse("/set-tags", defines.FromStrings(tags...))
	c.localTags.SetTags(tags...)
}

func (c *NewMasterNodeSlaveNode) CheckNetTag(tag string) bool {
	if c.localTags.CheckLocalTag(tag) {
		return true
	}
	rest, err := c.CallWithResponse("/check-tag", defines.FromString(tag)).BlockGetResult()
	if err != nil || rest.IsEmpty() {
		return false
	}
	hasTag, err := rest.ToBool()
	if err != nil {
		return false
	}
	return hasTag
}

func (n *NewMasterNodeSlaveNode) CheckLocalTag(tag string) bool {
	return n.localTags.CheckLocalTag(tag)
}

func (c *NewMasterNodeSlaveNode) TryLock(name string, acquireTime time.Duration) bool {
	rest, err := c.CallWithResponse("/try-lock", defines.FromString(name).Extend(defines.FromInt64(acquireTime.Milliseconds()))).BlockGetResult()
	if err != nil || rest.IsEmpty() {
		return false
	}
	locked, err := rest.ToBool()
	if err != nil {
		return false
	}
	return locked
}

func (c *NewMasterNodeSlaveNode) ResetLockTime(name string, acquireTime time.Duration) bool {
	rest, err := c.CallWithResponse("/reset-lock-time", defines.FromString(name).Extend(defines.FromInt64(acquireTime.Milliseconds()))).BlockGetResult()
	if err != nil || rest.IsEmpty() {
		return false
	}
	locked, err := rest.ToBool()
	if err != nil {
		return false
	}
	return locked
}

func (c *NewMasterNodeSlaveNode) Unlock(name string) {
	c.CallOmitResponse("/unlock", defines.FromString(name))
}

func NewSlaveNode(client defines.NewMasterNodeAPIClient) (defines.Node, error) {
	slave := &NewMasterNodeSlaveNode{
		client:            client,
		localAPI:          NewLocalAPINode(),
		localTags:         NewLocalTags(),
		localTopicNet:     NewLocalTopicNet(),
		CanCloseWithError: can_close.NewClose(client.Close),
	}
	client.ExposeAPI("/ping").InstantAPI(func(args defines.Values) (defines.Values, error) {
		return defines.Values{[]byte("pong")}, nil
	})
	client.ExposeAPI("/on_new_msg").InstantAPI(func(args defines.Values) (defines.Values, error) {
		topic, err := args.ToString()
		if err != nil {
			return defines.Empty, nil
		}
		msg := args.ConsumeHead()
		slave.localTopicNet.PublishMessage(topic, msg)
		return defines.Empty, nil
	})
	client.ExposeAPI("/suppress-api").InstantAPI(func(args defines.Values) (defines.Values, error) {
		apiName, _ := args.ToString()
		fmt.Printf("an api is suppress by net api with same name: %v\n", apiName)
		return defines.Empty, nil
	})

	go func() {
		slave.CloseWithError(<-client.WaitClosed())
	}()
	// go slave.heatBeat()

	if _, err := slave.client.CallWithResponse("/new_client", defines.Empty).BlockGetResult(); err != nil {
		return nil, fmt.Errorf("fail to reg new client to master: " + err.Error())
	} else {
		return slave, nil
	}
}
