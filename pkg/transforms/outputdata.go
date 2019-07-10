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
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

// OutputData houses transform for outputting data to configured trigger response, i.e. message bus
type OutputData struct {
}

func (f OutputData) SetOutputData(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Debug("Setting output data")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	// By setting this the data will be posted back to to configured trigger response, i.e. message bus
	edgexcontext.OutputData = data

	return true, params[0]
}
