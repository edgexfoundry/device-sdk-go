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

// models describes the data types that will be used when storing export data in Mongo.
package models

import (
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StoredObject is the atomic and most abstract description of what is collected by the export store system.
type StoredObject struct {
	// ObjectID uniquely identifies this object in Mongo
	ObjectID primitive.ObjectID `bson:"_id"`

	// UUID uniquely identifies this StoredObject
	UUID string `bson:"uuid"`

	// AppServiceKey identifies the app to which this data belongs.
	AppServiceKey string `bson:"appServiceKey"`

	// Payload is the data to be exported
	Payload []byte `bson:"payload"`

	// RetryCount is how many times this has tried to be exported
	RetryCount int `bson:"retryCount"`

	// PipelinePosition is where to pickup in the pipeline
	PipelinePosition int `bson:"pipelinePosition"`

	// Version is a hash of the functions to know if the pipeline has changed.
	Version string `bson:"version"`

	// CorrelationID is an identifier provided by EdgeX to track this record as it moves
	CorrelationID string `bson:"correlationID"`

	// EventID is used to identify an EdgeX event from the core services and mark it as pushed.
	EventID string `bson:"eventID"`

	// EventChecksum is used to identify CBOR encoded data from the core services and mark it as pushed.
	EventChecksum string `bson:"eventChecksum"`
}

func (o *StoredObject) FromContract(c contracts.StoredObject) error {
	var err error
	o.ObjectID, o.UUID, err = FromContractId(c.ID)
	if err != nil {
		return err
	}

	o.AppServiceKey = c.AppServiceKey
	o.Payload = c.Payload
	o.RetryCount = c.RetryCount
	o.PipelinePosition = c.PipelinePosition
	o.Version = c.Version
	o.CorrelationID = c.CorrelationID
	o.EventID = c.EventID
	o.EventChecksum = c.EventChecksum

	return nil
}

func (o StoredObject) ToContract() contracts.StoredObject {
	contract := contracts.NewStoredObject(
		o.AppServiceKey,
		o.Payload,
		o.PipelinePosition,
		o.Version)

	contract.ID = ToContractId(o.ObjectID, o.UUID)
	contract.RetryCount = o.RetryCount
	contract.CorrelationID = o.CorrelationID
	contract.EventID = o.EventID
	contract.EventChecksum = o.EventChecksum

	return contract
}

func FromContractId(id string) (primitive.ObjectID, string, error) {
	// In this first case, UUID is empty so this must be an add.
	// Generate new BSON/UUIDs
	if id == "" {
		return primitive.NewObjectID(), uuid.New().String(), nil
	}

	// In this case, we're dealing with an existing id
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Id is not a BSON UUID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return primitive.NilObjectID, "", errors.New("invalid id: " + id)
		}
		return primitive.NilObjectID, id, nil
	}

	// UUID of pre-existing event is a BSON UUID. We will query using the BSON UUID.
	return objID, "", nil
}

func ToContractId(id primitive.ObjectID, uuid string) string {
	if uuid != "" {
		return uuid
	}

	return id.Hex()
}
