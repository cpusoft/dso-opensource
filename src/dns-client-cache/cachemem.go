package clientcache

import (
	"fmt"
	"strings"
	"sync"

	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
)

var scm *subscribeCacheMem

type subscribeCacheMem struct {
	mutex             sync.RWMutex `json:"-"`
	subscribeRrModels map[string]*rr.RrModel
}

func initMem() error {
	scm = newSubscribeCacheMem()

	return nil
}
func newSubscribeCacheMem() *subscribeCacheMem {
	c := &subscribeCacheMem{}
	c.subscribeRrModels = make(map[string]*rr.RrModel)
	return c
}

func AddSubscribeRrModel(connKey string, messageId uint16, rrModel *rr.RrModel) {
	key := getKey(connKey, messageId)
	belogs.Debug("AddSubscribeRrModel(): connKey:", connKey, " messageId:", messageId, " key:", key)
	scm.mutex.Lock()
	defer scm.mutex.Unlock()
	scm.subscribeRrModels[key] = rrModel
}

func CheckSubscribeRrModel(connKey string, messageId uint16) bool {
	key := getKey(connKey, messageId)
	belogs.Debug("CheckSubscribeRrModel(): connKey:", connKey, " messageId:", messageId, " key:", key)
	scm.mutex.RLock()
	defer scm.mutex.RUnlock()
	_, ok := scm.subscribeRrModels[key]
	return ok
}

func DelSubscribeRrModelByConnKey(connKey string) {
	belogs.Debug("DelSubscribeRrModelByConnKey(): connKey:", connKey)
	scm.mutex.Lock()
	defer scm.mutex.Unlock()
	for key, _ := range scm.subscribeRrModels {
		if strings.HasPrefix(key, connKey) {
			delete(scm.subscribeRrModels, key)
		}
	}
}

func getKey(connKey string, messageId uint16) string {
	return fmt.Sprintf("%s_%d", connKey, messageId)
}
