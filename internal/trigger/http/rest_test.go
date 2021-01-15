//
// Copyright (c) 2020 Technotects
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

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestTriggerInitializeWitBackgroundChannel(t *testing.T) {
	background := make(chan types.MessageEnvelope)
	trigger := Trigger{}

	deferred, err := trigger.Initialize(nil, nil, background)

	assert.Nil(t, deferred)
	assert.NotNil(t, err)
	assert.Equal(t, "background publishing not supported for services using HTTP trigger", err.Error())
}
