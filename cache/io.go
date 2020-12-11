package cache

import (
	"bytes"
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

func (cache *LRUCache) Save(filepath string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	buffer := &bytes.Buffer{}
	if err := save(buffer, cache); err != nil {
		return err
	}
	err := ioutil.WriteFile(filepath, buffer.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadCache(filepath string) (*LRUCache, error) {
	fileBuffer, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	cache := &LRUCache{}
	if err := load(fileBuffer, cache); err != nil {
		return nil, err
	}
	cache.lock = sync.RWMutex{}
	cache.wg = sync.WaitGroup{}
	cache.evictionList = NewDLL(EvictionCapacity)

	for key, value := range cache.Items {
		value.evictionNode = &DLLNode{
			value: &key,
			next:  nil,
			prev:  nil,
		}
	}
	return cache, nil
}

func save(w io.Writer, i interface{}) error {
	return gob.NewEncoder(w).Encode(i)
}

func load(r io.Reader, i interface{}) error {
	return gob.NewDecoder(r).Decode(i)
}
