package transforms

import (
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

type CoreData struct {
}

// NewCoreData Is provided to interact with CoreData
func NewCoreData() *CoreData {
	coredata := &CoreData{}
	return coredata
}

// MarkAsPushed will make a request to CoreData to mark the event that triggered the pipeline as pushed.
// This function will not stop the pipeline if an error is returned from core data, however the error is logged.
func (cdc *CoreData) MarkAsPushed(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	err := edgexcontext.MarkAsPushed()
	if err != nil {
		edgexcontext.LoggingClient.Error(err.Error())
	}
	return true, params[0]
}
