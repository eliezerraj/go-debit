package service

import(
	"context"
	"net/http"
	"encoding/json"
	"errors"

	"github.com/go-debit/internal/infra/circuitbreaker"
	"github.com/go-debit/internal/core/model"
	"github.com/go-debit/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
	case http.StatusUnauthorized:
		err = erro.ErrUnauthorized
	case http.StatusForbidden:
		err = erro.ErrHTTPForbiden
	case http.StatusNotFound:
		err = erro.ErrNotFound
	default:
		err = erro.ErrServer
	}
	return err
}

func (s *WorkerService) AddDebit(ctx context.Context, debit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Debug().Msg("AddDebit")
	childLogger.Debug().Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddDebit")
	defer span.End()
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Business rules
	if debit.Type != "DEBIT" {
		return nil, erro.ErrTransInvalid
	}
	if debit.Amount > 0 {
		return nil, erro.ErrInvalidAmount
	}

	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + debit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	debit.FkAccountID = account_parsed.ID
	
	// Get transaction UUID 
	res_uuid, err := s.workerRepository.GetTransactionUUID(ctx)
	if err != nil {
		return nil, err
	}
	debit.TransactionID = res_uuid

	// Add the credit
	res, err := s.workerRepository.AddDebit(ctx, tx, debit)
	if err != nil {
		return nil, err
	}

	// Add (POST) the account statement Get the Account ID from Account-service
	_, statusCode, err = apiService.CallApi(ctx,
											s.apiService[1].Url,
											s.apiService[1].Method,
											&s.apiService[1].Header_x_apigw_api_id,
											nil, 
											debit)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	//Open CB - MOCK
	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	_, errCB := circuitBreaker.Execute(func() (interface{}, error) {		
		// Get financial script
		script := "script.debit"
		res_payload, statusCode, errCB := apiService.CallApi(ctx,
															s.apiService[2].Url + "/" + script,
															s.apiService[2].Method,
															&s.apiService[2].Header_x_apigw_api_id,
															nil, 
															nil)
		if errCB != nil {
			return nil, errorStatusCode(statusCode)
		}
		jsonString, errCB = json.Marshal(res_payload)
		if errCB != nil {
			childLogger.Error().Err(errCB).Msg("error Marshal")
			return nil, errors.New(errCB.Error())
		}
		var script_parsed model.Script
		json.Unmarshal(jsonString, &script_parsed)

		childLogger.Debug().Interface("script_parsed:",script_parsed).Msg("")

		return nil, nil
	})
	if (errCB != nil) {
		childLogger.Debug().Msg("--------------------------------------------------")
		childLogger.Error().Err(err).Msg(" ****** Circuit Breaker OPEN !!! ******")
		childLogger.Debug().Msg("--------------------------------------------------")

		res.Obs =  "circuit breaker open impossible to reach the pay fees !!!"
	}

	return res, nil
}

func (s *WorkerService) ListDebit(ctx context.Context, debit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListDebit")
	childLogger.Debug().Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListDebit")
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + debit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	debit.FkAccountID = account_parsed.ID
	debit.Type = "DEBIT"

	res, err := s.workerRepository.ListDebit(ctx, debit)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *WorkerService) ListDebitPerDate(ctx context.Context, debit *model.AccountStatement) (*[]model.AccountStatement, error){
	childLogger.Debug().Msg("ListDebitPerDate")
	childLogger.Debug().Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListDebit'PerDate")
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + debit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil, 
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// Business rule
	debit.FkAccountID = account_parsed.ID
	debit.Type = "DEBIT"

	res, err := s.workerRepository.ListDebitPerDate(ctx, debit)
	if err != nil {
		return nil, err
	}

	return res, nil
}