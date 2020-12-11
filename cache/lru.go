package cache

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type LRUCache struct {
	lock         sync.RWMutex
	wg           sync.WaitGroup
	Items        map[string]*LRUItem
	Namespaces   []string
	Length       int64
	Size         int64
	Period       int
	evictionList *DLL
	shutdown     chan bool
}

type LRUItem struct {
	Value        []byte
	Ttl          int64
	Timestamp    int64
	evictionNode *DLLNode
}

func Sizeof(item *LRUItem) (uintptr, error) {
	var size uintptr
	size += uintptr(len(item.Value)) + reflect.Indirect(reflect.ValueOf(item.Ttl)).Type().Size() + reflect.Indirect(reflect.ValueOf(item.Timestamp)).Type().Size() + reflect.Indirect(reflect.ValueOf(item.evictionNode)).Type().Size()
	return size, nil
}

func NewLRUCache(size int64) *LRUCache {
	cache := &LRUCache{
		lock:         sync.RWMutex{},
		wg:           sync.WaitGroup{},
		Items:        make(map[string]*LRUItem),
		Namespaces:   append([]string{}, "default"),
		Length:       0,
		Size:         size,
		Period:       0,
		evictionList: NewDLL(EvictionCapacity),
	}
	return cache
}

func (cache *LRUCache) Get(key string, Value interface{}) error {
	return cache.get(key, Value)
}

func (cache *LRUCache) Set(key string, Value interface{}, ttl int64) error {
	return cache.set(key, Value, ttl)
}

func (cache *LRUCache) Remove(key string) error {
	return cache.evict(key, true)
}

func (cache *LRUCache) Purge() error {
	// For now, this needs to be done in a single threaded fashion, since concurrent map access is forbidden
	// TODO: Have a map per namespace, so that the entire eviction can be done concurrently as a function of the namespace
	cache.lock.Lock()
	defer cache.lock.Unlock()

	for i := 0; i < len(cache.Namespaces); i++ {
		evicted, err := cache.clearNamespace(cache.Namespaces[i])
		cache.Length -= evicted
		if err != nil {
			return err
		}
	}
	err := cache.evictionList.clear()
	return err
}

func (cache *LRUCache) Keys() ([]string, error) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	keys := make([]string, 0)
	for key := range cache.Items {
		keys = append(keys, key)
	}
	return keys, nil
}

func (cache *LRUCache) Len() (int64, error) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	return cache.Length, nil
}

func (cache *LRUCache) Stats() (string, error) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	stats := "Number of Keys:" + strconv.Itoa(int(cache.Length)) + "\n" + "Size:" + strconv.Itoa(int(cache.Size)) + "\n"
	stats += "Eviction Size: " + strconv.Itoa(cache.evictionList.Size) + " "
	stats += "Eviction list:'"
	for i, temp := 0, cache.evictionList.list; i < cache.evictionList.Size; i, temp = i+1, temp.next {
		stats += " (Key): " + temp.value + ","
	}
	stats += "'\n"
	usage, _ := cache.getMemoryUsage()
	stats += "Memory Usage: " + strconv.Itoa(int(usage))
	return stats, nil
}

func (cache *LRUCache) evict(key string, useLock bool) error {
	// Evicts an item from the cache. Entries are removed from the evictionList as well as from the hashmap
	if useLock == true {
		// Otherwise, we'll get a deadlock situation when get() calls evict()
		cache.lock.Lock()
		defer cache.lock.Unlock()
	}
	err := cache.evictionList.removeNode(cache.Items[key].evictionNode)
	if err != nil {
		return err
	}
	delete(cache.Items, key)
	return nil
}

func (cache *LRUCache) clearNamespace(namespace string) (int64, error) {
	// Assuming that the function calling this already has acquired the mutex write lock
	prefix := namespace + ":"
	var evicted int64
	for key := range cache.Items {
		if key[:len(prefix)] == prefix {
			err := cache.evict(key, false)
			if err != nil {
				return evicted, err
			}
			evicted++
		}
	}
	return evicted, nil
}

