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
	plugin    Plugin = &fluentPlugin{}
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "redis", "Redis Output Plugin.")
}

type logmessage struct {
	data []byte
}

type Plugin interface {
	Environment(ctx unsafe.Pointer, key string) string
	Unregister(ctx unsafe.Pointer)
	GetRecord(dec *output.FLBDecoder) (ret int, ts interface{}, rec map[interface{}]interface{})
	NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder
	Send(values []*logmessage) error
	Exit(code int)
}

type fluentPlugin struct{}

func (p *fluentPlugin) Environment(ctx unsafe.Pointer, key string) string {
	return output.FLBPluginConfigKey(ctx, key)
}

func (p *fluentPlugin) Unregister(ctx unsafe.Pointer) {
	output.FLBPluginUnregister(ctx)
}

func (p *fluentPlugin) GetRecord(dec *output.FLBDecoder) (int, interface{}, map[interface{}]interface{}) {
	return output.GetRecord(dec)
}

func (p *fluentPlugin) NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder {
	return output.NewDecoder(data, int(length))
}

func (p *fluentPlugin) Exit(code int) {
	os.Exit(code)
}

func (p *fluentPlugin) Send(values []*logmessage) error {
	return rc.send(values)
}

//export FLBPluginInit
// ctx (context) pointer to fluentbit context (state/ c code)
func FLBPluginInit(ctx unsafe.Pointer) int {
	hosts := plugin.Environment(ctx, "Hosts")
	password := plugin.Environment(ctx, "Password")
	key := plugin.Environment(ctx, "Key")
	db := plugin.Environment(ctx, "DB")
	usetls := plugin.Environment(ctx, "UseTLS")
	tlsskipverify := plugin.Environment(ctx, "TLSSkipVerify")

	// create a pool of redis connection pools
	config, err := getRedisConfig(hosts, password, db, usetls, tlsskipverify, key)
	if err != nil {
		fmt.Printf("configuration errors: %v\n", err)
		// FIXME use fluent-bit method to err in init
		plugin.Unregister(ctx)
		plugin.Exit(1)
		return output.FLB_ERROR
	}
	rc = &redisClient{
		pools: newPoolsFromConfig(config),
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
	dec := plugin.NewDecoder(data, int(length))

	// Iterate Records

	var logs []*logmessage

	for {
		// Extract Record
		ret, ts, record = plugin.GetRecord(dec)
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

	err := plugin.Send(logs)
	if err != nil {
		fmt.Printf("%v\n", err)
		return output.FLB_RETRY
	}

	fmt.Printf("pushed %d logs\n", len(logs))

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
