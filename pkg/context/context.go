package context

import (
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/configuration"
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/trigger"
)

// Context ...
type Context struct {
	Trigger       trigger.ITrigger
	Configuration configuration.Configuration
}

// Complete called when ready to send output and function is finished
func (context Context) Complete(output string) {
	(context.Trigger).Complete(output)
}
