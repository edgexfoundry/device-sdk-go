package transformer

import (
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
)

func init() {
	lc := logger.NewClient("command_test", false, "./test.log", "DEBUG")
	common.LoggingClient = lc
}

func TestTransformReadScale(t *testing.T) {
	var val interface{} = int16(100)
	var scale = "0.1"
	var expected = int16(10)

	result, err := transformReadScale(val, scale)

	if result != expected || err != nil {
		t.Errorf("Convert new result(%v) failed, error: %v", val, err)
	}
}
