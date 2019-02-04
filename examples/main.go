package main

import (
	"fmt"

	edgexsdk "github.com/edgexfoundry-holdings/app-functions-sdk-go"
)

func main() {
	deviceIDs := []string{"GS1-AC-Drive01"}
	edgexsdk := &edgexsdk.AppFunctionsSDK{}
	edgexsdk.SetPipeline(
		// edgexsdk.FilterByDeviceID(deviceIDs),
		edgexsdk.TransformToXML(),
		edgexsdk.FilterByDeviceID(deviceIDs),
		myFunc,
	)
	edgexsdk.MakeItRun()
}

func myFunc(params ...interface{}) interface{} { // context context.Context, event event.Event) {
	fmt.Println("HELLO WORLD")

	// context.Complete("OUTPUT FROM FUNCTION")
	return nil
}
