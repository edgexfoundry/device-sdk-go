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

package transforms

import (
	"encoding/json"
	"encoding/xml"
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Conversion houses various built in conversion transforms (XML, JSON, CSV)
type Conversion struct {
}

// TransformToXML ...
func (f Conversion) TransformToXML(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, stringType interface{}) {
	if len(params) < 1 {
		return false, errors.New("No Event Received")
	}
	edgexcontext.LoggingClient.Debug("Transforming to XML")
	if result, ok := params[0].(models.Event); ok {
		b, err := xml.Marshal(result)
		if err != nil {
			// LoggingClient.Error(fmt.Sprintf("Error parsing XML. Error: %s", err.Error()))
			return false, errors.New("Incorrect type received, expecting models.Event")
		}
		// should we return a byte[] or string?
		// return b
		return true, string(b)
	}
	return false, errors.New("Unexpected type received")
}

// TransformToJSON ...
func (f Conversion) TransformToJSON(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, stringType interface{}) {
	if len(params) < 1 {
		return false, errors.New("No Event Received")
	}
	edgexcontext.LoggingClient.Debug("Transforming to JSON")
	if result, ok := params[0].(models.Event); ok {
		b, err := json.Marshal(result)
		if err != nil {
			// LoggingClient.Error(fmt.Sprintf("Error parsing JSON. Error: %s", err.Error()))
			return false, errors.New("Error marshalling JSON")
		}
		// should we return a byte[] or string?
		// return b
		return true, string(b)
	}
	return false, errors.New("Unexpected type received")
}
