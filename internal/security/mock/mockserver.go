//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0
//

package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
)

// GetMockTokenServer is a test helper for unit tests of vault token server
func GetMockTokenServer(tokenDataMap *sync.Map) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		urlPath := req.URL.String()
		switch req.Method {
		case http.MethodGet:
			if urlPath == "/v1/auth/token/lookup-self" {
				token := req.Header.Get(vault.AuthTypeHeader)
				sampleTokenLookup, exists := tokenDataMap.Load(token)
				if !exists {
					rw.WriteHeader(http.StatusForbidden)
					_, _ = rw.Write([]byte("permission denied"))
				} else {
					resp := &vault.TokenLookupResponse{
						Data: sampleTokenLookup.(vault.TokenLookupMetadata),
					}
					if ret, err := json.Marshal(resp); err != nil {
						rw.WriteHeader(http.StatusInternalServerError)
						_, _ = rw.Write([]byte(err.Error()))
					} else {
						rw.WriteHeader(http.StatusOK)
						_, _ = rw.Write(ret)
					}
				}
			} else {
				returnUnknownPath(rw, urlPath)
			}
		case http.MethodPost:
			if urlPath == "/v1/auth/token/renew-self" {
				token := req.Header.Get(vault.AuthTypeHeader)
				sampleTokenLookup, exists := tokenDataMap.Load(token)
				if !exists {
					rw.WriteHeader(http.StatusForbidden)
					_, _ = rw.Write([]byte("permission denied"))
				} else {
					currentTTL := sampleTokenLookup.(vault.TokenLookupMetadata).Ttl
					if currentTTL <= 0 {
						// already expired
						rw.WriteHeader(http.StatusForbidden)
						_, _ = rw.Write([]byte("permission denied"))
					} else {
						tokenPeriod := sampleTokenLookup.(vault.TokenLookupMetadata).Period

						tokenDataMap.Store(token, vault.TokenLookupMetadata{
							Renewable: true,
							Ttl:       tokenPeriod,
							Period:    tokenPeriod,
						})
						rw.WriteHeader(http.StatusOK)
						_, _ = rw.Write([]byte("token renewed"))
					}
				}
			} else {
				returnUnknownPath(rw, urlPath)
			}
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = rw.Write([]byte("method not allowed!"))
		}
	}))

	return server
}

func returnUnknownPath(rw http.ResponseWriter, urlPath string) {
	rw.WriteHeader(http.StatusNotFound)
	_, _ = rw.Write([]byte(fmt.Sprintf("Unknown urlPath: %s", urlPath)))
}
