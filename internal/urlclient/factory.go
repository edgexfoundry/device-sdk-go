/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

// urlclient provides functions to integrate the client code in go-mod-core-contracts with application specific code
package urlclient

import (
	"context"
	"sync"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/retry"
	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/endpoint"
)

// NewData returns a URLClient that connects to a service attached to data.
func NewData(
	ctx context.Context,
	registryClient registry.Client,
	wg *sync.WaitGroup,
	path string) interfaces.URLClient {

	return newClient(
		ctx,
		registryClient,
		wg,
		clients.CoreDataServiceKey,
		path,
		common.CurrentConfig.Clients[common.ClientData],
	)
}

// NewMetadata returns a URLClient that connects to a service attached to metadata.
func NewMetadata(
	ctx context.Context,
	registryClient registry.Client,
	wg *sync.WaitGroup,
	path string) interfaces.URLClient {

	return newClient(
		ctx,
		registryClient,
		wg,
		clients.CoreMetaDataServiceKey,
		path,
		common.CurrentConfig.Clients[common.ClientMetadata],
	)
}

// newClient is a factory function that uses pre-defined constants to reduce code duplication.
func newClient(
	ctx context.Context,
	registryClient registry.Client,
	wg *sync.WaitGroup,
	serviceKey string,
	path string,
	client bootstrapConfig.ClientInfo) interfaces.URLClient {

	interval := 15

	if registryClient != nil {
		return retry.New(
			endpoint.New(
				ctx,
				wg,
				registryClient,
				serviceKey,
				path,
				interval,
			).Monitor(),
			interval,    // retry interval == interval because we don't need to check for an update before an update
			interval*10, // this scalar multiplier was chosen because it seemed reasonable
		)
	}

	return local.New(client.Url() + path)
}
