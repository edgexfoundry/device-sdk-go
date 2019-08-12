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
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	envValue  = "envValue"
	rootKey   = "rootKey"
	rootValue = "rootValue"
	sub       = "sub"
	subKey    = "subKey"
	subValue  = "subValue"

	useRegistryValue = "useRegistry"
	urlValue         = "consul://localhost:8500"

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

func TestOverrideUseRegistryFromEnvironment(t *testing.T) {
	var tests = []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{"valid", map[string]string{envKeyUrl: urlValue}, urlValue},
		{"no variable", map[string]string{}, useRegistryValue},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sut := newSUT(t, test.env)

			result := sut.OverrideUseRegistryFromEnvironment(useRegistryValue)

			assert.Equal(t, test.expected, result)
		})
	}
}
