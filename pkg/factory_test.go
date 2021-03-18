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

package pkg

import (
	"os"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var expectedLogger logger.LoggingClient
var expectedCorrelationId string

var dic *di.Container

func TestMain(m *testing.M) {
	expectedCorrelationId = uuid.NewString()
	expectedLogger = logger.NewMockClient()

	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return expectedLogger
		},
	})

	os.Exit(m.Run())
}

func TestNewAppFuncContextForTest(t *testing.T) {
	expectedContentType := ""

	target := NewAppFuncContextForTest(expectedCorrelationId, expectedLogger)

	assert.Equal(t, expectedLogger, target.LoggingClient())
	assert.Equal(t, expectedCorrelationId, target.CorrelationID())
	assert.Equal(t, expectedContentType, target.InputContentType())
}
