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

package container

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// StoreClientName contains the name of interfaces.StoreClient implementation in the DIC.
var StoreClientName = di.TypeInstanceToName((*interfaces.StoreClient)(nil))

// StoreClientFrom helper function queries the DIC and returns interfaces.StoreClient implementation.
func StoreClientFrom(get di.Get) interfaces.StoreClient {
	item := get(StoreClientName)

	if item == nil {
		return nil
	}

	return item.(interfaces.StoreClient)
}