func (cache *LRUCache) get(key string, Value interface{}) error {
	// This will try to fetch from the cache, also evicting a key if it has expired
	cache.lock.RLock()
	var writeLock bool
	defer func() {
		// This will defer for unlocking either the Read-only lock or the RW lock
		if writeLock {
			cache.lock.Unlock()
		} else {
			cache.lock.RUnlock()
		}
	}()

	currTime := makeTimestamp()
	_key, err := cache.makeKey(key, cache.Namespaces[0])
	if err != nil {
		return err
	}
	entry, ok := cache.Items[_key]
	if ok == false {
		Value = nil
		return nil
	} else {
		if (entry.Ttl != 0) && ((currTime - entry.Timestamp) > 1000*entry.Ttl) {
			// Expired. Evict the key
			// Remove the R-lock and place a W-lock
			writeLock = true
			cache.lock.RUnlock()
			cache.lock.Lock()

			err := cache.evict(_key, false)
			cache.Length--
			if err != nil {
				return err
			} else {
				// Set the nil Value to the interface
				Value = nil
				return nil
			}
		}
	}
	err = msgpack.Unmarshal(entry.Value, Value)
	if err != nil {
		return err
	}

	if cache.evictionList.Size >= cache.evictionList.Capacity {
		err = cache.evictionList.removeTail()
		if err != nil {
			return err
		}
	}
	if entry.evictionNode == nil {
		entry.evictionNode = &DLLNode{
			value: key,
			prev:  nil,
			next:  nil,
		}
	}
	err = cache.evictionList.shiftToHead(entry.evictionNode, false)
	return err
}

func (cache *LRUCache) set(key string, Value interface{}, ttl int64) error {
	// Sets an item with a given TTL in seconds. Items are lazily evicted, only on the next get
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if ttl < 0 {
		ttl = DefaultTTLCache
	}
	_, ok := cache.Items[key]
	byteValue, err := msgpack.Marshal(Value)
	if err != nil {
		return err
	}
	key, err = cache.makeKey(key, cache.Namespaces[0])
	if err != nil {
		return err
	}
	cache.Items[key] = &LRUItem{
		Value:     byteValue,
		Ttl:       ttl,
		Timestamp: makeTimestamp(),
		evictionNode: &DLLNode{
			value: key,
			prev:  nil,
			next:  nil,
		},
	}
	node := cache.Items[key].evictionNode
	if ok == false {
		cache.Length++
	} else {
		if cache.Length >= cache.Size {
			return errors.New("cache is full")
		}
	}
	if cache.evictionList.Size >= cache.evictionList.Capacity {
		err = cache.evictionList.removeTail()
		if err != nil {
			return err
		}
	} else {
		err = cache.evictionList.shiftToHead(node, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cache *LRUCache) expired(key string) bool {
	if item, ok := cache.Items[key]; ok == true {
		if (item.Ttl != 0) && ((makeTimestamp() - item.Timestamp) > 1000*item.Ttl) {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (cache *LRUCache) makeKey(key string, namespace string) (string, error) {
	// Constructs the key name, given the namespace name
	cacheKey := namespace + ":" + key
	return cacheKey, nil
}

func (cache *LRUCache) getMemoryUsage() (uintptr, error) {
	// Gets the total memory usage in bytes
	usage := uintptr(0)
	for key, Value := range cache.Items {
		size, err := Sizeof(Value)
		if err != nil {
			return 0, err
		}
		size += uintptr(len(key))
		usage += size
	}
	for curr, node := 0, cache.evictionList.list; curr < cache.evictionList.Size; curr, node = curr+1, node.next {
		usage += reflect.Indirect(reflect.ValueOf(node)).Type().Size() + uintptr(len(node.value)) + reflect.Indirect(reflect.ValueOf(node.next)).Type().Size() + reflect.Indirect(reflect.ValueOf(node.prev)).Type().Size()
	}
	return usage, nil
}

func (cache *LRUCache) Cron() {
	if cache.Period < 1 {
		return
	}
	cache.wg.Add(1)
	for {
		select {
		case status := <-cache.shutdown:
			if status == true {
				cache.wg.Done()
				return
			}
		case <-time.After(time.Duration(cache.Period) * time.Millisecond * 1000):
			// Check for expiration of all keys
			cache.lock.Lock()
			for key := range cache.Items {
				if cache.expired(key) == true {
					cache.evict(key, false)
				}
			}
			cache.lock.Unlock()
		}
	}
}

func (cache *LRUCache) Start(period int) error {
	cache.Period = period
	cache.shutdown = make(chan bool, 1)
	go cache.Cron()
	return nil
}

func (cache *LRUCache) Stop() {
	cache.shutdown <- true
	close(cache.shutdown)
	<-time.After(1 * time.Microsecond) // give goroutines time to shutdown
}

func makeTimestamp() int64 {
	// Return the current unix timestamp in milliseconds
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
