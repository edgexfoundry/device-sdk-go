//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package functions

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func PrintFloatValuesToConsole(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	event := params[0].(models.Event)

	for _, eventReading := range event.Readings {
		fmt.Printf("%s readable value from %s is %s\n", eventReading.Name, event.Device, eventReading.Value)
	}

	return true, event

}

func Publish(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Debug("Publish")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	event := params[0].(models.Event)
	payload, _ := json.Marshal(event)

	// By calling Complete, the filtered and converted events will be posted back to the message bus on the new topic defined in the configuration.
	edgexcontext.Complete(payload)

	return false, nil
}
