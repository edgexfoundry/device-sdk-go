//
// Copyright (c) 2017 Cavium
// Copyright (c) 2021 Intel Corporation
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

package transforms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1" //nolint: gosec
	"encoding/base64"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"
)

const (
	plainString = "This is the test string used for testing"
	iv          = "123456789012345678901234567890"
	key         = "aquqweoruqwpeoruqwpoeruqwpoierupqoweiurpoqwiuerpqowieurqpowieurpoqiweuroipwqure"
)

type encryptionDetails struct {
	Algo       string
	Key        string
	InitVector string
}

var aesData = encryptionDetails{
	Algo:       "AES",
	Key:        key,
	InitVector: iv,
}

func aesDecrypt(crypt []byte, aesData encryptionDetails) []byte {
	hash := sha1.New() //nolint: gosec

	hash.Write([]byte((aesData.Key)))
	key := hash.Sum(nil)
	key = key[:blockSize]

	iv := make([]byte, blockSize)
	copy(iv, aesData.InitVector)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("key error")
	}

	decodedData, _ := base64.StdEncoding.DecodeString(string(crypt))

	ecb := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(decodedData))
	ecb.CryptBlocks(decrypted, decodedData)

	trimmed := pkcs5Trimming(decrypted)

	return trimmed
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func TestNewEncryption(t *testing.T) {
	enc := NewEncryption(aesData.Key, aesData.InitVector)

	continuePipeline, encrypted := enc.EncryptWithAES(ctx, []byte(plainString))
	assert.True(t, continuePipeline)

	decrypted := aesDecrypt(encrypted.([]byte), aesData)

	assert.Equal(t, plainString, string(decrypted))
	assert.Equal(t, ctx.ResponseContentType(), common.ContentTypeText)
}

func TestNewEncryptionWithSecrets(t *testing.T) {
	secretPath := "AES"
	secretName := "aesKey"

	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecret", secretPath, secretName).Return(map[string]string{secretName: key}, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	enc := NewEncryptionWithSecrets(secretPath, secretName, aesData.InitVector)

	continuePipeline, encrypted := enc.EncryptWithAES(ctx, []byte(plainString))
	assert.True(t, continuePipeline)

	decrypted := aesDecrypt(encrypted.([]byte), aesData)

	assert.Equal(t, plainString, string(decrypted))
	assert.Equal(t, ctx.ResponseContentType(), common.ContentTypeText)
}

func TestAESNoData(t *testing.T) {
	aesData := encryptionDetails{
		Algo:       "AES",
		Key:        key,
		InitVector: iv,
	}

	enc := NewEncryption(aesData.Key, aesData.InitVector)

	continuePipeline, result := enc.EncryptWithAES(ctx, nil)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error), "expect an error")
}
