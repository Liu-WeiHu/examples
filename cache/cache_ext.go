package cache

import (
	"container/list"
	"sync"
	"time"
	"unsafe"
)

type entry struct {
	value interface{}
	key   string
	exp   int64
}

func NewCache() *MemoryCache {
	cache := &MemoryCache{
		list:   list.New(),
		data:   make(map[string]*list.Element),
		expMap: make(map[int64][]interface{}),
		stop:   make(chan struct{}),
		lock:   new(sync.Mutex),
	}
	go cache.expRemove()
	return cache
}

// lru 溢出 删除
func (m *MemoryCache) lruRemove() {
	for m.list.Back() != nil {
		ele := m.list.Back()
		m.list.Remove(ele)
		ent := ele.Value.(*entry)
		delete(m.data, ent.key)
		m.delExpMap(ent)
		m.usedBytes -= int64(len(ent.key)) + int64(len(ent.value.([]byte))) + int64(unsafe.Sizeof(ent.exp))
		if m.maxBytes >= m.usedBytes {
			break
		}
	}
}

// 超时 删除
func (m *MemoryCache) expRemove() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	now := time.Now().Unix()
	type delStruct struct {
		keys []interface{}
		exp  int64
	}
	delCh := make(chan *delStruct, len(m.expMap))

	go func() {
		for v := range delCh {
			delete(m.expMap, v.exp)
			for _, key := range v.keys {
				ele := m.data[key.(string)]
				delete(m.data, key.(string))
				m.list.Remove(ele)
				val := ele.Value.(*entry)
				m.usedBytes -= int64(len(key.(string))) + int64(len(val.value.([]byte))) + int64(unsafe.Sizeof(val.exp))
			}
		}
	}()

	for {
		select {
		case <-m.stop:
			close(delCh)
			return
		case <-ticker.C:
			now++
			if keys, ok := m.expMap[now]; ok {
				delCh <- &delStruct{keys: keys, exp: now}
			}
		}
	}
}

// 删除超时时间
func (m *MemoryCache) delExpMap(ent *entry) {
	expKeys := m.expMap[ent.exp]
	if len(expKeys) > 1 {
		for i := range expKeys {
			if expKeys[i] == ent.key {
				expKeys[i] = expKeys[len(expKeys)-1]
				expKeys = expKeys[:len(expKeys)-1]
				break
			}
		}
		m.expMap[ent.exp] = expKeys
	} else {
		delete(m.expMap, ent.exp)
	}
}
