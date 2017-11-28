package main

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestGetRedisConfig(t *testing.T) {
	// test for defaults
	c, err := getRedisConfig("", "", "", "", "", "")
	if err != nil {
		assert.Fail(t, "configuration failed with:%v", err)
	}
	assert.Equal(t, false, c.usetls, "usetls expected to be false by default")
	assert.Equal(t, true, c.tlsskipverify, "tlsskipverify expected to be true by default")
	assert.Equal(t, 0, c.db, "db expected to be 0 by default")
	assert.Equal(t, "", c.password, "password expected to be '' by default")
	assert.Equal(t, "logstash", c.key, "key expected to be 'logstash' by default")
	assert.Equal(t, 1, len(c.hosts), "it is expected to have one host by default")
	assert.Equal(t, "127.0.0.1", c.hosts[0].hostname, "it is expected to have 127.0.0.1 as host by default")
	assert.Equal(t, 6379, c.hosts[0].port, "it is expected to have 6379 as port by default")
	assert.Equal(t, "hosts:[{127.0.0.1 6379}] db:0 usetls:false tlsskipverify:true key:logstash", c.String())

	// valid configuration parameter passed
	c, err = getRedisConfig("", "geheim", "1", "true", "false", "elastic")
	if err != nil {
		assert.Fail(t, "configuration failed with:%v", err)
	}
	assert.Equal(t, true, c.usetls, "usetls expected to be true")
	assert.Equal(t, false, c.tlsskipverify, "tlsskipverify expected to be false")
	assert.Equal(t, 1, c.db, "db expected to be 1")
	assert.Equal(t, "geheim", c.password, "password expected to be 'geheim'")
	assert.Equal(t, "elastic", c.key, "key expected to be 'elastic'")

	// valid configuration for hosts without port
	c, err = getRedisConfig("1.2.3.4", "", "", "", "", "")
	if err != nil {
		assert.Fail(t, "configuration failed with:%v", err)
	}
	assert.Equal(t, 1, len(c.hosts), "it is expected to have one host")
	assert.Equal(t, "1.2.3.4", c.hosts[0].hostname, "it is expected to have 1.2.3.4")
	assert.Equal(t, 6379, c.hosts[0].port, "it is expected to have 6379")

	// valid configuration for hosts with port
	c, err = getRedisConfig("1.2.3.4:42", "", "", "", "", "")
	if err != nil {
		assert.Fail(t, "configuration failed with:%v", err)
	}
	assert.Equal(t, 1, len(c.hosts), "it is expected to have one host")
	assert.Equal(t, "1.2.3.4", c.hosts[0].hostname, "it is expected to have 1.2.3.4")
	assert.Equal(t, 42, c.hosts[0].port, "it is expected to have 6379")

	// valid configuration for hosts with port
	c, err = getRedisConfig("1.2.3.4:42 1.2.3.5", "", "", "", "", "")
	if err != nil {
		assert.Fail(t, "configuration failed with:%v", err)
	}
	assert.Equal(t, 2, len(c.hosts), "it is expected to have two host")
	assert.Equal(t, "1.2.3.4", c.hosts[0].hostname, "it is expected to have 1.2.3.4")
	assert.Equal(t, 42, c.hosts[0].port, "it is expected to have 42")

	assert.Equal(t, "1.2.3.5", c.hosts[1].hostname, "it is expected to have 1.2.3.5")
	assert.Equal(t, 6379, c.hosts[1].port, "it is expected to have 6379")
	assert.Equal(t, "hosts:[{1.2.3.4 42} {1.2.3.5 6379}] db:0 usetls:false tlsskipverify:true key:logstash", c.String())

	// invalid configurations
	c, err = getRedisConfig("", "", "A", "", "", "")
	if err != nil {
		assert.Equal(t, "db must be a integer: strconv.Atoi: parsing \"A\": invalid syntax", err.Error())
	}

	c, err = getRedisConfig("", "", "", "xxx", "", "")
	if err != nil {
		assert.Equal(t, "usetls must be a bool: strconv.ParseBool: parsing \"xxx\": invalid syntax", err.Error())
	}

	c, err = getRedisConfig("", "", "", "", "xxx", "")
	if err != nil {
		assert.Equal(t, "tlsskipverify must be a bool: strconv.ParseBool: parsing \"xxx\": invalid syntax", err.Error())
	}

	c, err = getRedisConfig("ahost:aport", "", "", "", "", "")
	if err != nil {
		assert.Equal(t, "port must be numeric:strconv.Atoi: parsing \"aport\": invalid syntax", err.Error())
	}

	c, err = getRedisConfig("ahost:42:43", "", "", "", "", "")
	if err != nil {
		assert.Equal(t, "hosts must be in the form host:port but is:ahost:42:43", err.Error())
	}

}

func TestGetRedisConnectionFromPools(t *testing.T) {
	pools := []*redis.Pool{}
	rp := &redisPools{
		pools: pools,
	}

	p, err := rp.getRedisPoolFromPools()
	if err != nil {
		assert.Equal(t, "pool is empty", err.Error())
	}

	pool := newPool("1.2.3.5", 6379, 0, "", false, false)
	rp.pools = append(rp.pools, pool)
	p, err = rp.getRedisPoolFromPools()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.NotNil(t, p, "pool is not to be expected nil")

	pool = newPool("1.2.3.4", 42, 0, "", false, false)
	rp.pools = append(rp.pools, pool)
	p1, _ := rp.getRedisPoolFromPools()
	p2, _ := rp.getRedisPoolFromPools()

	assert.NotNil(t, p1, "pool is not to be expected nil")
	assert.NotNil(t, p2, "pool is not to be expected nil")

}
