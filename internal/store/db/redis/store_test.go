// +build redisRunning

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

// This test will only be executed if the tag redisRunning is added when running
// the tests with a command like:
// go test -tags redisRunning

package redis

import (
	"testing"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"

	"github.com/google/uuid"
)

const (
	TestHost      = "localhost"
	TestPort      = 6379
	TestTimeout   = "5s"
	TestMaxIdle   = 5000
	TestBatchSize = 1337

	TestRetryCount       = 100
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
	TestPipelineId       = "test-pipeline"
)

var TestPayload = []byte("brandon was here")

var TestValidNoAuthConfig = db.DatabaseInfo{
	Type:      db.RedisDB,
	Host:      TestHost,
	Port:      TestPort,
	Timeout:   TestTimeout,
	MaxIdle:   TestMaxIdle,
	BatchSize: TestBatchSize,
}

var TestContractBase = contracts.StoredObject{
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractBadID = contracts.StoredObject{
	ID:               "brandon!",
	AppServiceKey:    "brandon!",
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoAppServiceKey = contracts.StoredObject{
	ID:               uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoPayload = contracts.StoredObject{
	AppServiceKey:    uuid.New().String(),
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoVersion = contracts.StoredObject{
	AppServiceKey:    uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	CorrelationID:    TestCorrelationID,
}

func TestClient_NewClient(t *testing.T) {
	tests := []struct {
		name          string
		config        db.DatabaseInfo
		expectedError bool
	}{
		{"Success, no auth", TestValidNoAuthConfig, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewClient(test.config, bootstrapConfig.Credentials{})

			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Store(t *testing.T) {
	TestContractUUID := TestContractBase
	TestContractUUID.ID = uuid.New().String()
	TestContractUUID.AppServiceKey = uuid.New().String()

	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()

	TestContractNoAppServiceKey := TestContractBase
	TestContractUUID.ID = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig, bootstrapConfig.Credentials{})

	tests := []struct {
		name          string
		toStore       contracts.StoredObject
		expectedError bool
	}{
		{
			"Success, no ID",
			TestContractValid,
			false,
		},
		{
			"Success, no ID double store",
			TestContractValid,
			false,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no app service key double store",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Success, object with UUID",
			TestContractUUID,
			false,
		},
		{
			"Failure, object with UUID double store",
			TestContractUUID,
			true,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
		{
			"Failure, bad ID",
			TestContractBadID,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			returnVal, err := client.Store(test.toStore)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			require.NotEqual(t, "", returnVal, "Function did not error but did not return a valid ID")
		})
	}
}

func TestClient_RetrieveFromStore(t *testing.T) {
	UUIDAppServiceKey := uuid.New().String()

	UUIDContract0 := TestContractBase
	UUIDContract0.ID = uuid.New().String()
	UUIDContract0.AppServiceKey = UUIDAppServiceKey

	UUIDContract1 := TestContractBase
	UUIDContract1.ID = uuid.New().String()
	UUIDContract1.AppServiceKey = UUIDAppServiceKey

	client, _ := NewClient(TestValidNoAuthConfig, bootstrapConfig.Credentials{})

	tests := []struct {
		name          string
		toStore       []contracts.StoredObject
		key           string
		expectedError bool
	}{
		{
			"Success, single object",
			[]contracts.StoredObject{UUIDContract0},
			UUIDAppServiceKey,
			false,
		},
		{
			"Success, multiple object",
			[]contracts.StoredObject{UUIDContract0, UUIDContract1},
			UUIDAppServiceKey,
			false,
		},
		{
			"Failure, no app service key",
			[]contracts.StoredObject{},
			"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, object := range test.toStore {
				_, _ = client.Store(object)
			}

			actual, err := client.RetrieveFromStore(test.key)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, len(actual), len(test.toStore), "Returned slice length doesn't match expected")
		})
	}
}

func TestClient_Update(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()
	TestContractValid.Version = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig, bootstrapConfig.Credentials{})

	// add the objects we're going to update in the database now so we have a known state
	TestContractValid.ID, _ = client.Store(TestContractValid)

	tests := []struct {
		name          string
		expectedVal   contracts.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no UUID",
			TestContractBase,
			true,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := client.Update(test.expectedVal)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			// only do a lookup on tests that we aren't expecting errors
			actual, _ := client.RetrieveFromStore(test.expectedVal.AppServiceKey)
			require.NotNil(t, actual, "No objects retrieved from store")
			require.Equal(t, test.expectedVal, actual[0], "Return value doesn't match expected")
		})
	}
}

func TestClient_RemoveFromStore(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig, bootstrapConfig.Credentials{})

	// add the objects we're going to update in the database now so we have a known state
	TestContractValid.ID, _ = client.Store(TestContractValid)

	tests := []struct {
		name          string
		testObject    contracts.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := client.RemoveFromStore(test.testObject)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			// only do a lookup on tests that we aren't expecting errors
			actual, _ := client.RetrieveFromStore(test.testObject.AppServiceKey)
			require.Nil(t, actual, "Object retrieved, should have been nil")
		})
	}
}
