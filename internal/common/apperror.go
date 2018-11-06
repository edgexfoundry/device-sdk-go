// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import "net/http"

type AppError interface {
	Error() error
	Message() string
	Code() int
}

type appError struct {
	err  error
	msg  string
	code int
}

func (a appError) Error() error {
	return a.err
}

func (a appError) Message() string {
	return a.msg
}

func (a appError) Code() int {
	return a.code
}

func NewNotFoundError(msg string, err error) AppError {
	return appError{err: err, msg: msg, code: http.StatusNotFound}
}

func NewServerError(msg string, err error) AppError {
	return appError{err: err, msg: msg, code: http.StatusInternalServerError}
}

func NewBadRequestError(msg string, err error) AppError {
	return appError{err: err, msg: msg, code: http.StatusBadRequest}
}

func NewLockedError(msg string, err error) AppError {
	return appError{err: err, msg: msg, code: http.StatusLocked}
}
