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
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

const (
	envValue  = "envValue"
	rootKey   = "rootKey"
	rootValue = "rootValue"
	sub       = "sub"
	subKey    = "subKey"
	subValue  = "subValue"

	testToml = `
` + rootKey + `="` + rootValue + `"
[` + sub + `]
` + subKey + `="` + subValue + `"`
)

func newSUT(t *testing.T, env map[string]string) *environment {
	os.Clearenv()
	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			t.Fail()
		}
	}
	return NewEnvironment()
}

func newOverrideFromEnvironmentSUT(t *testing.T, envKey string, envValue string) (*toml.Tree, *environment) {
	tree, err := toml.Load(testToml)
	if err != nil {
		t.Fail()
	}
	return tree, newSUT(t, map[string]string{envKey: envValue})
}

func TestKeyMatchOverwritesValue(t *testing.T) {
	var tests = []struct {
		name          string
		key           string
		envKey        string
		envValue      string
		expectedValue string
	}{
		{"generic root", rootKey, rootKey, envValue, envValue},
		{"generic sub", sub + "." + subKey, sub + "." + subKey, envValue, envValue},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree, sut := newOverrideFromEnvironmentSUT(t, test.key, test.envValue)

			result := sut.OverrideFromEnvironment(tree)

			assert.Equal(t, result.Get(test.key), test.envValue)
		})
	}
}

func TestNonMatchingKeyDoesNotOverwritesValue(t *testing.T) {
	var tests = []struct {
		name          string
		key           string
		envKey        string
		envValue      string
		expectedValue string
	}{
		{"root", rootKey, rootKey, envValue, rootValue},
		{"sub", sub + "." + subKey, sub + "." + subKey, envValue, rootValue},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree, sut := newOverrideFromEnvironmentSUT(t, test.key, test.envValue)

			result := sut.OverrideFromEnvironment(tree)

			assert.Equal(t, result.Get(test.key), test.envValue)
		})
	}
}

const (
	expectedTypeValue = "consul"
	expectedHostValue = "localhost"
	expectedPortValue = 8500

	defaultHostValue = "defaultHost"
	defaultPortValue = 987654321
	defaultTypeValue = "defaultType"
)

func initializeTest(t *testing.T) common.RegistryInfo {
	os.Clearenv()
	return common.RegistryInfo{
		Host: defaultHostValue,
		Port: defaultPortValue,
		Type: defaultTypeValue,
	}
}

func TestEnvVariableUpdatesRegistryInfo(t *testing.T) {
	registryInfo := initializeTest(t)
	sut := newSUT(t, map[string]string{envKeyUrl: expectedTypeValue + "://" + expectedHostValue + ":" + strconv.Itoa(expectedPortValue)})

	registryInfo = sut.OverrideRegistryInfoFromEnvironment(registryInfo)

	assert.Equal(t, registryInfo.Host, expectedHostValue)
	assert.Equal(t, registryInfo.Port, expectedPortValue)
	assert.Equal(t, registryInfo.Type, expectedTypeValue)
}

func TestNoEnvVariableDoesNotUpdateRegistryInfo(t *testing.T) {
	registryInfo := initializeTest(t)
	sut := newSUT(t, map[string]string{})

	registryInfo = sut.OverrideRegistryInfoFromEnvironment(registryInfo)

	assert.Equal(t, registryInfo.Host, defaultHostValue)
	assert.Equal(t, registryInfo.Port, defaultPortValue)
	assert.Equal(t, registryInfo.Type, defaultTypeValue)
}
