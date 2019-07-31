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

package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/pelletier/go-toml"
)

const (
	configDirectory = "./res"
	configDirEnv    = "EDGEX_CONF_DIR"
)

// LoadFromFile loads .toml file for configuration
func LoadFromFile(profile string, configDir string) (configuration *ConfigurationStruct, err error) {
	path := determinePath(configDir)
	fileName := path + "/" + internal.ConfigFileName //default profile
	if len(profile) > 0 {
		fileName = path + "/" + profile + "/" + internal.ConfigFileName
	}
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Could not load configuration file (%s): %v", fileName, err.Error())
	}

	// Decode the configuration from TOML
	configuration = &ConfigurationStruct{}
	err = toml.Unmarshal(contents, configuration)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse configuration file (%s): %v", fileName, err.Error())
	}

	return configuration, nil
}

func determinePath(configDir string) string {
	path := configDir

	if len(path) == 0 { //No cmd line param passed
		//Assumption: one service per container means only one var is needed, set accordingly for each deployment.
		//For local dev, do not set this variable since configs are all named the same.
		path = os.Getenv(configDirEnv)
	}

	if len(path) == 0 { //Var is not set
		path = configDirectory
	}

	return path
}
