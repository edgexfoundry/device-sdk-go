//
// Copyright (c) 2020 Technotects
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

package http

import (
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestTriggerInitializeWitBackgroundChannel(t *testing.T) {
	background := make(chan types.MessageEnvelope)
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
	trigger := NewTrigger(dic, nil, nil)

	deferred, err := trigger.Initialize(nil, nil, background)

	assert.Nil(t, deferred)
	assert.NotNil(t, err)
	assert.Equal(t, "background publishing not supported for services using HTTP trigger", err.Error())
}
