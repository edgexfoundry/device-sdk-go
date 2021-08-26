//
// Copyright (c) 2021 Intel Corporation
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
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// Conversion houses various built in conversion transforms (XML, JSON, CSV)
type Conversion struct {
}

// NewConversion creates, initializes and returns a new instance of Conversion
func NewConversion() Conversion {
	return Conversion{}
}

// TransformToXML transforms an EdgeX event to XML.
// It will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Conversion) TransformToXML(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, stringType interface{}) {
	if data == nil {
		return false, fmt.Errorf("function TransformToXML in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	ctx.LoggingClient().Debugf("Transforming to XML in pipeline '%s'", ctx.PipelineId())

	if event, ok := data.(dtos.Event); ok {
		xml, err := event.ToXML()
		if err != nil {
			return false, fmt.Errorf("unable to marshal Event to XML in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}

		ctx.SetResponseContentType(common.ContentTypeXML)
		return true, xml
	}

	return false, fmt.Errorf("function TransformToXML in pipeline '%s': unexpected type received", ctx.PipelineId())
}

// TransformToJSON transforms an EdgeX event to JSON.
// It will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Conversion) TransformToJSON(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, stringType interface{}) {
	if data == nil {
		return false, fmt.Errorf("function TransformToJSON in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	ctx.LoggingClient().Debugf("Transforming to JSON in pipeline '%s'", ctx.PipelineId())

	if result, ok := data.(dtos.Event); ok {
		b, err := json.Marshal(result)
		if err != nil {
			return false, fmt.Errorf("unable to marshal Event to JSON in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}
		ctx.SetResponseContentType(common.ContentTypeJSON)
		// should we return a byte[] or string?
		// return b
		return true, string(b)
	}

	return false, fmt.Errorf("function TransformToJSON in pipeline '%s': unexpected type received", ctx.PipelineId())
}
