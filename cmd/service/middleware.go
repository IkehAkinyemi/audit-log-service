package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
)

// authenticate does stateful authentication for each service
func (svc *service) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
