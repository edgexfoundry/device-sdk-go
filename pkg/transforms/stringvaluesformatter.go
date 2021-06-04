//
// Copyright (c) 2021 Technotects
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
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

// StringValuesFormatter defines a function signature to perform string formatting operations using an AppFunction payload.
type StringValuesFormatter func(string, interfaces.AppFunctionContext, interface{}) (string, error)

// invoke will attempt to invoke the underlying function, returning the result of ctx.ApplyValues(format) if nil
func (f StringValuesFormatter) invoke(format string, ctx interfaces.AppFunctionContext, data interface{}) (string, error) {
	if f == nil {
		return ctx.ApplyValues(format)
	} else {
		return f(format, ctx, data)
	}
}
