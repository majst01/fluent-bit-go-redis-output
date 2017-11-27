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
	// Example to retrieve an optional configuration parameter
	key := os.Getenv("REDIS_KEY")
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	usetls, _ := strconv.ParseBool(os.Getenv("REDIS_USETLS"))
	tlsskipverify, _ := strconv.ParseBool(os.Getenv("REDIS_TLSSKIP_VERIFY"))

	redisPool := newPool(host, port, db, password, usetls, tlsskipverify)
	rc = &redisClient{
		pool: redisPool,
		key:  key,
	}

	fmt.Printf("[flb-go] redis connection to: %s:%s db: %d with key:%s\n", host, port, db, key)
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
	rc.pool.Close()
	return output.FLB_OK
}

func main() {
}
