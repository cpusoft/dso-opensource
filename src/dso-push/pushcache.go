package push

import (
	"strings"
	"sync"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type PushCache struct {
	// map[connAndRrKey]subscribeMessageId
	// map[connAndRrAnyKey]subscribeMessageId
	// map[connAndRrDelKey]subscribeMessageId
	subscribedRrs map[string]uint16
	rrMutex       sync.RWMutex
}

func NewPushCache() *PushCache {
	c := &PushCache{}
	c.subscribedRrs = make(map[string]uint16)
	return c
}

func (c *PushCache) subscribe(connAndRrKey, connAndRrAnyKey,
	connAndRrDelKey string, subscribeMessageId uint16) {
	c.rrMutex.Lock()
	defer c.rrMutex.Unlock()
	belogs.Debug("PushCache.subscribe(): connAndRrKey:", connAndRrKey, "  connAndRrAnyKey:", connAndRrAnyKey,
		"  connAndRrDelKey:", connAndRrDelKey, "  subscribeMessageId:", subscribeMessageId)

	c.subscribedRrs[connAndRrKey] = subscribeMessageId
	c.subscribedRrs[connAndRrAnyKey] = subscribeMessageId
	c.subscribedRrs[connAndRrDelKey] = subscribeMessageId
	belogs.Debug("PushCache.subscribe(): after add, subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
}

func (c *PushCache) unsubscribe(subscribeMessageId uint16) {
	c.rrMutex.Lock()
	defer c.rrMutex.Unlock()
	belogs.Debug("PushCache.unsubscribe(): subscribeMessageId:", subscribeMessageId)
	for k, v := range c.subscribedRrs {
		if v == subscribeMessageId {
			delete(c.subscribedRrs, k)
		}
	}

	belogs.Debug("PushCache.unsubscribe(): after del, subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
}

func (c *PushCache) delConnKey(connKey string) {
	c.rrMutex.Lock()
	defer c.rrMutex.Unlock()
	belogs.Debug("PushCache.delConnKey():before connKey:", connKey, "  subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
	rrs := make(map[string]uint16)
	for k, v := range c.subscribedRrs {
		if !strings.Contains(k, connKey) {
			rrs[k] = v
		}
	}
	c.subscribedRrs = rrs
	belogs.Debug("PushCache.delConnKey():after connKey:", connKey, "  subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
}

// may no resultRrModels, no should push
/*
rfc8765 6.3.1. PUSH Message

del:
 if TTL==0xFFFFFFFF and Class!=ANY and Type!=ANY , then del by FullDomain and Class and Type and RData
 if TTL==0xFFFFFFFE
	if Class!=ANY and Type!=ANY, then del by FullDomain and Type and Class
 	if Class!=ANY and Type==ANY, then del by FullDomain and Class
 	if Class==ANY , delete by fullDomain
*/
func (c *PushCache) foundInSubscribedRrs(rrKey, rrAnyKey, rrDelKey string) (connAndRrKeys []string, err error) {
	c.rrMutex.RLock()
	defer c.rrMutex.RUnlock()
	belogs.Debug("PushCache.foundInSubscribedRrs(): subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
	connAndRrKeys = make([]string, 0)
	for k, _ := range c.subscribedRrs {
		// add: rrKey or rrAnyKey
		if strings.Contains(k, rrKey) || strings.Contains(k, rrAnyKey) {
			connAndRrKeys = append(connAndRrKeys, k)
			belogs.Debug("PushCache.foundInSubscribedRrs(): found in rrKey or rrAnyKey, k:", k,
				" rrKey:", rrKey, "  rrAnyKey:", rrAnyKey, "  connAndRrKeys:", connAndRrKeys)
		} else if len(rrDelKey) > 0 && strings.Contains(k, rrDelKey) {
			connAndRrKeys = append(connAndRrKeys, k)
			belogs.Debug("PushCache.foundInSubscribedRrs(): found in rrDelKey, k:", k,
				"  rrDelKey:", rrDelKey, "  connAndRrKeys:", connAndRrKeys)
		}
	}
	belogs.Debug("PushCache.foundInSubscribedRrs(): connAndRrKeys:", jsonutil.MarshalJson(connAndRrKeys))
	return connAndRrKeys, nil
}
func (c *PushCache) foundAllConnAndRrKeys() (connAndRrKeys []string, err error) {
	c.rrMutex.RLock()
	defer c.rrMutex.RUnlock()
	belogs.Debug("PushCache.foundAllConnAndRrKeys(): subscribedRrs:", jsonutil.MarshalJson(c.subscribedRrs))
	connAndRrKeys = make([]string, 0)
	for k, _ := range c.subscribedRrs {
		connAndRrKeys = append(connAndRrKeys, k)
		belogs.Debug("PushCache.foundAllConnAndRrKeys(): found in rrDelKey, k:", k)
	}
	belogs.Debug("PushCache.foundAllConnAndRrKeys(): connAndRrKeys:", jsonutil.MarshalJson(connAndRrKeys))
	return connAndRrKeys, nil
}
func (c *PushCache) getAllSubscribedRrs() map[string]uint16 {
	c.rrMutex.RLock()
	defer c.rrMutex.RUnlock()
	tmp := make(map[string]uint16, 0)
	for k, v := range c.subscribedRrs {
		tmp[k] = v
	}
	return tmp
}
