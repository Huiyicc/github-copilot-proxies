package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"time"
)

type Redis struct {
	Host string
	Port string
	Psw  string
	Pool *redis.Pool
}

func NewRedisInstance(host string, port string, psw string) *Redis {
	r := &Redis{Host: host, Port: port, Psw: psw}
	r.init()
	return r
}

func (r *Redis) Get(k string) (interface{}, error) {
	return r.getConn().Do("get", k)
}

func (r *Redis) Set(k string, v interface{}, ttl int) error {
	_, err := r.getConn().Do("set", k, v, "EX", ttl)
	return err
}

func (r *Redis) Exist(k string) (bool, error) {
	return redis.Bool(r.getConn().Do("EXISTS", k))
}

func (r *Redis) Del(k string) error {
	_, err := r.getConn().Do("del", k)
	return err
}

func (r *Redis) getConn() redis.Conn {
	return r.Pool.Get()
}

func (r *Redis) init() {
	r.Pool = &redis.Pool{
		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool.
		//最大活跃连接数，0代表无限
		MaxActive: 1000,
		//最大闲置连接数
		// Maximum number of idle connections in the pool.
		MaxIdle: 50,
		//闲置连接的超时时间
		// Close connections after remaining idle for this duration. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		IdleTimeout: time.Second * 100,
		//定义拨号获得连接的函数
		// Dial is an application supplied function for creating and configuring a
		// connection.
		//
		// The connection returned from Dial must not be in a special state
		// (subscribed to pubsub channel, transaction started, ...).
		Dial: func() (redis.Conn, error) {
			port, _ := strconv.Atoi(r.Port)
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", r.Host, port))
			if err != nil {
				return nil, err
			}
			password := r.Psw
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
	}

}
