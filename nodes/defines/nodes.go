package defines

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type Values [][]byte
type ValueWithErr Values

var ErrNoResult = errors.New("no result")

func (vs ValueWithErr) Unwrap() (Values, error) {
	if Values(vs).IsEmpty() {
		return Empty, ErrNoResult
	} else {
		if Values(vs).EqualString("ok") {
			return Values(vs).ConsumeHead(), nil
		} else {
			errString, _ := Values(vs).ToString()
			return Values(vs).ConsumeHead(), fmt.Errorf(strings.TrimPrefix(errString, "err:"))
		}
	}
}

func WrapError(rets Values, err error) ValueWithErr {
	if err != nil {
		return ValueWithErr(FromString("err:" + err.Error()).Extend(rets))
	} else {
		return ValueWithErr(FromString("ok").Extend(rets))
	}
}

type API func(args Values) (result Values, err error)
type MsgListener func(msg Values)

type APINode interface {
	// Point-to-Point Remote Process Call
	ExposeAPI(apiName string, api API, newGoroutine bool) error
	CallOmitResponse(api string, args Values)
	CallWithResponse(api string, args Values) *async_wrapper.AsyncWrapper[Values]
}

type TopicNetNode interface {
	// Multi-to-Multi Message Publish & Subscribe
	PublishMessage(topic string, msg Values)
	ListenMessage(topic string, listener MsgListener, newGoroutine bool)
}

type FundamentalNode interface {
	APINode
	TopicNetNode
}

type KVDataNode interface {
	// Global KV data
	GetValue(key string) (val Values, found bool)
	SetValue(key string, val Values)
}

type RolesNode interface {
	// Tags
	SetTags(tags ...string)
	CheckNetTag(tag string) bool
	CheckLocalTag(tag string) bool
}

type TimeLockNode interface {
	// Lock
	TryLock(name string, acquireTime time.Duration) bool
	ResetLockTime(name string, acquireTime time.Duration) bool
	Unlock(name string)
}

type Node interface {
	can_close.CanCloseWithError
	FundamentalNode
	KVDataNode
	RolesNode
	TimeLockNode
}

type AsyncAPI func(args Values, setResult func(rets Values, err error))
