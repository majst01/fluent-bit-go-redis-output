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
	hosts := output.FLBPluginConfigKey(ctx, "Hosts")
	password := output.FLBPluginConfigKey(ctx, "Password")
	key := output.FLBPluginConfigKey(ctx, "Key")
	db := output.FLBPluginConfigKey(ctx, "DB")
	usetls := output.FLBPluginConfigKey(ctx, "UseTLS")
	tlsskipverify := output.FLBPluginConfigKey(ctx, "TLSSkipVerify")

	// create a pool of redis connection pools
	config, err := getRedisConfig(hosts, password, db, usetls, tlsskipverify, key)
	if err != nil {
		fmt.Printf("configuration errors: %v\n", err)
		// FIXME use fluent-bit method to err in init
		output.FLBPluginUnregister(ctx)
		os.Exit(1)
	}
	redisPools, err := newPoolsFromConfig(config)
	if err != nil {
		fmt.Printf("cannot create a pool of redis connections: %v\n", err)
		output.FLBPluginUnregister(ctx)
		os.Exit(1)
	}

	rc = &redisClient{
		pools: redisPools,
		key:   config.key,
	}

	fmt.Printf("[out-redis] redis connection to: %s\n", config)
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
			fmt.Printf("error creating message for REDIS: %s\n", err)
			return output.FLB_RETRY
		}
		err = rc.write(js)
		if err != nil {
			fmt.Printf("error %v\n", err)
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
