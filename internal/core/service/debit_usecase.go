package service

import(
	"fmt"
	"time"
	"context"
	"net/http"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

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

// About add credit
func (s *WorkerService) AddDebit(ctx context.Context, debit *model.AccountStatement) (*model.AccountStatement, error){
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("AddDebit")
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddDebit")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

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
														&trace_id,
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
											&trace_id, 
											debit)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	//Open CB - MOCK
	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	_, errCB := circuitBreaker.Execute(func() (interface{}, error) {		
		
		// Add accountStamentFee
		accountStatementFee := model.AccountStatementFee{}
		accountStatementFee.FkAccountStatementID = debit.ID
		accountStatementFee.Currency = debit.Currency
		accountStatementFee.Amount	 = debit.Amount
		accountStatementFee.TenantID = debit.TenantID

		_, errCB := s.AddAccountStatementFee(ctx, tx , accountStatementFee)
		if errCB != nil {
			return nil, errCB
		}

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
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("ListDebit")
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListDebit")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + debit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id,
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
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("ListDebitPerDate")
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("debit: ", debit).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListDebit'PerDate")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// Get the Account ID from Account-service
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[0].Url + "/" + debit.AccountID,
														s.apiService[0].Method,
														&s.apiService[0].Header_x_apigw_api_id,
														nil,
														&trace_id,
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

func (s *WorkerService) AddAccountStatementFee(ctx context.Context, tx pgx.Tx, accountStatementFee model.AccountStatementFee) (*model.AccountStatementFee, error){
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Msg("AddAccountStatementFee")
	childLogger.Info().Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("accountStatementFee: ", accountStatementFee).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddAccountStatementFee")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()

	// Get financial script
	script := "script.debit"
	res_payload, statusCode, err := apiService.CallApi(ctx,
														s.apiService[2].Url + "/" + script,
														s.apiService[2].Method,
														&s.apiService[2].Header_x_apigw_api_id,
														nil,
														&trace_id,
														nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	// Unmarshall to struct
	jsonString, err := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var script_parsed model.Script
	json.Unmarshal(jsonString, &script_parsed)
	
	// Get all fees
	for _, v_fee := range script_parsed.Fee {
		res_fee, statusCode, err := apiService.CallApi(ctx,
														s.apiService[3].Url + "/" + v_fee,
														s.apiService[3].Method,
														&s.apiService[3].Header_x_apigw_api_id,
														nil,
														&trace_id,
														nil)
		if err != nil {
			return nil, errorStatusCode(statusCode)
		}

		// Unmarshall to struct
		jsonString, err := json.Marshal(res_fee)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		var fee_parsed model.Fee
		json.Unmarshal(jsonString, &fee_parsed)

		// Prepare to insert AccountStatementFee
		new_accountStatementFee := accountStatementFee
		new_accountStatementFee.TypeFee = fee_parsed.Name
		new_accountStatementFee.ValueFee = fee_parsed.Value
		new_accountStatementFee.ChargeAt = time.Now()
		new_accountStatementFee.Amount	= (accountStatementFee.Amount * (fee_parsed.Value/100))

		_, err = s.workerRepository.AddAccountStatementFee(ctx, tx, new_accountStatementFee)
		if err != nil {
			return nil, err
		}
	}

	childLogger.Debug().Interface("script_parsed:",script_parsed).Msg("")

	return &accountStatementFee, nil
}