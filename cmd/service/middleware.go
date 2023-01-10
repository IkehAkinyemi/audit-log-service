package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	"github.com/felixge/httpsnoop"
)

// authenticate does stateful authentication for each service
func (svc *service) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This indicates to any caches that the response may
		// vary based on the value of Authorization.
		w.Header().Set("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// using the model.AnonymousService if no Authorization header found
			r = svc.contextSetService(r, &model.AnonymousService)
			next.ServeHTTP(w, r)
			return
		}

		keyParse := strings.Split(authHeader, " ")
		if len(keyParse) != 2 || keyParse[0] != "Key" {
			svc.invalidAuthenticationTokenResponse(w, r)
			return
		}
		APIKey := keyParse[1]

		v := utils.NewValidator()
		if utils.ValidateTokenPlaintext(v, APIKey); !v.Valid() {
			svc.invalidAuthenticationTokenResponse(w, r)
			return
		}

		serviceID, err := svc.db.GetTokenByAPIKey(r.Context(), APIKey)
		if err != nil {
			switch {
			case errors.Is(err, model.ErrRecordNotFound):
				svc.invalidAuthenticationTokenResponse(w, r)
			default:
				svc.serverErrorResponse(w, r, err)
			}
			return
		}

		r = svc.contextSetService(r, serviceID)

		next.ServeHTTP(w, r)
	})
}

// requiredAuthenticatedService controls access to restricted endpoints – Authorization
func (svc *service) requiredAuthenticatedService(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceID := svc.contextGetService(r)

		if serviceID.IsAnonymous() {
			svc.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoverPanic graciouly recovers any panic within the goroutine handling the request
func (svc *service) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				svc.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// metrics measures specific request-response metrics for monitoring.
func (svc *service) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
