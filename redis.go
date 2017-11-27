package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisClient struct {
	key   string
	pools *redisPools
}

type redisPools struct {
	pools []*redis.Pool
}

func (rp *redisPools) getRedisConnectionFromPools() (*redis.Pool, error) {
	// FIXME check for equally used active connections, and if Pool is active and healthy
	next := rand.Intn(len(rp.pools))
	pool := rp.pools[next]
	if pool == nil {
		return nil, fmt.Errorf("Pool is nil in Pools")
	}
	return pool, nil
}

func (rp *redisPools) closeAll() {
	for _, pool := range rp.pools {
		pool.Close()
	}
}

func newPools(hostAndPorts []string, db int, password string, usetls, tlsskipverify bool) (*redisPools, error) {
	pools := make([]*redis.Pool, len(hostAndPorts))
	i := 0
	for _, hostAndPort := range hostAndPorts {
		hostAndPortArray := strings.Split(hostAndPort, ":")
		if len(hostAndPortArray) != 2 {
			return nil, fmt.Errorf("hosts must be in the form host:port but is:%s", hostAndPort)
		}
		host := hostAndPortArray[0]
		port := hostAndPortArray[1]

		pool := newPool(host, port, db, password, usetls, tlsskipverify)
		pools[i] = pool
		i++
	}
	fmt.Printf("pools size: %d.\n", len(pools))
	return &redisPools{
		pools: pools,
	}, nil
}

func newPool(host string, port string, db int, password string, usetls, tlsskipverify bool) *redis.Pool {
	var c redis.Conn
	var err error
	if port == "" {
		port = "6379"
	}
	server := fmt.Sprintf("%s:%s", host, port)
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			if usetls {
				c, err = redis.Dial("tcp", server, redis.DialDatabase(db),
					redis.DialUseTLS(usetls),
					redis.DialTLSSkipVerify(tlsskipverify),
				)
			} else {
				c, err = redis.Dial("tcp", server, redis.DialDatabase(db))
			}
			if err != nil {
				return nil, err
			}
			// In case redis needs authentication
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (r *redisClient) write(value []byte) error {
	pool, err := r.pools.getRedisConnectionFromPools()
	if err != nil {
		return err
	}
	conn := pool.Get()
	defer conn.Close()
	_, err = conn.Do("RPUSH", r.key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", r.key, v, err)
	}
	// fmt.Printf("done with RPUSH %s %s\n", r.key, value)
	return err
}
