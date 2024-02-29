package cache

import (
	"container/list"
	"encoding/json"
	"sync"
	"time"
	"unsafe"
)

type MemoryCache struct {
	maxBytes  int64                    // 最大缓存
	usedBytes int64                    // 已使用缓存
	list      *list.List               // 双链表，lru缓存策略使用
	data      map[string]*list.Element // 实际缓存数据
	expMap    map[int64][]interface{}  // 过期时间与key数组
	stop      chan struct{}            // 停止
	lock      *sync.Mutex              // 锁
}

// size 是⼀个字符串。⽀持以下参数: 1KB，100KB，1MB，2MB，1GB 等
func (m *MemoryCache) SetMaxMemory(size string) bool {
	maxBytes, err := ToBytes(size)
	if err != nil {
		return false
	}
	m.maxBytes = maxBytes
	return true
}

// 设置⼀个缓存项，并且在expire时间之后过期
func (m *MemoryCache) Set(key string, val interface{}, expire ...time.Duration) {
	bytes, _ := json.Marshal(val)
	var exp int64 = forever // 默认 -1 永久
	if len(expire) > 0 {
		exp = int64(expire[0]) + time.Now().Unix()
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if ele, ok := m.data[key]; ok {
		m.list.MoveToFront(ele)
		ent := ele.Value.(*entry)
		m.usedBytes += int64(len(bytes)) - int64(len(ent.value.([]byte)))
		// 删除旧过期时间
		m.delExpMap(ent)
		ent.value = bytes
		ent.exp = exp
	} else {
		ele = m.list.PushFront(&entry{
			value: bytes,
			key:   key,
			exp:   exp,
		})
		m.data[key] = ele
		m.usedBytes += int64(len(key)) + int64(len(bytes)) + int64(unsafe.Sizeof(exp))
	}
	// 添加新过期时间
	m.expMap[exp] = append(m.expMap[exp], key)
	// 判断溢出
	if m.maxBytes < m.usedBytes {
		m.lruRemove()
	}
}

// 获取⼀个值
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if ele, ok := m.data[key]; ok {
		m.list.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

// 删除⼀个值
func (m *MemoryCache) Del(key string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if ele, ok := m.data[key]; ok {
		m.list.Remove(ele)
		delete(m.data, key)
		m.delExpMap(ele.Value.(*entry))
		return true
	}
	return false
}

// 检测⼀个值 是否存在
func (m *MemoryCache) Exists(key string) bool {
	m.lock.Lock()
	m.lock.Unlock()
	_, ok := m.data[key]
	return ok
}

// 查询剩余时间
func (m *MemoryCache) Ttl(key string) int64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	if ele, ok := m.data[key]; ok {
		exp := ele.Value.(*entry).exp
		if exp == forever {
			return exp
		}
		return exp - time.Now().Unix()
	}
	return notExist
}

// 返回所有的key 多少
func (m *MemoryCache) Keys() int64 {
	m.lock.Lock()
	m.lock.Unlock()
	return int64(len(m.data))
}

// 关闭
func (m *MemoryCache) Close() {
	m.stop <- struct{}{}
	m.list = nil
	m.data = nil
	m.expMap = nil
}
