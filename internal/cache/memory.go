package cache

import (
	"fmt"
	"sync"
	"time"
)

// MemoryMap 用于内存缓存
type MemoryMap struct {
	cache       map[string]interface{}
	expirations map[string]int64
	mu          sync.Mutex
}

func NewMemoryMap() *MemoryMap {
	m := &MemoryMap{}
	m.init()
	return m
}

// init 初始化 MemoryMap 的缓存
func (m *MemoryMap) init() {
	m.cache = make(map[string]interface{})
	m.expirations = make(map[string]int64)
}

func (m *MemoryMap) Get(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiration, exists := m.expirations[key]
	currentTime := time.Now().UnixMilli()
	if exists && currentTime > expiration {
		// 键已过期，删除并返回 nil
		fmt.Printf("Get: key=%s has expired, deleting...\n", key)
		delete(m.cache, key)
		delete(m.expirations, key)
		return nil, nil
	}

	value, ok := m.cache[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}

// Set 设置缓存中的值，并指定过期时间（秒）
func (m *MemoryMap) Set(key string, value interface{}, ttl int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[key] = value
	if ttl == 0 {
		// 默认半小时
		ttl = 30 * 60
	}

	if ttl == -1 {
		// -1 表示永久缓存，不设置过期时间
		delete(m.expirations, key)
	} else {
		expiration := time.Now().UnixMilli() + int64(ttl*1000)
		m.expirations[key] = expiration
	}
	return nil
}

func (m *MemoryMap) Exist(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiration, exists := m.expirations[key]
	currentTime := time.Now().UnixMilli()
	if exists && currentTime > expiration {
		// 键已过期，删除并返回 nil
		fmt.Printf("Get: key=%s has expired, deleting...\n", key)
		delete(m.cache, key)
		delete(m.expirations, key)
	}

	_, ok := m.cache[key]
	return ok, nil
}

func (m *MemoryMap) Del(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cache[key]; !ok {
		return nil
	}
	delete(m.cache, key)
	delete(m.expirations, key)
	return nil
}

// 编译时检查
var _ Cacheable = (*MemoryMap)(nil)
