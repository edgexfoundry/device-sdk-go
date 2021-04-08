//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getUniqueOrigin(t *testing.T) {
	for i := 0; i < rand.Intn(1000); i++ {
		t.Run(fmt.Sprintf("TestCase%d", i), func(t *testing.T) {
			t.Parallel()
			o1 := getUniqueOrigin()
			o2 := getUniqueOrigin()
			assert.NotEqual(t, o1, o2)
		})
	}
}
