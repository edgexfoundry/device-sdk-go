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

// This test will only be executed if the tag brokerRunning is added when running
// the tests with a command like:
// go test -tags brokerRunning
package secure

import (
	"errors"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCACert = `-----BEGIN CERTIFICATE-----
MIIDhTCCAm2gAwIBAgIUQl1RUGewZOXaSLnmH1i12zSYOtswDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDELMAkGA1UEAwwCY2EwHhcNMjAwNDA4
MDExNDQ2WhcNMjUwNDA4MDExNDQ2WjBSMQswCQYDVQQGEwJVUzETMBEGA1UECAwK
U29tZS1TdGF0ZTEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMQsw
CQYDVQQDDAJjYTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOqslFtX
nxr6yBZdLDKp1iTmsnFreEit7Z1BnNy9vQW6xrKRH+nxZWr0n9UIbx7KtmFkSBQ9
Bb5zC/3ZdjcuQAuKSTgQB7AP1D2dX6geJPo1Ph9NS0aVmuUqQ6dU+/4R5ATfoWag
M7slCixfkBzbHEh0mCqr7FoDWq2h+Cz2n8K85tBZjLyUuzyRaqH7ZkHfJD1cxkGK
FcwudCg4zpKYOSctm+JpTlF6YPjlngN79jaJIQEAmx/twv1lOCAGBw/hZM3FGmQx
5dA1W7qaJ6NHgNRXWRS1AERtHpAAsWNBT1CKuAS/j0PlreRyR3aMgQYQ5camxi9a
qCrMiHybaqj+UCkCAwEAAaNTMFEwHQYDVR0OBBYEFPNCbvrfw2QDoOyYfNjT9sNO
52xOMB8GA1UdIwQYMBaAFPNCbvrfw2QDoOyYfNjT9sNO52xOMA8GA1UdEwEB/wQF
MAMBAf8wDQYJKoZIhvcNAQELBQADggEBAHdFTqe6vi3BzgOMJEMO+81ZmiMohgKZ
Alyo8wH1C5RgwWW5w1OU+2RQfdOZgDfFkuQzmj0Kt2gzqACuAEtKzDt78lJ4f+WZ
MmRKBudJONUHTTm1micK3pqmn++nSygag0KxDvVbL+stSEgZwEBSOEvGDPXrL5qs
5yVOCi4xvsOCa1ymSnW6sX0z5GcgJQj2Znrr5QbEKHFSG86+WYEYnZ2zCNV7ahQo
bwXGZPOCUkpQzOstie/lPsf3Sd13/NIAk23TQ+rtaWIP9syQ85XWGRKRAUFOJEK0
2/jr0Xot+Y/3raEfNSrq6sHTzX1q4PoWkSwNEEGXifBqDr+9PXK3mOQ=
-----END CERTIFICATE-----
`
const testClientCert = `-----BEGIN CERTIFICATE-----
MIIDLzCCAhcCFG+y+oEr87O2iQH90ayO4hU/GvSqMA0GCSqGSIb3DQEBCwUAMFIx
CzAJBgNVBAYTAlVTMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQxCzAJBgNVBAMMAmNhMB4XDTIwMDQwODAxMTY1
OVoXDTIxMDQwMzAxMTY1OVowVjELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUt
U3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEPMA0GA1UE
AwwGY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4TlobJoF
gNoCc5Znb0OzVoMypoay1RSTAhnU0arpHVugUMZMO6oxSt371MN+e4cUxoes4uhN
qeVG7AxUkdMCNJbzjAmJeDQtLKYHcY4YI30HHWCW0c8SxEsrj6DzjizgKZcUdX4H
6HwAltOp/RZYJTBVVexE1WYOheTNJuw5QeNbTGpfpKM7RuHADnytLbrSiK09FZYx
23PIsLhx8b7+k1AtRFGhFqDRMF6Fqbo6xdU8hZ1eAvJP5t87U/PWeQ9ld2lxd3fQ
xiP4IBQs1QI2gTp5O41ifRCpO7scXRaFweyPAgMVOQ42eVZiJUR37AF/nVzXxB5N
iTH9Ij/c/shJvQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQDZ1tvo2JbA27qs+DzH
PQudMgCPqHylnqlbX94FtKrIh6kP4YwrMNoOCdcU/MHGG2b3ldoMgx9qrTnkk8g1
3/gX/r4MDiTw2LocmIPYSukfR0J4k0ijlZtbtr9EtNPvy5iSla8Xi+iSm70wj+Zi
Z0GE0gOi8JfYPlxCtw3uVpsdqaHEevI70D4H1yAG22YYXUZt0QK02zztgBA2c7nE
kX0EMnYch0e7urs9o1M6JWJGlWZQxgVnxekbFDPfRelR1m0zFnbfXG2rnfuRpVEL
6SGxFU8+v1VepAHLvhS2VULYbWBOHZsh1yCteUXdePMYIN7c71qaCyC89N3GBia5
uXOR
-----END CERTIFICATE-----
`
const testClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA4TlobJoFgNoCc5Znb0OzVoMypoay1RSTAhnU0arpHVugUMZM
O6oxSt371MN+e4cUxoes4uhNqeVG7AxUkdMCNJbzjAmJeDQtLKYHcY4YI30HHWCW
0c8SxEsrj6DzjizgKZcUdX4H6HwAltOp/RZYJTBVVexE1WYOheTNJuw5QeNbTGpf
pKM7RuHADnytLbrSiK09FZYx23PIsLhx8b7+k1AtRFGhFqDRMF6Fqbo6xdU8hZ1e
AvJP5t87U/PWeQ9ld2lxd3fQxiP4IBQs1QI2gTp5O41ifRCpO7scXRaFweyPAgMV
OQ42eVZiJUR37AF/nVzXxB5NiTH9Ij/c/shJvQIDAQABAoIBADPL4BgZ0+ouOSIc
FO2hxDzBL4TctYQLl0OEbU1K4RG/YL8y25VdLrjpFGF6FDyUdFK0IS6N/k50TDs9
GrXusTMnBBvQlazvUvRRuqSC6UpAFsLK0+SsmsRKBVqiyWCJMYRfGnVq5qaw3fHR
++YYnWzwELASBkKNlgl09TleWkysbnZbWIMQ5Qm0k+s/9vvjooA2aMXTeLtyhGfI
49OvyCrrX5v7ILdHl7RGAyPRT+ipyt1i0fAqHk4ouLdTRrAx4S5TvUpszrts1P8f
5ggLd1s6RVTz27uASu3U/gLH630m1PU46d02UI1tWen3TgRm/VqjO2aqkZaZispQ
HwTRZIECgYEA9rL7KoZflVQJ4ndg3V522BhAciN99taYWHr018kG5vNVGFBHSVOt
De0gb7z8FhK0Zs4MifU3b03qr7Ac1+p0zIAwATPT4TOLzc4SKBd33TZk/JCZCGSR
hqQPF0FZ+EKJqh7yif+ssFXp0xKrNybm58Z7jfF8vWMdz0QkJ1pZkn8CgYEA6bcp
YkH6IoHmCZ5hWE3/hYQcvfcM10z0cWTTKstxgSid9dj0HUqxMsFhBF1yzUtsDZQB
E933gZyj/LE5Z/EbqUSX0H/M0P7Uwtj9lS7W/vQdOQMfAciqggNKhyaBnBYsxw9l
5IelOxGF+taEvDkPsVt9cvZm/nbf+irU5JLCzcMCgYEA8o3/jUwY5oV+QoAFaSHb
z5PoqVBkJTHREA20dgVdF+3fmMw1is8Os0aWQcaaREmXvgyRH4NOQc1mFd8ePNx0
giz3BfejNySrLGqUR37rh0BYAktZa3sV6j+b5s2GXCVvnShYZ35OmAGgqLsORGen
V/M6v9DTSJIPWR4yPc8DipkCgYEAhmtW/PFPaRtm7+9Ms5ogtWz3jvaRRx82lCVW
Io3iGVQADc8bD+HOqo94Oid5CMQxQFn4iLGoUb6Cvqo7hyGwNBmEa2GlripyuiJN
LslC1F4YlJrL8Z21G5PDAJpP/zLtzAt6Igc2LBP3B/7rVspG0U36h+1Z7U73oQ2T
ZmdWbTsCgYALxjB0NvqBk+TNYMZFysqZnI3CxYQXwHfElQQQUqcQnunAOLJ8H+nb
JryGx90ylYY2Mh2U273435uwQcX1g5gu3rBF8McHKj5EYSVDgpeBMx8ej2ENvW7q
CR6KVnoNdMwJZM3ARpBYNlhFTzDyew2WYLitZsN/uV8t+XxJFDyJQA==
-----END RSA PRIVATE KEY-----
`

func TestValidateSecrets(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	tests := []struct {
		Name             string
		AuthMode         string
		secrets          mqttSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"Invalid AuthMode", "BadAuthMode", mqttSecrets{}, true, "Invalid AuthMode selected"},
		{"No Auth No error", AuthModeNone, mqttSecrets{}, false, ""},
		{"UsernamePassword No Error", AuthModeUsernamePassword, mqttSecrets{
			username: "user",
			password: "Password",
		}, false, ""},
		{"UsernamePassword Error no Username", AuthModeUsernamePassword, mqttSecrets{
			password: "Password",
		}, true, "AuthModeUsernamePassword selected however Username or Password was not found at secret path"},
		{"UsernamePassword Error no Password", AuthModeUsernamePassword, mqttSecrets{
			username: "user",
		}, true, "AuthModeUsernamePassword selected however Username or Password was not found at secret path"},
		{"ClientCert No Error", AuthModeCert, mqttSecrets{
			certPemBlock: []byte("----"),
			keyPemBlock:  []byte("----"),
		}, false, ""},
		{"ClientCert No Key", AuthModeCert, mqttSecrets{
			certPemBlock: []byte("----"),
		}, true, "AuthModeCert selected however the key or cert PEM block was not found at secret path"},
		{"ClientCert No Cert", AuthModeCert, mqttSecrets{
			keyPemBlock: []byte("----"),
		}, true, "AuthModeCert selected however the key or cert PEM block was not found at secret path"},
		{"CACert no error", AuthModeCA, mqttSecrets{
			caPemBlock: []byte(testCACert),
		}, false, ""},
		{"CACert invalid error", AuthModeCA, mqttSecrets{
			caPemBlock: []byte(`------`),
		}, true, "Error parsing CA Certificate"},
		{"CACert no ca error", AuthModeCA, mqttSecrets{}, true, "AuthModeCA selected however no PEM Block was found at secret path"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target.authMode = test.AuthMode
			result := target.validateSecrets(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Equal(t, test.ErrorMessage, result.(error).Error())
			} else {
				assert.Nil(t, result, "Should be nil")
			}
		})
	}
}

func TestGetSecrets(t *testing.T) {
	// setup mock secret client
	expectedMqttSecrets := map[string]string{
		"username": "TEST_USER",
		"password": "TEST_PASS",
	}
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecrets", "").Return(nil)
	mockSecretProvider.On("GetSecrets", "/notfound").Return(nil, errors.New("Not Found"))
	mockSecretProvider.On("GetSecrets", "/mqtt").Return(expectedMqttSecrets, nil)

	target := NewMqttFactory(logger.NewMockClient(), mockSecretProvider, "", "", false)
	tests := []struct {
		Name            string
		AuthMode        string
		SecretPath      string
		ExpectedSecrets *mqttSecrets
		ExpectingError  bool
	}{
		{"No Auth No error", AuthModeNone, "", nil, false},
		{"Auth No Secrets found", AuthModeCA, "/notfound", nil, true},
		{"Auth With Secrets", AuthModeUsernamePassword, "/mqtt", &mqttSecrets{
			username:     "TEST_USER",
			password:     "TEST_PASS",
			keyPemBlock:  []uint8{},
			certPemBlock: []uint8{},
			caPemBlock:   []uint8{},
		}, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target.authMode = test.AuthMode
			target.secretPath = test.SecretPath
			mqttSecrets, err := target.getSecrets()
			if test.ExpectingError {
				assert.Error(t, err, "Expecting error")
				return
			}
			require.Equal(t, test.ExpectedSecrets, mqttSecrets)
		})
	}
}

func TestConfigureMQTTClientForAuth(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	tests := []struct {
		Name             string
		AuthMode         string
		secrets          mqttSecrets
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"Username and Password should be set", AuthModeUsernamePassword, mqttSecrets{username: MQTTSecretUsername, password: MQTTSecretPassword}, false, ""},
		{"No AuthMode", AuthModeNone, mqttSecrets{}, false, ""},
		{"Invalid AuthMode", "", mqttSecrets{}, false, ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target.authMode = test.AuthMode
			result := target.configureMQTTClientForAuth(test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Equal(t, test.ErrorMessage, result.(error).Error())
			} else {
				assert.Nil(t, result, "Should be nil")
			}
		})
	}
}
func TestConfigureMQTTClientForAuthWithUsernamePassword(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeUsernamePassword
	err := target.configureMQTTClientForAuth(mqttSecrets{
		username: "Username",
		password: "Password",
	})
	require.NoError(t, err)
	assert.Equal(t, target.opts.Username, "Username")
	assert.Equal(t, target.opts.Password, "Password")
	assert.Nil(t, target.opts.TLSConfig.ClientCAs)
	assert.Nil(t, target.opts.TLSConfig.Certificates)

}
func TestConfigureMQTTClientForAuthWithUsernamePasswordAndCA(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeUsernamePassword
	err := target.configureMQTTClientForAuth(mqttSecrets{
		username:   "Username",
		password:   "Password",
		caPemBlock: []byte(testCACert),
	})
	require.NoError(t, err)
	assert.Equal(t, target.opts.Username, "Username")
	assert.Equal(t, target.opts.Password, "Password")
	assert.Nil(t, target.opts.TLSConfig.Certificates)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
}

func TestConfigureMQTTClientForAuthWithCACert(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeCA
	err := target.configureMQTTClientForAuth(mqttSecrets{
		username:   "Username",
		password:   "Password",
		caPemBlock: []byte(testCACert),
	})

	require.NoError(t, err)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.Nil(t, target.opts.TLSConfig.Certificates)
}
func TestConfigureMQTTClientForAuthWithClientCert(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeCert
	err := target.configureMQTTClientForAuth(mqttSecrets{
		username:     "Username",
		password:     "Password",
		certPemBlock: []byte(testClientCert),
		keyPemBlock:  []byte(testClientKey),
		caPemBlock:   []byte(testCACert),
	})
	require.NoError(t, err)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.NotNil(t, target.opts.TLSConfig.Certificates)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
}

func TestConfigureMQTTClientForAuthWithClientCertNoCA(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeCert
	err := target.configureMQTTClientForAuth(mqttSecrets{
		username:     MQTTSecretUsername,
		password:     MQTTSecretPassword,
		certPemBlock: []byte(testClientCert),
		keyPemBlock:  []byte(testClientKey),
	})

	require.NoError(t, err)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.NotNil(t, target.opts.TLSConfig.Certificates)
	assert.Nil(t, target.opts.TLSConfig.ClientCAs)
}
func TestConfigureMQTTClientForAuthWithNone(t *testing.T) {
	target := NewMqttFactory(logger.NewMockClient(), nil, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = AuthModeNone
	err := target.configureMQTTClientForAuth(mqttSecrets{})

	require.NoError(t, err)
}
