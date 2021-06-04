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
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStringValuesFormatter_invoke_nil(t *testing.T) {
	var sut StringValuesFormatter

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	ctx.AddValue("key", "value")

	format := "injected-{key}"
	result, err := sut.invoke(format, ctx, nil)

	require.NoError(t, err)
	require.Equal(t, "injected-value", result)
}

func TestStringValuesFormatter_invoke(t *testing.T) {
	var sut StringValuesFormatter = func(s string, functionContext interfaces.AppFunctionContext, i interface{}) (string, error) {
		return "custom-formatted", nil
	}

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	ctx.AddValue("key", "value")

	format := "injected-{key}"
	result, err := sut.invoke(format, ctx, nil)

	require.NoError(t, err)
	require.Equal(t, "custom-formatted", result)
}
