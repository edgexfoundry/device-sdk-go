//
// Copyright (c) 2017 Cavium
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

package transforms

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

type Encryption struct {
	Key                  string
	InitializationVector string
}

// NewEncryption creates, initializes and returns a new instance of Encryption
func NewEncryption(key string, initializationVector string) Encryption {
	return Encryption{
		Key:                  key,
		InitializationVector: initializationVector,
	}
}

// IV and KEY must be 16 bytes
const blockSize = 16

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// EncryptWithAES encrypts a string, []byte, or json.Marshaller type using AES encryption.
// It will return a Base64 encode []byte of the encrypted data.
func (aesData Encryption) EncryptWithAES(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		return false, errors.New("no data received to encrypt")
	}
	edgexcontext.LoggingClient.Debug("Encrypting with AES")
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}

	iv := make([]byte, blockSize)
	copy(iv, []byte(aesData.InitializationVector))

	hash := sha1.New()

	hash.Write([]byte((aesData.Key)))
	key := hash.Sum(nil)
	key = key[:blockSize]

	block, err := aes.NewCipher(key)
	if err != nil {
		return false, err
	}

	ecb := cipher.NewCBCEncrypter(block, iv)
	content := pkcs5Padding(data, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	encodedData := []byte(base64.StdEncoding.EncodeToString(crypted))

	// Set response "content-type" header to "text/plain"
	edgexcontext.ResponseContentType = clients.ContentTypeText

	return true, encodedData
}
