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

// models describes the data types that will be used when storing export data in Redis.
package models

import (
	"encoding/json"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
)

// StoredObject is the atomic and most abstract description of what is collected by the export store system.
type StoredObject struct {
	// ID uniquely identifies this StoredObject
	ID string `json:"id"`
	// AppServiceKey identifies the app to which this data belongs.
	AppServiceKey string `json:"appServiceKey"`
	// Payload is the data to be exported
	Payload []byte `json:"payload"`
	// RetryCount is how many times this has tried to be exported
	RetryCount int `json:"retryCount"`
	// PipelineId is the ID of the pipeline that needs to be restarted.
	PipelineId string `json:"pipelineId"`
	// PipelinePosition is where to pickup in the pipeline
	PipelinePosition int `json:"pipelinePosition"`
	// Version is a hash of the functions to know if the pipeline has changed.
	Version string `json:"version"`
	// CorrelationID is an identifier provided by EdgeX to track this record as it moves
	CorrelationID string `json:"correlationID"`
	// ContextData is a snapshot of data used by the pipeline at runtime
	ContextData map[string]string
}

// ToContract builds a contract out of the supplied model.
func (o StoredObject) ToContract() contracts.StoredObject {
	return contracts.StoredObject{
		ID:               o.ID,
		AppServiceKey:    o.AppServiceKey,
		Payload:          o.Payload,
		RetryCount:       o.RetryCount,
		PipelineId:       o.PipelineId,
		PipelinePosition: o.PipelinePosition,
		Version:          o.Version,
		CorrelationID:    o.CorrelationID,
		ContextData:      o.ContextData,
	}
}

// FromContract builds a model out of the supplied contract.
func (o *StoredObject) FromContract(c contracts.StoredObject) {
	o.ID = c.ID
	o.AppServiceKey = c.AppServiceKey
	o.Payload = c.Payload
	o.RetryCount = c.RetryCount
	o.PipelineId = c.PipelineId
	o.PipelinePosition = c.PipelinePosition
	o.Version = c.Version
	o.CorrelationID = c.CorrelationID
	o.ContextData = c.ContextData
}

// MarshalJSON returns the object as a JSON encoded byte array.
func (o StoredObject) MarshalJSON() ([]byte, error) {
	test := struct {
		ID               *string           `json:"id,omitempty"`
		AppServiceKey    *string           `json:"appServiceKey,omitempty"`
		Payload          []byte            `json:"payload,omitempty"`
		RetryCount       int               `json:"retryCount,omitempty"`
		PipelineId       string            `json:"pipelineId,omitempty"`
		PipelinePosition int               `json:"pipelinePosition,omitempty"`
		Version          *string           `json:"version,omitempty"`
		CorrelationID    *string           `json:"correlationID,omitempty"`
		EventID          *string           `json:"eventID,omitempty"`
		EventChecksum    *string           `json:"eventChecksum,omitempty"`
		ContextData      map[string]string `json:"contextData,omitempty"`
	}{
		Payload:          o.Payload,
		RetryCount:       o.RetryCount,
		PipelineId:       o.PipelineId,
		PipelinePosition: o.PipelinePosition,
		ContextData:      o.ContextData,
	}

	// Empty strings are null
	if o.ID != "" {
		test.ID = &o.ID
	}
	if o.AppServiceKey != "" {
		test.AppServiceKey = &o.AppServiceKey
	}
	if o.Version != "" {
		test.Version = &o.Version
	}
	if o.CorrelationID != "" {
		test.CorrelationID = &o.CorrelationID
	}

	return json.Marshal(test)
}

// UnmarshalJSON returns an object from JSON.
func (o *StoredObject) UnmarshalJSON(data []byte) error {
	alias := new(struct {
		ID               *string           `json:"id"`
		AppServiceKey    *string           `json:"appServiceKey"`
		Payload          []byte            `json:"payload"`
		RetryCount       int               `json:"retryCount"`
		PipelineId       string            `json:"pipelineId"`
		PipelinePosition int               `json:"pipelinePosition"`
		Version          *string           `json:"version"`
		CorrelationID    *string           `json:"correlationID"`
		EventID          *string           `json:"eventID"`
		EventChecksum    *string           `json:"eventChecksum"`
		ContextData      map[string]string `json:"contextData,omitempty"`
	})

	// Error with unmarshaling
	if err := json.Unmarshal(data, alias); err != nil {
		return err
	}

	// Check nil fields
	if alias.ID != nil {
		o.ID = *alias.ID
	}
	if alias.AppServiceKey != nil {
		o.AppServiceKey = *alias.AppServiceKey
	}
	if alias.Version != nil {
		o.Version = *alias.Version
	}
	if alias.CorrelationID != nil {
		o.CorrelationID = *alias.CorrelationID
	}

	o.Payload = alias.Payload
	o.RetryCount = alias.RetryCount
	o.PipelineId = alias.PipelineId
	o.PipelinePosition = alias.PipelinePosition
	o.ContextData = alias.ContextData

	return nil
}
