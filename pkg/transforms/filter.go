package transforms

import (
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Filter houses various built in filter transforms
type Filter struct {
	DeviceIDs []string
}

// FilterByDeviceID ...
func (f Filter) FilterByDeviceID(params ...interface{}) interface{} {

	println("FILTER BY DEVICEID")

	if len(params) != 1 {
		return nil
	}
	deviceIDs := f.DeviceIDs
	event := params[0].(models.Event)

	for _, devID := range deviceIDs {
		if event.Device == devID {
			// LoggingClient.Debug(fmt.Sprintf("Event accepted: %s", event.Device))
			return event
		}
	}
	return nil
	// fmt.Println(event.Data)
	// edgexcontext.Complete("")
}
