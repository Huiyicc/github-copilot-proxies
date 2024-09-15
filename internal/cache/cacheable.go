package cache

type Cacheable interface {
	Set(key string, value interface{}, ttl int) error
	Get(key string) (interface{}, error)
	Exist(key string) (bool, error)
	Del(key string) error
}
