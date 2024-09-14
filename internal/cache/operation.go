package cache

import (
	"os"
)

var cache Cacheable

func init() {
	cacheImp := "memory"

	switch cacheImp {
	case "redis":
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		psw := os.Getenv("REDIS_PASSWORD")
		instance := NewRedisInstance(host, port, psw)
		cache = instance
	case "memory":
		cache = NewMemoryMap()
	default:
		cache = NewMemoryMap()
	}
}
func Set(key string, value interface{}, ttl int) error {
	return cache.Set(key, value, ttl)
}
func Get(key string) (interface{}, error) {
	return cache.Get(key)
}
func Exist(key string) (bool, error) {
	return cache.Exist(key)
}
func Del(key string) error {
	return cache.Del(key)
}
