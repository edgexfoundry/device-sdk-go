package trigger

import (
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/configuration"
)

// ITrigger interface is used to hold event data and allow function to
type ITrigger interface {
	// Initialize performs post creation initializations
	Initialize() error

	// function to call to get current configuration for function
	GetConfiguration() configuration.Configuration
	// function to call to retrieve data from input
	GetData() interface{}

	Complete(interface{})
}
