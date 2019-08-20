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

package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/pelletier/go-toml"
)

const (
	envKeyRegistryUrl = "edgex_registry"
	envKeyServiceUrl  = "edgex_service"
)

// environment is receiver that holds environment variables and encapsulates toml.Tree-based configuration field
// overrides.  Assumes "_" embedded in environment variable key separates substructs; e.g. foo_bar_baz might refer to
//
// 		type foo struct {
// 			bar struct {
//          	baz string
//  		}
//		}
type environment struct {
	env map[string]interface{}
}

// NewEnvironment constructor reads/stores os.Environ() for use by environment receiver methods.
func NewEnvironment() *environment {
	osEnv := os.Environ()
	e := &environment{
		env: make(map[string]interface{}, len(osEnv)),
	}
	for _, env := range osEnv {
		kv := strings.Split(env, "=")
		if len(kv) == 2 && len(kv[0]) > 0 && len(kv[1]) > 0 {
			e.env[kv[0]] = kv[1]
		}
	}
	return e
}

// OverrideRegistryInfoFromEnvironment method overrides registry location with environment variables.
func (e *environment) OverrideRegistryInfoFromEnvironment(registry common.RegistryInfo) common.RegistryInfo {
	if env := os.Getenv(envKeyRegistryUrl); env != "" {
		if u, err := url.Parse(env); err == nil {
			if p, err := strconv.ParseInt(u.Port(), 10, 0); err == nil {
				registry.Port = int(p)
				registry.Host = u.Hostname()
				registry.Type = u.Scheme
			}
		}
	}
	return registry
}

// OverrideServiceInfoFromEnvironment method overrides Service location with environment variables.
func (e *environment) OverrideServiceInfoFromEnvironment(service common.ServiceInfo) common.ServiceInfo {
	if env := os.Getenv(envKeyServiceUrl); env != "" {
		if u, err := url.Parse(env); err == nil {
			if p, err := strconv.ParseInt(u.Port(), 10, 0); err == nil {
				service.Port = int(p)
				service.Host = u.Hostname()
				service.Protocol = u.Scheme
			}
		}
	}
	return service
}

// OverrideRegistryConfigFromEnvironment method replaces values in the toml.Tree for matching environment variable keys.
func (e *environment) OverrideFromEnvironment(tree *toml.Tree) *toml.Tree {
	for k, v := range e.env {
		k = strings.Replace(k, "_", ".", -1)
		switch {
		case tree.Has(k):
			// global key
			tree.Set(k, v)
		}
	}
	return tree
}
