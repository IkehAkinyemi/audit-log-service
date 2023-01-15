package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (svc *service) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(svc.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(svc.methodNotAllowed)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", svc.healthcheck)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/register", svc.registerService)
	router.HandlerFunc(http.MethodPatch, "/v1/tokens/reset", svc.requiredAuthenticatedService(svc.resetToken))

	router.HandlerFunc(http.MethodGet, "/v1/audit", svc.requiredAuthenticatedService(svc.auditTrail))

	return svc.recoverPanic(svc.authenticate(router))
}
