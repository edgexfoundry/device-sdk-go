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

package runtime

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/context"
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
