package controller

import (	
	"net/http"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-debit/internal/service"
	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/lib"
)

var childLogger = log.With().Str("handler", "controller").Logger()

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")

	return HttpWorkerAdapter{
		workerService: workerService,
	}
}

type APIError struct {
	StatusCode	int  `json:"statusCode"`
	Msg			string `json:"msg"`
}

func (e APIError) Error() string {
	return e.Msg
}

func NewAPIError(statusCode int, err error) APIError {
	return APIError{
		StatusCode: statusCode,
		Msg:		err.Error(),
	}
}

func (h *HttpWorkerAdapter) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
}

func (h *HttpWorkerAdapter) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
}

func (h *HttpWorkerAdapter) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
}

func (h *HttpWorkerAdapter) Add( rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("Add")

	span := lib.Span(req.Context(), "handler.Add")
	defer span.End()

	debit := core.AccountStatement{}
	err := json.NewDecoder(req.Body).Decode(&debit)
    if err != nil {
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrUnmarshal)
		return apiError
    }

	defer req.Body.Close()

	res, err := h.workerService.Add(req.Context(), &debit)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) List(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("List")

	span := lib.Span(req.Context(), "handler.List")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	debit := core.AccountStatement{}
	debit.AccountID = varID
	
	res, err := h.workerService.List(req.Context(), &debit)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func WriteJSON(rw http.ResponseWriter, code int, v any) error{
	rw.WriteHeader(code)
	return json.NewEncoder(rw).Encode(v)
}