package main

import (
	"C"
	"fmt"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/garyburd/redigo/redis"
)
import (
	"encoding/json"
	"os"
	"time"
)

var (
	rc *redisClient
)

type redisClient struct {
	key  string
	pool *redis.Pool
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
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

	reply, err := conn.Do("RPUSH", r.key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", r.key, v, err)
	}
	fmt.Printf("wrote: %v", reply)
	return err
}

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "redis", "Redis Output Plugin.")
}

//export FLBPluginInit
// ctx (context) pointer to fluentbit context (state/ c code)
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Example to retrieve an optional configuration parameter
	key := os.Getenv("REDIS_KEY")
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	redisHost := fmt.Sprintf("%s:%s", host, port)

	rc = &redisClient{
		pool: newPool(redisHost),
		key:  key,
	}

	fmt.Printf("[flb-go] redis connection to: %s with key:%s\n", redisHost, key)
	return output.FLB_OK
}

//export FLBPluginFlush
// FLBPluginFlush is called from fluent-bit when data need to be sent. is called from fluent-bit when data need to be sent.
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	// Iterate Records
	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		// Print record keys and values
		timestamp := ts.(output.FLBTime)
		m := make(map[string]interface{})
		m["@timestamp"] = timestamp.String()
		m["@tag"] = C.GoString(tag)
		for k, v := range record {
			m[k.(string)] = v
		}
		js, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("error creating message for REDIS: %s", err)
			return output.FLB_RETRY
		}
		fmt.Printf("%s\n", js)
		err = rc.write(js)
		if err != nil {
			fmt.Printf("error %v", err)
			return output.FLB_RETRY
		}
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
