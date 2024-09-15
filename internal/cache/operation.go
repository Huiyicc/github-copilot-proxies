package cache

var cache Cacheable

func init() {
	cache = NewMemoryMap()

	// 已废弃redis缓存实现
	/*host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	psw := os.Getenv("REDIS_PASSWORD")
	cache = NewRedisInstance(host, port, psw)*/
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
