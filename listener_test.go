package jobq

import (
	"reflect"
	"testing"
)

func Test_makeListener(t *testing.T) {
	conninfo := "test conninfo"
	opts := listenerOpts{
		aliveCheckInterval:   100,
		callback:             nil,
		maxReconnectInterval: 101,
		minReconnectInterval: 102,
	}
	got := makeListener(conninfo, opts)
	if got == nil {
		t.Error("makeListener(); got == nil")
	}
	if got.events == nil {
		t.Error("makeListener(); got.events == nil")
	}
	if !reflect.DeepEqual(got.listenerOpts, opts) {
		t.Errorf("makeListener(); got.listenerOpts %v, want %v", got.listenerOpts, opts)
	}
	if got.conninfo != conninfo {
		t.Errorf("makeListener(); got.conninfo %s, want %s", got.conninfo, conninfo)
	}
}
