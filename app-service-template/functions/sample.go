// TODO: Change Copyright to your company if open sourcing or remove header
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

package functions

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// TODO: Create your custom type and function(s) and remove these samples

// NewSample ...
// TODO: Add parameters that the function(s) will need each time one is executed
func NewSample() Sample {
	return Sample{}
}

// Sample ...
type Sample struct {
	// TODO: Add properties that the function(s) will need each time one is executed
}

// LogEventDetails is example of processing an Event and passing the original Event to next function in the pipeline
// For more details on the Context API got here: https://docs.edgexfoundry.org/1.3/microservices/application/ContextAPI/
func (s *Sample) LogEventDetails(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("LogEventDetails called in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// Go here for details on Error Handle: https://docs.edgexfoundry.org/1.3/microservices/application/ErrorHandling/
		return false, fmt.Errorf("function LogEventDetails in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, fmt.Errorf("function LogEventDetails in pipeline '%s', type received is not an Event", ctx.PipelineId())
	}

	lc.Infof("Event received in pipeline '%s': ID=%s, Device=%s, and ReadingCount=%d",
		ctx.PipelineId(),
		event.Id,
		event.DeviceName,
		len(event.Readings))
	for index, reading := range event.Readings {
		switch strings.ToLower(reading.ValueType) {
		case strings.ToLower(common.ValueTypeBinary):
			lc.Infof(
				"Reading #%d received in pipeline '%s' with ID=%s, Resource=%s, ValueType=%s, MediaType=%s and BinaryValue of size=`%d`",
				index+1,
				ctx.PipelineId(),
				reading.Id,
				reading.ResourceName,
				reading.ValueType,
				reading.MediaType,
				len(reading.BinaryValue))
		default:
			lc.Infof("Reading #%d received in pipeline '%s' with ID=%s, Resource=%s, ValueType=%s, Value=`%s`",
				index+1,
				ctx.PipelineId(),
				reading.Id,
				reading.ResourceName,
				reading.ValueType,
				reading.Value)
		}
	}

	// Returning true indicates that the pipeline execution should continue with the next function
	// using the event passed as input in this case.
	return true, event
}

// ConvertEventToXML is example of transforming an Event and passing the transformed data to next function in the pipeline
func (s *Sample) ConvertEventToXML(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("ConvertEventToXML called in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		return false, fmt.Errorf("function ConvertEventToXML in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, fmt.Errorf("function ConvertEventToXML in pipeline '%s': type received is not an Event", ctx.PipelineId())
	}

	xml, err := event.ToXML()
	if err != nil {
		return false, fmt.Errorf("function ConvertEventToXML in pipeline '%s': failed to convert event to XML", ctx.PipelineId())
	}

	// Example of DEBUG message which by default you don't want to be logged.
	//     To see debug log messages, Set WRITABLE_LOGLEVEL=DEBUG environment variable or
	//     change LogLevel in configuration.toml before running app service.
	lc.Debugf("Event converted to XML in pipeline '%s': %s", ctx.PipelineId(), xml)

	// Returning true indicates that the pipeline execution should continue with the next function
	// using the event passed as input in this case.
	return true, xml
}

// OutputXML is an example of processing transformed data
func (s *Sample) OutputXML(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("OutputXML called in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		return false, fmt.Errorf("function OutputXML in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	xml, ok := data.(string)
	if !ok {
		return false, fmt.Errorf("function ConvertEventToXML in pipeline '%s': type received is not an string", ctx.PipelineId())
	}

	lc.Debugf("Outputting the following XML in pipeline '%s': %s", ctx.PipelineId(), xml)

	// This sends the XML as a response. i.e. publish for MessageBus/MQTT triggers as configured or
	// HTTP response to for the HTTP Trigger
	// For more details on the SetResponseData() function go here: https://docs.edgexfoundry.org/1.3/microservices/application/ContextAPI/#complete
	ctx.SetResponseData([]byte(xml))
	ctx.SetResponseContentType(common.ContentTypeXML)

	// Returning false terminates the pipeline execution, so this should be last function specified in the pipeline,
	// which is typical in conjunction with usage of .SetResponseData() function.
	return false, nil
}
