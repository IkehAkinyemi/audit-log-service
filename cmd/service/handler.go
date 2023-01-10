package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
)

// healthcheck maps to "GET /v1/healthcheck". Return info about the server state.
func (svc *service) healthcheck(w http.ResponseWriter, r *http.Request) {
	data := utils.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviroment": svc.config.Env,
			"version":    "1.0.0",
		},
	}

	err := utils.WriteJSON(w, http.StatusOK, data, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
	}
}

// registerService maps to "POST /v1/tokens/register". Saves a service
// and return API Key for the service.
func (svc *service) registerService(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ServiceID string `json:"service_id"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		svc.badRequestResponse(w, r, err)
		return
	}

	token, err := svc.db.NewAPIKey(r.Context(), input.ServiceID)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"api_key": token}, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
	}
}

// resetToken maps to "PATCH /v1/tokens/reset". Revokes existing API Key
// and generates a new API Key for calling service.
func (svc *service) resetToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ServiceID string `json:"service_id"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		svc.badRequestResponse(w, r, err)
		return
	}

	token, err := svc.db.UpdateToken(r.Context(), model.ServiceID(input.ServiceID))
	if err != nil {
		switch {
		case errors.Is(err, model.ErrRecordNotFound):
			svc.invalidCredentialResponse(w, r)
		default:
			svc.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"api_key": token}, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
	}
}

func (svc *service) AddEventLog(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Timestamp time.Time      `json:"created_at"`
		Action    string         `json:"action"`
		Actor     model.Actor    `json:"actor"`
		Entity    model.Entity   `json:"entity"`
		Context   model.Context  `json:"context"`
		Extension map[string]any `json:"extension,omitempty"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		svc.badRequestResponse(w, r, err)
		return
	}

	eventLog := &model.AuditEvent{
		Timestamp: input.Timestamp,
		Action:    input.Action,
		Actor:     input.Actor,
		Entity:    input.Entity,
		Context:   input.Context,
		Extension: input.Extension,
	}

	v := utils.NewValidator()

	if utils.ValidateAuditEvent(v, eventLog); !v.Valid() {
		svc.failedValidationResponse(w, r, v.Errors)
	}

	_, err = svc.db.AddLog(r.Context(), eventLog)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"message": "resource created"}, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
		return
	}
}

func (svc *service) auditTrail(w http.ResponseWriter, r *http.Request) {
	// TODO
	// - Retrieve audit trail based on query string
}

func (svc *service) showEventLog(w http.ResponseWriter, r *http.Request) {
	// TODO
	// - Retrieve a log using its uuid
}
