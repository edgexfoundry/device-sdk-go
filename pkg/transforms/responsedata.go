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
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

// ResponseData houses transform for outputting data to configured trigger response, i.e. message bus
type ResponseData struct {
	ResponseContentType string
}

// NewResponseData creates, initializes and returns a new instance of ResponseData
func NewResponseData() ResponseData {
	return ResponseData{}
}

// SetResponseData sets the response data to that passed in from the previous function.
// It will return an error and stop the pipeline if the input data is not of type []byte, string or json.Marshaller
func (f ResponseData) SetResponseData(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {

	ctx.LoggingClient().Debugf("Setting response data in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// We didn't receive a result
		return false, fmt.Errorf("function SetResponseData in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	if len(f.ResponseContentType) > 0 {
		ctx.SetResponseContentType(f.ResponseContentType)
	}

	// By setting this the data will be posted back to to configured trigger response, i.e. message bus
	ctx.SetResponseData(byteData)

	return true, data
}
