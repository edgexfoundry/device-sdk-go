package runtime

import (
	"testing"

	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/context"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	devID1        = "id1"
	devID2        = "id2"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func TestProcessEventNoTransforms(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	runtime := GolangRuntime{}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be nil since no transforms have been passed")
	}
}
func TestProcessEventOneCustomTransform(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	transform1WasCalled := false
	transform1 := func(params ...interface{}) interface{} {
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}
		transform1WasCalled = true
		return "Hello"
	}
	runtime := GolangRuntime{
		Transforms: []func(params ...interface{}) interface{}{transform1},
	}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
}
func TestProcessEventTwoCustomTransforms(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(params ...interface{}) interface{} {
		transform1WasCalled = true
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}

		return "Transform1Result"
	}
	transform2 := func(params ...interface{}) interface{} {
		transform2WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return "Hello"
	}
	runtime := GolangRuntime{
		Transforms: []func(params ...interface{}) interface{}{transform1, transform2},
	}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
	if transform2WasCalled == false {
		t.Fatal("transform2 should have been called")
	}
}
