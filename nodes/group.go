package nodes

import (
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type group struct {
	defines.Node
	name          string
	allowAbsolute bool
}

func (n *group) translateName(name string) string {
	// absolute path
	if n.allowAbsolute && strings.HasPrefix(name, "/") {
		return name
	} else {
		return n.name + "/" + name
	}
}

func (n *group) ExposeAPI(apiName string) async_wrapper.AsyncAPISetHandler[defines.Values, defines.Values] {
	return n.Node.ExposeAPI(n.translateName(apiName))
}
func (n *group) CallOmitResponse(api string, args defines.Values) {
	n.Node.CallOmitResponse(n.translateName(api), args)
}
func (n *group) CallWithResponse(api string, args defines.Values) async_wrapper.AsyncResult[defines.Values] {
	return n.Node.CallWithResponse(n.translateName(api), args)
}
func (n *group) PublishMessage(topic string, msg defines.Values) {
	n.Node.PublishMessage(n.translateName(topic), msg)
}
func (n *group) ListenMessage(topic string, listener func(msg defines.Values), newGoroutine bool) {
	n.Node.ListenMessage(n.translateName(topic), listener, newGoroutine)
}
func (n *group) GetValue(key string) (val defines.Values, found bool) {
	return n.Node.GetValue(n.translateName(key))
}
func (n *group) SetValue(key string, val defines.Values) {
	n.Node.SetValue(n.translateName(key), val)
}
func (n *group) SetTags(tags ...string) {
	ttrags := []string{}
	for _, tag := range tags {
		ttrags = append(ttrags, n.translateName(tag))
	}
	n.Node.SetTags(ttrags...)
}
func (n *group) CheckNetTag(tag string) bool {
	return n.Node.CheckNetTag(n.translateName(tag))
}
func (n *group) CheckLocalTag(tag string) bool {
	return n.Node.CheckLocalTag(n.translateName(tag))
}
func (n *group) TryLock(name string, acquireTime time.Duration) bool {
	return n.Node.TryLock(n.translateName(name), acquireTime)
}

func (n *group) ResetLockTime(name string, acquireTime time.Duration) bool {
	return n.Node.ResetLockTime(n.translateName(name), acquireTime)
}

func (n *group) Unlock(name string) {
	n.Node.Unlock(n.translateName(name))
}

func NewGroup(name string, node defines.Node, allowAbsolute bool) defines.Node {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return &group{node, name, allowAbsolute}
}
