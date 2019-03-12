// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import "fmt"

// OverflowError is used to throw the error of transformed value is out of range
type OverflowError struct {
	origin      interface{}
	transformed float64
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("overflow failed, transformed value '%v' is not within the '%T' value type range", e.transformed, e.origin)
}

func (e OverflowError) String() string {
	return fmt.Sprintf("overflow failed, transformed value '%v' is not within the '%T' value type range", e.transformed, e.origin)
}

func NewOverflowError(origin interface{}, transformed float64) OverflowError {
	return OverflowError{origin: origin, transformed: transformed}
}
