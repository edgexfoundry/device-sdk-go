// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package correlation

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/google/uuid"
)

func ManageHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get(common.CorrelationHeader)
		if hdr == "" {
			hdr = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), common.CorrelationHeader, hdr) // nolint:staticcheck
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(lc logger.LoggingClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if lc.LogLevel() == models.TraceLog {
				begin := time.Now()
				correlationId := IdFromContext(r.Context())
				lc.Trace("Begin request", common.CorrelationHeader, correlationId, "path", r.URL.Path)
				next.ServeHTTP(w, r)
				lc.Trace("Response complete", common.CorrelationHeader, correlationId, "duration", time.Since(begin).String())
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// UrlDecodeMiddleware decode the path variables
// After invoking the router.UseEncodedPath() func, the path variables needs to decode before passing to the controller
func UrlDecodeMiddleware(lc logger.LoggingClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			for k, v := range vars {
				unescape, err := url.PathUnescape(v)
				if err != nil {
					lc.Debugf("failed to decode the %s from the value %s", k, v)
					return
				}
				vars[k] = unescape
			}
			next.ServeHTTP(w, r)
		})
	}
}
