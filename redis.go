package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisClient struct {
	key  string
	pool *redis.Pool
}

func newPool(host string, port string, db int, password string, usetls, tlsskipverify bool, certificate *x509.Certificate) *redis.Pool {
	var clientTLSConfig tls.Config
	var c redis.Conn
	var err error

	server := fmt.Sprintf("%s:%s", host, port)
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			if usetls {
				clientTLSConfig.RootCAs = x509.NewCertPool()
				clientTLSConfig.RootCAs.AddCert(certificate)
				c, err = redis.Dial("tcp", server, redis.DialDatabase(db),
					redis.DialTLSConfig(&clientTLSConfig),
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
	conn := r.pool.Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", r.key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", r.key, v, err)
	}
	fmt.Printf("done with RPUSH %s %s\n", r.key, value)
	return err
}
