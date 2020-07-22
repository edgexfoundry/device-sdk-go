/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package models

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/google/uuid"
)

var TestObjectIDNil = primitive.NilObjectID

var TestUUIDValid = uuid.New().String()
var TestPayload = []byte("brandon wrote this")

const (
	TestUUIDNil          = ""
	TestAppServiceKey    = "apps"
	TestRetryCount       = 2
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
	TestEventID          = "probably"
	TestEventChecksum    = "failed :("
)

var TestModelNoID = StoredObject{
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestModelUUID = StoredObject{
	ObjectID:         TestObjectIDNil,
	UUID:             TestUUIDValid,
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractUUID = contracts.StoredObject{
	ID:               TestUUIDValid,
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractBadID = contracts.StoredObject{
	ID:               "brandon!",
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractNilID = contracts.StoredObject{
	ID:               primitive.NilObjectID.Hex(),
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

func TestFromContract(t *testing.T) {
	tests := []struct {
		testName       string
		fromContract   contracts.StoredObject
		expectedResult StoredObject
		expectedError  bool
	}{
		{
			"Success, UUID",
			TestContractUUID,
			TestModelUUID,
			false,
		},
		{
			"Bad ID",
			TestContractBadID,
			StoredObject{},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			actual := StoredObject{}
			err := actual.FromContract(test.fromContract)
			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedResult, actual)
		})
	}
}

func TestToContract(t *testing.T) {
	tests := []struct {
		testName       string
		fromModel      StoredObject
		expectedResult contracts.StoredObject
	}{
		{
			"Success, UUID",
			TestModelUUID,
			TestContractUUID,
		},
		{
			"No ID",
			TestModelNoID,
			TestContractNilID,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			actual := test.fromModel.ToContract()

			if !reflect.DeepEqual(actual, test.expectedResult) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, actual)
			}
		})
	}
}
