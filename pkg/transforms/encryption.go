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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

type Encryption struct {
	SecretPath           string
	SecretName           string
	EncryptionKey        string
	InitializationVector string
}

// NewEncryption creates, initializes and returns a new instance of Encryption
func NewEncryption(encryptionKey string, initializationVector string) Encryption {
	return Encryption{
		EncryptionKey:        encryptionKey,
		InitializationVector: initializationVector,
	}
}

// NewEncryptionWithSecrets creates, initializes and returns a new instance of Encryption configured
// to retrieve the encryption key from the Secret Store
func NewEncryptionWithSecrets(secretPath string, secretName string, initializationVector string) Encryption {
	return Encryption{
		SecretPath:           secretPath,
		SecretName:           secretName,
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
func (aesData Encryption) EncryptWithAES(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no data received to encrypt")
	}

	ctx.LoggingClient().Debug("Encrypting with AES")

	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	iv := make([]byte, blockSize)
	copy(iv, aesData.InitializationVector)

	hash := sha1.New()

	// If using Secret Store for the encryption key
	if len(aesData.SecretPath) != 0 && len(aesData.SecretName) != 0 {
		// Note secrets are cached so this call doesn't result in unneeded calls to SecretStore Service and
		// the cache is invalidated when StoreSecrets is used.
		secretData, err := ctx.GetSecret(aesData.SecretPath, aesData.SecretName)
		if err != nil {
			return false, fmt.Errorf(
				"unable to retieve encryption key at secret path=%s and name=%s",
				aesData.SecretPath,
				aesData.SecretName)
		}

		key, ok := secretData[aesData.SecretName]
		if !ok {
			return false, fmt.Errorf("unable find encryption key in secret data for name=%s", aesData.SecretName)
		}

		ctx.LoggingClient().Debugf(
			"Using encryption key from Secret Store at path=%s & name=%s",
			aesData.SecretPath,
			aesData.SecretName)

		aesData.EncryptionKey = key
	}

	if len(aesData.EncryptionKey) == 0 {
		return false, fmt.Errorf("AES encryption key not set")
	}

	hash.Write([]byte((aesData.EncryptionKey)))
	key := hash.Sum(nil)
	key = key[:blockSize]

	block, err := aes.NewCipher(key)
	if err != nil {
		return false, err
	}

	ecb := cipher.NewCBCEncrypter(block, iv)
	content := pkcs5Padding(byteData, block.BlockSize())
	encrypted := make([]byte, len(content))
	ecb.CryptBlocks(encrypted, content)

	encodedData := []byte(base64.StdEncoding.EncodeToString(encrypted))

	// Set response "content-type" header to "text/plain"
	ctx.SetResponseContentType(clients.ContentTypeText)

	return true, encodedData
}
