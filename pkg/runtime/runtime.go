package runtime

import (
	"fmt"

	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/context"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	Transforms []func(...interface{}) interface{}
}

// ProcessEvent handles processing the event
func (gr GolangRuntime) ProcessEvent(edgexcontext context.Context, event models.Event) error {
	fmt.Println("EVENT PROCESSED BY GO")
	var result interface{}
	for _, trxFunc := range gr.Transforms {
		if result != nil {
			result = trxFunc(result)
		} else {
			result = trxFunc(event)
		}
	}
	return nil
}
