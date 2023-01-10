package main

import (
	"fmt"
	"net/http"

	"github.com/IkehAkinyemi/logaudit/internal/utils"
)

// The logError method is a generic helper for logging an error message and
// additional information from the request including the HTTP method and URL.
func (svc *service) logError(r *http.Request, err error) {
	svc.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// errorResponse method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code
func (svc *service) errorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message interface{}) {
	env := utils.Envelope{"error": message}
	err := utils.WriteJSON(w, statusCode, env, nil)
	if err != nil {
		svc.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse() method reports runtime errors/problems.
func (svc *service) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	svc.logError(r, err)
	msg := "the server encountered an error and could not process your request"
	svc.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (svc *service) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource could not be found"
	svc.errorResponse(w, r, http.StatusNotFound, msg)
}

func (svc *service) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	svc.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (svc *service) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	svc.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// invalidAuthenticationTokenResponse reports service authentication errors in regards to token
func (svc *service) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// Keeps a reminder for the client about the bearer token
	w.Header().Add("WWW-Authentication", "Bearer")
	msg := "invalid or missing authentication token"
	svc.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// authenticationRequiredResponse reports error relating to token-based authentation.
func (svc *service) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "service must be authenticated to access this resource"
	svc.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// invalidCredentialResponse reports service authentication errors.
func (svc *service) invalidCredentialResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid authentication credentials"
	svc.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// failedValidationResponse reports errors from JSON validation
func (svc *service) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	svc.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
