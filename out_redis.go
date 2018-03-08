package main

import (
	"C"
	"fmt"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/json-iterator/go"

	"os"
	"time"
)

var (
	rc   *redisClient
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// both variables are set in Makefile
	revision  string
	builddate string
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "redis", "Redis Output Plugin.")
}

type logmessage struct {
	data []byte
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
	fmt.Printf("[out-redis] build:%s version:%s redis connection to: %s\n", builddate, revision, config)
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

	var logs []*logmessage

	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		// Print record keys and values
		var timeStamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timeStamp = ts.(output.FLBTime).Time
		case uint64:
			timeStamp = time.Unix(int64(t), 0)
		default:
			fmt.Print("given time is not in a known format, defaulting to now.\n")
			timeStamp = time.Now()
		}

		js, err := createJSON(timeStamp, C.GoString(tag), record)
		if err != nil {
			fmt.Printf("%v\n", err)
			// DO NOT RETURN HERE becase one message has an error when json is
			// generated, but a retry would fetch ALL messages again. instead an
			// error should be printed to console
			continue
		}
		logs = append(logs, js)
	}

	err := rc.send(logs)
	if err != nil {
		fmt.Printf("%v\n", err)
		return output.FLB_RETRY
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

func createJSON(timestamp time.Time, tag string, record map[interface{}]interface{}) (*logmessage, error) {
	m := make(map[string]interface{})
	// convert timestamp to RFC3339Nano which is logstash format
	m["@timestamp"] = timestamp.UTC().Format(time.RFC3339Nano)
	m["@tag"] = tag
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
		return nil, fmt.Errorf("error creating message for REDIS: %v", err)
	}
	return &logmessage{data: js}, nil
}

//export FLBPluginExit
func FLBPluginExit() int {
	rc.pools.closeAll()
	return output.FLB_OK
}

func main() {
}
