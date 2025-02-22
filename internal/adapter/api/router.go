package api

import (
	"encoding/json"
	"net/http"
	"github.com/rs/zerolog/log"
	"github.com/go-debit/internal/core/service"
	"github.com/go-debit/internal/core/model"
	"github.com/go-debit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"
	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"
)

var childLogger = log.With().Str("adapter", "api.router").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var core_tools go_core_tools.ToolsCore
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	return HttpRouters{
		workerService: workerService,
	}
}

func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
}

func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
}

func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
}

func (h *HttpRouters) AddDebit(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("AddDebit")

	//trace
	span := tracerProvider.Span(req.Context(), "adapter.api.AddDebit")
	defer span.End()

	// prepare body
	debit := model.AccountStatement{}
	err := json.NewDecoder(req.Body).Decode(&debit)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	//call service
	res, err := h.workerService.AddDebit(req.Context(), &debit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		case erro.ErrTransInvalid:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)
		case erro.ErrInvalidAmount:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)	
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) ListDebit(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("ListDebit")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListDebit")
	defer span.End()

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	debit := model.AccountStatement{}
	debit.AccountID = varID

	// call service
	res, err := h.workerService.ListDebit(req.Context(), &debit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) ListDebitPerDate(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("ListDebitPerDate")

	//Trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListDebitPerDate")
	defer span.End()

	// parameter
	params := req.URL.Query()
	varAcc := params.Get("account")
	varDate := params.Get("date_start")

	debit := model.AccountStatement{}
	debit.AccountID = varAcc

	convertDate, err := core_tools.ConvertToDate(varDate)
	if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
		return &core_apiError
	}
	debit.ChargeAt = *convertDate

	//service
	res, err := h.workerService.ListDebitPerDate(req.Context(), &debit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}