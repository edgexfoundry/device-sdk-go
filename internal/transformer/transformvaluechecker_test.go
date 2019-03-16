// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"
	"testing"
)

func TestCheckTransformedValueInRange_uint8(t *testing.T) {
	origin := uint8(10)
	transformed := float64(math.MaxUint8)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		fmt.Println(OverflowError{origin, transformed})
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint8_false(t *testing.T) {
	origin := uint8(10)
	transformed := float64(math.MaxUint8 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint16(t *testing.T) {
	origin := uint16(10)
	transformed := float64(math.MaxUint16)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint16_false(t *testing.T) {
	origin := uint16(10)
	transformed := float64(math.MaxUint16 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint32(t *testing.T) {
	origin := uint32(10)
	transformed := float64(math.MaxUint32)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint32_false(t *testing.T) {
	origin := uint32(10)
	transformed := float64(math.MaxUint32 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint64(t *testing.T) {
	origin := uint64(10)
	transformed := float64(math.MaxUint64)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_uint64_false(t *testing.T) {
	origin := uint64(10)
	transformed := float64(math.MaxUint64 * 2)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int8(t *testing.T) {
	origin := int8(10)
	transformed := float64(math.MaxInt8)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int8_false(t *testing.T) {
	origin := int8(10)
	transformed := float64(math.MaxInt8 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int16(t *testing.T) {
	origin := int16(10)
	transformed := float64(math.MaxInt16)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int16_false(t *testing.T) {
	origin := int16(10)
	transformed := float64(math.MaxInt16 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int32(t *testing.T) {
	origin := int32(10)
	transformed := float64(math.MaxInt32)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int32_false(t *testing.T) {
	origin := int32(10)
	transformed := float64(math.MaxInt32 + 1)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int64(t *testing.T) {
	origin := int64(10)
	transformed := float64(math.MaxInt64)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_int64_false(t *testing.T) {
	origin := int64(10)
	transformed := float64(math.MaxUint64)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_float32(t *testing.T) {
	origin := float32(10)
	transformed := float64(math.MaxFloat32)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_float32_false(t *testing.T) {
	origin := float32(10)
	transformed := float64(math.MaxFloat32 * 2)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != false {
		t.Fatalf("Transformed value '%v' should out of the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_float64(t *testing.T) {
	origin := float64(10)
	transformed := float64(math.MaxFloat64)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange != true {
		t.Fatalf("Transformed value '%v' is not within the '%T' value type range", transformed, origin)
	}
}

func TestCheckTransformedValueInRange_unsupportedDataType(t *testing.T) {
	origin := "123"
	transformed := float64(123)

	inRange := checkTransformedValueInRange(origin, transformed)

	if inRange == true {
		t.Fatalf("Unexpected test result. Data type %T should not support range checking", origin)
	}
}
