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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var precision = 4

func ConvertToReadableFloatValues(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Debug("Convert to Readable Float Values")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	event := params[0].(models.Event)
	for index := range event.Readings {
		eventReading := &event.Readings[index]
		edgexcontext.LoggingClient.Debug(fmt.Sprintf("Event Reading for %s: %s is '%s'", event.Device, eventReading.Name, eventReading.Value))

		data, err := base64.StdEncoding.DecodeString(eventReading.Value)
		if err != nil {
			edgexcontext.LoggingClient.Error(fmt.Sprintf("Unable to Base 64 decode float32/64 value ('%s'): %s", eventReading.Value, err.Error()))
		}

		switch eventReading.Name {
		case "RandomValue_Float32":
			var value float32
			err = binary.Read(bytes.NewReader(data), binary.BigEndian, &value)
			if err != nil {
				edgexcontext.LoggingClient.Error("Unable to decode float32 value bytes" + err.Error())
			}

			eventReading.Value = strconv.FormatFloat(float64(value), 'f', precision, 32)

		case "RandomValue_Float64":
			var value float64
			err := binary.Read(bytes.NewReader(data), binary.BigEndian, &value)
			if err != nil {
				edgexcontext.LoggingClient.Error("Unable to decode float64 value bytes: " + err.Error())
			}

			eventReading.Value = strconv.FormatFloat(value, 'f', precision, 64)
		}

		edgexcontext.LoggingClient.Debug(fmt.Sprintf("Converted Event Reading for %s: %s is '%s'", event.Device, eventReading.Name, eventReading.Value))
	}

	return true, event
}
