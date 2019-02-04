package messagebustrigger

import (
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/configuration"
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/runtime"
)

// MessageBusTrigger implements ITrigger to support MessageBusData
type MessageBusTrigger struct {
	Configuration configuration.Configuration
	Runtime       runtime.GolangRuntime
	outputData    interface{}
}

// Initialize ...
func (mb *MessageBusTrigger) Initialize() error {
	return nil
}

// GetConfiguration ...
func (mb *MessageBusTrigger) GetConfiguration() configuration.Configuration {
	//
	return mb.Configuration
}

// GetData ...
func (mb *MessageBusTrigger) GetData() interface{} {
	return "data"
}

// Complete ...
func (mb *MessageBusTrigger) Complete(outputData interface{}) {
	//
	mb.outputData = outputData
}
