package main

import (
	"testing"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/stretchr/testify/assert"
)

const (
	timeFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
)

func TestCreateJSON(t *testing.T) {
	record := make(map[interface{}]interface{})
	record["key"] = "value"
	record["five"] = 5
	ts, _ := time.Parse(timeFormat, "2006-01-02 15:04:05.999999999 -0700 MST")
	js, err := createJSON(ts, "atag", record)

	if err != nil {
		assert.Fail(t, "it is not expected that the call to createJSON fails:%v", err)
	}
	assert.NotNil(t, js, "json must not be nil")
	result := make(map[string]interface{})
	err = json.Unmarshal(js.data, &result)
	if err != nil {
		assert.Fail(t, "it is not expected that unmarshal of json fails:%v", err)
	}
	assert.Equal(t, result["@timestamp"], "2006-01-02T22:04:05.999999999Z")
	assert.Equal(t, result["@tag"], "atag")
	assert.Equal(t, result["key"], "value")
	assert.Equal(t, result["five"], float64(5))
}

func BenchmarkCreateJSON(b *testing.B) {
	record := make(map[interface{}]interface{})
	record["key"] = "value"
	record["five"] = 5
	ts, _ := time.Parse(time.RFC3339Nano, "2006-01-02 15:04:05.999999999 -0700 MST")
	for i := 0; i < b.N; i++ {
		createJSON(ts, "atag", record)
	}
}

type testFluentPlugin struct {
	hosts string
	db    string
}

func (p *testFluentPlugin) Environment(ctx unsafe.Pointer, key string) string {
	switch key {
	case "Hosts":
		return p.hosts
	case "Password":
		return "mypasswd"
	case "Key":
		return "testkey"
	case "DB":
		return p.db
	case "UseTLS":
		return "false"
	case "TLSSkipVerify":
		return "false"
	}
	return "unknown-" + key
}

func (p *testFluentPlugin) Unregister(ctx unsafe.Pointer)                                 {}
func (p *testFluentPlugin) NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder { return nil }
func (p *testFluentPlugin) Exit(code int)                                                 {}
func (p *testFluentPlugin) Send(values []*logmessage) error                               { return nil }
func (p *testFluentPlugin) GetRecord(dec *output.FLBDecoder) (int, interface{}, map[interface{}]interface{}) {
	return 0, nil, nil
}

func TestPluginInitialization(t *testing.T) {
	plugin = &testFluentPlugin{hosts: "hosta hostb", db: "0"}
	res := FLBPluginInit(nil)
	assert.Equal(t, output.FLB_OK, res)
	assert.Len(t, rc.pools.pools, 2)
}

func TestPluginInitializationFailure(t *testing.T) {
	plugin = &testFluentPlugin{hosts: "hosta hostb", db: "a"}
	res := FLBPluginInit(nil)
	assert.Equal(t, output.FLB_ERROR, res)
}
