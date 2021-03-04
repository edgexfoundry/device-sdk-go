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
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// This file contains example of how to unit test pipeline functions
// TODO: Change these sample unit tests to test your custom type and function(s)

func TestSample_LogEventDetails(t *testing.T) {
	expectedEvent := createTestEvent(t)
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualEvent := target.LogEventDetails(createTestAppSdkContext(), expectedEvent)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedEvent, actualEvent)
}

func TestSample_ConvertEventToXML(t *testing.T) {
	event := createTestEvent(t)
	expectedXml, _ := event.ToXML()
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualXml := target.ConvertEventToXML(createTestAppSdkContext(), event)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedXml, actualXml)

}

func TestSample_OutputXML(t *testing.T) {
	testEvent := createTestEvent(t)
	expectedXml, _ := testEvent.ToXML()
	expectedContinuePipeline := false
	appContext := createTestAppSdkContext()

	target := NewSample()
	actualContinuePipeline, result := target.OutputXML(appContext, expectedXml)
	actualXml := string(appContext.OutputData)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Nil(t, result)
	assert.Equal(t, expectedXml, actualXml)
}

func createTestEvent(t *testing.T) dtos.Event {
	profileName := "MyProfile"
	deviceName := "MyDevice"
	sourceName := "MySource"
	resourceName := "MyResource"

	event := dtos.NewEvent(profileName, deviceName, sourceName)
	err := event.AddSimpleReading(resourceName, v2.ValueTypeInt32, int32(1234))
	require.NoError(t, err)

	event.Tags = map[string]string{
		"WhereAmI": "NotKansas",
	}

	return event
}

func createTestAppSdkContext() *appcontext.Context {
	return &appcontext.Context{
		CorrelationID: uuid.New().String(),
		LoggingClient: logger.NewMockClient(),
	}
}
