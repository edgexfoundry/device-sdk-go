//
// Copyright (c) 2019 Intel Corporation
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

package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Configuration struct which contains
// an array of bindings
type Configuration struct {
	Runtime  string    `json:"runtime"`
	Bindings []Binding `json:"bindings"`
}

// Binding contains Metadata associated with each binding
type Binding struct {
	Type      string `json:"type"`
	Direction string `json:"direction"`
	Name      string `json:"name"`
	Topic     string `json:"topic"`
}

// LoadConfiguration loads the configuration from file
func LoadConfiguration() Configuration {
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened config.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result Configuration
	json.Unmarshal([]byte(byteValue), &result)

	return result
}
