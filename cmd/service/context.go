package main

import (
	"context"
	"net/http"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
)

type contextKey string

var servicContextKey = contextKey("service")

// contextSetService registers an authenticated service per connection
func (svc *service) contextSetService(r *http.Request, serviceID *model.ServiceID) *http.Request {
	ctx := context.WithValue(r.Context(), servicContextKey, serviceID)
	return r.WithContext(ctx)
}

// contextGetService retrieves an authenticated service.
func (svc *service) contextGetService(r *http.Request) *model.ServiceID {
	serviceID, ok := r.Context().Value(servicContextKey).(*model.ServiceID)
	if !ok {
		panic("missing service value in request context")
	}

	return serviceID
}
