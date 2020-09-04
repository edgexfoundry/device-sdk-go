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

package appsdk

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewBackgroundPublisherAndPublish(t *testing.T) {
	background, pub := newBackgroundPublisher(1)

	payload := []byte("something")
	correlationId := "id"
	contentType := "type"

	pub.Publish(payload, correlationId, contentType)

	waiting := true

	for waiting {
		select {
		case msgs := <-background:
			assert.Equal(t, correlationId, msgs.CorrelationID)
			assert.Equal(t, contentType, msgs.ContentType)
			assert.Equal(t, payload, msgs.Payload)
			waiting = false
		case <-time.After(1 * time.Second):
			assert.Fail(t, "message timed out, background channel likely not configured correctly")
			waiting = false
		}
	}
}
