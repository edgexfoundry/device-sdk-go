//
// Copyright (c) 2020 Intel Corporation
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

package requests

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// SecretsKeyValue is a secret key/value pair to be stored in the Secret Store
// See detail specified by the V2 API swagger in openapi/v2
type SecretsKeyValue struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// SecretsRequest is the request DTO for storing supplied secrets at specified Path in the Secret Store
// See detail specified by the V2 API swagger in openapi/v2
type SecretsRequest struct {
	common.BaseRequest `json:",inline"`
	Path               string            `json:"path" validate:"required"`
	Secrets            []SecretsKeyValue `json:"secrets" validate:"required,gt=0,dive"`
}

// Validate satisfies the Validator interface
func (sr SecretsRequest) Validate() error {
	err := v2.Validate(sr)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the SecretsRequest type
func (sr *SecretsRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		common.BaseRequest
		Path    string
		Secrets []SecretsKeyValue
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal SecretsRequest body as JSON.", err)
	}

	*sr = SecretsRequest(alias)

	// validate SecretsRequest DTO
	if err := sr.Validate(); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "SecretsRequest validation failed.", err)
	}
	return nil
}
