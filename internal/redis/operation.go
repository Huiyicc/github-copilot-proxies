package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

func GetCoon() redis.Conn {
	return pool.Get()
}

func GetOneString(key string) (string, error) {
	redisconn := pool.Get()
	defer redisconn.Close()
	return redis.String(redisconn.Do("get", key))
}

func GetTokenVersion(db, key string) (int64, error) {
	redisconn := pool.Get()
	defer redisconn.Close()
	return redis.Int64(redisconn.Do("get", fmt.Sprintf("Hzer:JWT:Version:%s:%s", db, key)))
}

func SetTokenVersion(db, key string, value, expire int64, cover bool) (err error) {
	redisconn := pool.Get()
	defer redisconn.Close()
	keys := fmt.Sprintf("Hzer:JWT:Version:%s:%s", db, key)
	if !cover {
		var exi bool
		if exi, err = redis.Bool(redisconn.Do("EXISTS", keys)); err != nil {
			return err
		}
		if exi {
			return nil
		}
	}
	if _, err = redisconn.Do("set", keys, value); err != nil {
		return
	}
	_, err = redisconn.Do("expire", keys, expire)
	return
}
