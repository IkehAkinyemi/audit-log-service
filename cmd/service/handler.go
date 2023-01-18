package main

import (
	"errors"
	"net/http"

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

// registerService maps to "POST /v1/tokens/register". Registers a service
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

	if input.ServiceID == "" {
		svc.badRequestResponse(w, r, errors.New("service_id must be provided"))
		return
	}

	token, err := svc.db.NewAPIKey(r.Context(), input.ServiceID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrDuplicateService):
			svc.failedValidationResponse(w, r, map[string]string{
				"message": "a service with serviceID already exists",
			})
		default:
			svc.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"api_key": token}, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
	}

	svc.logger.PrintInfo("New API key generated", map[string]string{
		"service_id": input.ServiceID,
	})
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
	svc.logger.PrintInfo("API key updated", map[string]string{
		"service_id": input.ServiceID,
	})
}

// auditTrail maps to "GET /v1/audit-trail?<query_string>".
// Retrieves logs based on the query_string values.
func (svc *service) GetLogs(w http.ResponseWriter, r *http.Request) {
	var input utils.Filters

	v := utils.NewValidator()

	query := r.URL.Query()
	input.Action = utils.ReadStr(query, "action", "")
	input.ActorID = utils.ReadStr(query, "actor_id", "")
	input.ActorType = utils.ReadStr(query, "actor_type", "")
	input.EntityType = utils.ReadStr(query, "entity_type", "")
	input.StartTimestamp = utils.ParseTime(query, "start_timestamp")
	input.EndTimestamp = utils.ParseTime(query, "end_timestamp")
	input.SortField, input.SortDescending = utils.SortValues(query, "sort")
	input.Page = utils.ReadInt(query, "page", 1, v)
	input.PageSize = utils.ReadInt(query, "page_size", 20, v)

	if utils.ValidateFilters(v, input); !v.Valid() {
		svc.failedValidationResponse(w, r, v.Errors)
		return
	}

	logs, metadata, err := svc.db.GetAllLogs(r.Context(), input)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"logs": logs, "metadata": metadata}, nil)
	if err != nil {
		svc.serverErrorResponse(w, r, err)
	}

	svc.logger.PrintInfo("Queried for logs", map[string]string{
		"service_id":   string(*svc.contextGetService(r)),
		"query_string": r.URL.String(),
	})
}
