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
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"

	"github.com/google/uuid"
)

const (
	TestHost      = "localhost"
	TestPort      = 6379
	TestTimeout   = 5000
	TestBatchSize = 1337

	TestRetryCount       = 100
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
	TestEventID          = "probably"
	TestEventChecksum    = "failed :("
)

var TestPayload = []byte("brandon was here")

var TestValidNoAuthConfig = db.DatabaseInfo{
	Type:      db.RedisDB,
	Host:      TestHost,
	Port:      TestPort,
	Timeout:   TestTimeout,
	MaxIdle:   TestTimeout,
	BatchSize: TestBatchSize,
}

var TestContractBase = contracts.StoredObject{
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
	AppServiceKey:    "brandon!",
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractNoAppServiceKey = contracts.StoredObject{
	ID:               uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractNoPayload = contracts.StoredObject{
	AppServiceKey:    uuid.New().String(),
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractNoVersion = contracts.StoredObject{
	AppServiceKey:    uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
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
			_, connectErr := NewClient(test.config)

			if test.expectedError && connectErr == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && connectErr != nil {
				t.Fatalf("Unexpectedly encountered error: %s", connectErr.Error())
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

	client, _ := NewClient(TestValidNoAuthConfig)

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
			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && returnVal == "" {
				t.Fatal("Function did not error but did not return a valid ID")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}
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

	client, _ := NewClient(TestValidNoAuthConfig)

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

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			if len(actual) != len(test.toStore) {
				t.Fatalf("Returned slice length doesn't match expected.\nExpected: %v\nActual: %v\n", len(test.toStore), len(actual))
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()
	TestContractValid.Version = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig)

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

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			// only do a lookup on tests that we aren't expecting errors
			if !test.expectedError {
				actual, _ := client.RetrieveFromStore(test.expectedVal.AppServiceKey)
				if actual == nil {
					t.Fatal("No objects retrieved from store")
				}

				if !reflect.DeepEqual(actual[0], test.expectedVal) {
					t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actual[0])
				}
			}
		})
	}
}

func TestClient_RemoveFromStore(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig)

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

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			// only do a lookup on tests that we aren't expecting errors
			if !test.expectedError {
				actual, _ := client.RetrieveFromStore(test.testObject.AppServiceKey)
				if actual != nil {
					t.Fatal("Object retrieved, should have been nil")
				}
			}
		})
	}
}
