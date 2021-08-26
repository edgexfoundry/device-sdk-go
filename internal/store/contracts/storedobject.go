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

// contracts are implementation agnostic data storage models.
package contracts

import (
	"errors"

	"github.com/google/uuid"
)

// StoredObject is the atomic and most abstract description of what is collected by the export store system.
type StoredObject struct {
	// ID uniquely identifies this StoredObject
	ID string
	// AppServiceKey identifies the app to which this data belongs.
	AppServiceKey string
	// Payload is the data to be exported
	Payload []byte
	// RetryCount is how many times this has tried to be exported
	RetryCount int
	// PipelineId is the ID of the pipeline that needs to be restarted.
	PipelineId string
	// PipelinePosition is where to pickup in the pipeline
	PipelinePosition int
	// Version is a hash of the functions to know if the pipeline has changed.
	Version string
	// CorrelationID is an identifier provided by EdgeX to track this record as it moves
	CorrelationID string
	// ContextData is a snapshot of data used by the pipeline at runtime
	ContextData map[string]string
}

// NewStoredObject creates a new instance of StoredObject and is the preferred way to create one.
func NewStoredObject(appServiceKey string, payload []byte, pipelineId string, pipelinePosition int,
	version string, contextData map[string]string) StoredObject {
	return StoredObject{
		AppServiceKey:    appServiceKey,
		Payload:          payload,
		RetryCount:       0,
		PipelineId:       pipelineId,
		PipelinePosition: pipelinePosition,
		Version:          version,
		ContextData:      contextData,
	}
}

// ValidateContract ensures that the required fields are present on the object.
func (o *StoredObject) ValidateContract(IDRequired bool) error {
	if IDRequired {
		if o.ID == "" {
			return errors.New("invalid contract, ID cannot be empty")
		}
	} else {
		if o.ID == "" {
			o.ID = uuid.New().String()
		}
	}

	parsed, err := uuid.Parse(o.ID)
	if err != nil {
		return errors.New("invalid contract, ID must be UUID")
	}

	o.ID = parsed.String()

	if o.AppServiceKey == "" {
		return errors.New("invalid contract, app service key cannot be empty")
	}
	if len(o.Payload) == 0 {
		return errors.New("invalid contract, payload cannot be empty")
	}
	if o.Version == "" {
		return errors.New("invalid contract, version cannot be empty")
	}

	return nil
}
