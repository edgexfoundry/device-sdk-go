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

package functions

import (
	"encoding/xml"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

type PhoneInfo struct {
	CountryCode int `json:"country_code"`
	AreaCode    int `json:"area_code"`
	LocalPrefix int `json:"local_prefix"`
	LocalNumber int `json:"local_number"`
}

type Person struct {
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        PhoneInfo `json:"phone"`
	PhoneDisplay string    `json:"phone_display"`
}

func FormatPhoneDisplay(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Debug("Format Phone Number")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	person, ok := params[0].(Person)
	if !ok {
		edgexcontext.LoggingClient.Error("type received is not a Person")
	}

	person.PhoneDisplay = fmt.Sprintf("+%02d(%03d) %03d-%04d",
		person.Phone.CountryCode, person.Phone.AreaCode, person.Phone.LocalPrefix, person.Phone.LocalNumber)

	return true, person
}

func ConvertToXML(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	edgexcontext.LoggingClient.Debug("Convert to XML")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	person, ok := params[0].(Person)
	if !ok {
		edgexcontext.LoggingClient.Error("type received is not a Person")
	}

	result, err := xml.MarshalIndent(person, "", "   ")
	if err != nil {
		return false, fmt.Sprintf("Error parsing XML. Error: %s", err.Error())
	}

	return true, string(result)
}
