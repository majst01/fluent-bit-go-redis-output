package main

import (
	"C"
	"fmt"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)
import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	rc *redisClient
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "redis", "Redis Output Plugin.")
}

//export FLBPluginInit
// ctx (context) pointer to fluentbit context (state/ c code)
func FLBPluginInit(ctx unsafe.Pointer) int {
	// argument parsing from environment variables
	// REDIS_HOSTS must be in the form "hosta:porta hostb:portb"
	hosts := os.Getenv("REDIS_HOSTS")
	if hosts == "" {
		hosts = "127.0.0.1:6379"
	}
	hostAndPorts := strings.Split(hosts, " ")

	password := os.Getenv("REDIS_PASSWORD")
	key := os.Getenv("REDIS_KEY")
	if key == "" {
		key = "logstash"
	}

	dbValue := os.Getenv("REDIS_DB")
	db, err := strconv.Atoi(dbValue)
	if dbValue != "" && err != nil {
		fmt.Printf("REDIS_DB must be a integer: %v\n", err)
		os.Exit(1)
	}
	usetls, err := strconv.ParseBool(os.Getenv("REDIS_USETLS"))
	if err != nil {
		fmt.Printf("REDIS_USETLS must be a bool: %v\n", err)
		os.Exit(1)
	}
	tlsskipverify, err := strconv.ParseBool(os.Getenv("REDIS_TLSSKIP_VERIFY"))
	if err != nil {
		fmt.Printf("REDIS_TLSSKIP_VERIFY must be a bool: %v\n", err)
		os.Exit(1)
	}

	// create a pool of redis connection pools
	redisPools, err := newPools(hostAndPorts, db, password, usetls, tlsskipverify)
	if err != nil {
		fmt.Printf("cannot create a pool of redis connections: %v\n", err)
		os.Exit(1)
	}

	rc = &redisClient{
		pools: redisPools,
		key:   key,
	}

	fmt.Printf("[out-redis] redis connection to: %s db: %d with key:%s\n", hostAndPorts, db, key)
	return output.FLB_OK
}

//export FLBPluginFlush
// FLBPluginFlush is called from fluent-bit when data need to be sent. is called from fluent-bit when data need to be sent.
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}
	var m map[string]interface{}

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
		// convert timestamp to RFC3339Nano which is logstash format
		timestamp := ts.(output.FLBTime)
		const timeFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
		t, _ := time.Parse(timeFormat, timestamp.String())
		m = make(map[string]interface{})
		m["@timestamp"] = t.UTC().Format(time.RFC3339Nano)
		m["@tag"] = C.GoString(tag)
		for k, v := range record {
			switch t := v.(type) {
			case []byte:
				// prevent encoding to base64
				m[k.(string)] = string(t)
			default:
				m[k.(string)] = v
			}

		}
		js, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("error creating message for REDIS: %s", err)
			return output.FLB_RETRY
		}
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
	rc.pools.closeAll()
	return output.FLB_OK
}

func main() {
}
