package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisClient struct {
	key   string
	pools *redisPools
}

type redisHost struct {
	hostname string
	port     int
}
type redisConfig struct {
	hosts         []redisHost
	db            int
	password      string
	usetls        bool
	tlsskipverify bool
	key           string
}
type redisPools struct {
	pools []*redis.Pool
}

func (rc *redisConfig) String() string {
	return fmt.Sprintf("hosts:%v db:%d usetls:%t tlsskipverify:%t key:%s", rc.hosts, rc.db, rc.usetls, rc.tlsskipverify, rc.key)
}

func getRedisConfig(hosts, password, db, usetls, tlsskipverify, key string) (*redisConfig, error) {
	rc := &redisConfig{}
	// defaults
	if hosts == "" {
		hosts = "127.0.0.1:6379"
	}
	if usetls == "" {
		usetls = "False"
	}
	if tlsskipverify == "" {
		tlsskipverify = "True"
	}
	if key == "" {
		key = "logstash"
	}

	hostAndPorts := strings.Split(hosts, " ")
	for _, hostAndPort := range hostAndPorts {
		rh := redisHost{}
		if strings.Contains(hostAndPort, ":") {
			hostAndPortArray := strings.Split(hostAndPort, ":")
			if len(hostAndPortArray) != 2 {
				return nil, fmt.Errorf("hosts must be in the form host:port but is:%s", hostAndPort)
			}

			port, err := strconv.Atoi(hostAndPortArray[1])
			if err != nil {
				return nil, fmt.Errorf("port must be numeric:%v", err)
			}
			if port < 0 || port > 65535 {
				return nil, fmt.Errorf("port must between 0-65535 not:%d", port)
			}
			rh.hostname = hostAndPortArray[0]
			rh.port = port
		} else {
			rh.hostname = hostAndPort
			rh.port = 6379
		}
		rc.hosts = append(rc.hosts, rh)
	}

	dbValue, err := strconv.Atoi(db)
	if db != "" && err != nil {
		return nil, fmt.Errorf("db must be a integer: %v", err)
	}
	rc.db = dbValue

	tls, err := strconv.ParseBool(usetls)
	if err != nil {
		return nil, fmt.Errorf("usetls must be a bool: %v", err)
	}
	rc.usetls = tls

	tlsverify, err := strconv.ParseBool(tlsskipverify)
	if err != nil {
		return nil, fmt.Errorf("tlsskipverify must be a bool: %v", err)
	}
	rc.tlsskipverify = tlsverify
	rc.password = password
	rc.key = key

	return rc, nil
}

func (rp *redisPools) getRedisPoolFromPools() (*redis.Pool, error) {
	// FIXME check for equally used active connections, and if Pool is active and healthy
	if len(rp.pools) == 0 {
		return nil, fmt.Errorf("pool is empty")
	}
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

func newPoolsFromConfig(rc *redisConfig) (*redisPools, error) {
	pools := make([]*redis.Pool, len(rc.hosts))
	i := 0
	for _, host := range rc.hosts {
		pool := newPool(host.hostname, host.port, rc.db, rc.password, rc.usetls, rc.tlsskipverify)
		pools[i] = pool
		i++
	}
	return &redisPools{
		pools: pools,
	}, nil
}

func newPool(host string, port int, db int, password string, usetls, tlsskipverify bool) *redis.Pool {
	var c redis.Conn
	var err error
	server := fmt.Sprintf("%s:%d", host, port)
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
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (r *redisClient) write(value []byte) error {
	pool, err := r.pools.getRedisPoolFromPools()
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
