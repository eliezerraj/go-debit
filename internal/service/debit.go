package service

import (
	"time"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"encoding/json"

	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/lib"
	"github.com/go-debit/internal/adapter/restapi"
	"github.com/go-debit/internal/repository/storage"
	"github.com/sony/gobreaker"
)

var childLogger = log.With().Str("service", "service").Logger()
var restApiCallData core.RestApiCallData

type WorkerService struct {
	workerRepo	*storage.WorkerRepository
	appServer	*core.AppServer
	restApiService	*restapi.RestApiService
	circuitBreaker	*gobreaker.CircuitBreaker
}

func NewWorkerService(	workerRepo	*storage.WorkerRepository,
						appServer	*core.AppServer,
						restApiService	*restapi.RestApiService,
						circuitBreaker	*gobreaker.CircuitBreaker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepo: 		workerRepo,
		appServer:			appServer,
		restApiService:		restApiService,
		circuitBreaker: 	circuitBreaker,
	}
}

func (s WorkerService) Add(	ctx context.Context, debit *core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("--------------- Add ------------------------")
	childLogger.Debug().Interface("1) debit :",debit).Msg("")

	span := lib.Span(ctx, "service.Add")	
	tx, conn, err := s.workerRepo.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepo.ReleaseTx(conn)
		span.End()
	}()

	debit.Type = "DEBIT"
	if debit.Amount > 0 {
		err = erro.ErrInvalidAmount
		return nil, err
	}

	// Get account data
	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + debit.AccountID
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	rest_interface_acc_from, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error CallApiRest /fundBalanceAccount")
		return nil, err
	}
	jsonString, err  := json.Marshal(rest_interface_acc_from)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed core.Account
	json.Unmarshal(jsonString, &account_parsed)

	childLogger.Debug().Interface("account_parsed:",account_parsed).Msg("")

	// Add the Data 
	debit.FkAccountID = account_parsed.ID
	res, err := s.workerRepo.Add(ctx, tx, debit)
	if err != nil {
		return nil, err
	}
	childLogger.Debug().Interface("2) debit:",debit).Msg("")
	debit.ID = res.ID
	debit.ChargeAt = res.ChargeAt

	restApiCallData.Method = "POST"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/add/fund"
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	_, err = s.restApiService.CallApiRest(ctx, restApiCallData, debit)
	if err != nil {
		childLogger.Error().Err(err).Msg("error CallApiRest/add/fund")
		return nil, err
	}

	// Get financial script
	script := "script.debit"

	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomainPayFee + "/script/get/" + script
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwIdPayFee

	res_script, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error CallApiRest /key/get/")
		return nil, err
	}
	
	jsonString, err = json.Marshal(res_script)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
	}
	var script_parsed core.Script
	json.Unmarshal(jsonString, &script_parsed)

	childLogger.Debug().Interface("script_parsed:",script_parsed).Msg("")

	// Get the fees
	_, err = s.circuitBreaker.Execute(func() (interface{}, error) {
		for _, v := range script_parsed.Fee {
			childLogger.Debug().Interface("v:",v).Msg("")
	
			restApiCallData.Method = "GET"
			restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomainPayFee + "/key/get/" + v
			restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwIdPayFee
		
			res_fee, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
			if err != nil {
				childLogger.Error().Err(err).Msg("error CallApiRest /key/get/")
				return nil, err
			}
			
			childLogger.Debug().Interface("res_fee:",res_fee).Msg("")

			jsonString, err  := json.Marshal(res_fee)
			if err != nil {
				childLogger.Error().Err(err).Msg("error Marshal")
				return nil, errors.New(err.Error())
			}
			var fee_parsed core.Fee
			json.Unmarshal(jsonString, &fee_parsed)

			accountStatementFee := core.AccountStatementFee{}
			accountStatementFee.FkAccountStatementID = res.ID
			accountStatementFee.TypeFee = fee_parsed.Name
			accountStatementFee.ValueFee = fee_parsed.Value
			accountStatementFee.ChargeAt = time.Now()
			accountStatementFee.Currency = debit.Currency
			accountStatementFee.Amount	 = (debit.Amount * (fee_parsed.Value/100))
			accountStatementFee.TenantID = debit.TenantID
	
			_, err = s.workerRepo.AddAccountStatementFee(ctx, tx, &accountStatementFee)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if (err != nil) {
		childLogger.Debug().Msg("--------------------------------------------------")
		childLogger.Error().Err(err).Msg(" ****** Circuit Breaker OPEN !!! ******")
		childLogger.Debug().Msg("--------------------------------------------------")
	}

	return debit, nil
}

func (s WorkerService) List(ctx context.Context, debit *core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "service.List")	
    defer span.End()

	restApiCallData.Method = "GET"
	restApiCallData.Url = s.appServer.RestEndpoint.ServiceUrlDomain + "/get/" + debit.AccountID
	restApiCallData.X_Api_Id = &s.appServer.RestEndpoint.XApigwId

	rest_interface_acc_from, err := s.restApiService.CallApiRest(ctx, restApiCallData, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error CallApiRest /get/")
		return nil, err
	}
	jsonString, err  := json.Marshal(rest_interface_acc_from)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Marshal")
		return nil, errors.New(err.Error())
    }
	var account_parsed core.Account
	json.Unmarshal(jsonString, &account_parsed)

	debit.FkAccountID = account_parsed.ID
	debit.Type = "DEBIT"

	res, err := s.workerRepo.List(ctx, debit)
	if err != nil {
		return nil, err
	}

	return res, nil
}