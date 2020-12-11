package cache

const DefaultTTLCache = int64(300) // 300 seconds TTL as default
const EvictionCapacity = 100

type Cache interface {
	Purge() error
	Set(key string, value interface{}, ttl int64) error
	Get(key string, value interface{}) error
	Remove(key string) error
	Keys() ([]string, error)
	Len() (int, error)
	Stats() (string, error)
}
