package service

import (
	"time"
	"context"
	"errors"
	"github.com/rs/zerolog/log"

	"github.com/mitchellh/mapstructure"
	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/lib"
	"github.com/go-debit/internal/adapter/restapi"
	"github.com/go-debit/internal/repository/postgre"
	"github.com/go-debit/internal/repository/pg"
	"github.com/sony/gobreaker"
)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepo		 		*pg.WorkerRepository
	workerRepository 		*postgre.WorkerRepository
	restEndpoint			*core.RestEndpoint
	restApiService			*restapi.RestApiService
	circuitBreaker			*gobreaker.CircuitBreaker
}

func NewWorkerService(	workerRepository 	*postgre.WorkerRepository,
						workerRepo		 	*pg.WorkerRepository,
						restEndpoint		*core.RestEndpoint,
						restApiService		*restapi.RestApiService,
						circuitBreaker		*gobreaker.CircuitBreaker) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		workerRepo: 		workerRepo,
		restEndpoint:		restEndpoint,
		restApiService:		restApiService,
		circuitBreaker: 	circuitBreaker,
	}
}

func (s WorkerService) SetSessionVariable(	ctx context.Context, 
											userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(	ctx, 
														userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(	ctx context.Context, debit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("--------------- Add ------------------------")
	childLogger.Debug().Interface("1) debit :",debit).Msg("")

	span := lib.Span(ctx, "service.Add")	

	tx, err := s.workerRepo.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		span.End()
	}()

	debit.Type = "DEBIT"
	if debit.Amount > 0 {
		err = erro.ErrInvalidAmount
		return nil, err
	}

	// Get account data
	urlDomain := s.restEndpoint.ServiceUrlDomain + "/get/"  + debit.AccountID
	rest_interface_data, err := s.restApiService.GetData(ctx, urlDomain, s.restEndpoint.XApigwId, debit.AccountID )
	if err != nil {
		return nil, err
	}

	var account_parsed core.Account
	err = mapstructure.Decode(rest_interface_data, &account_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

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

	urlDomain = s.restEndpoint.ServiceUrlDomain + "/add/fund"
	_, err = s.restApiService.PostData(ctx, urlDomain, s.restEndpoint.XApigwId, debit)
	if err != nil {
		return nil, err
	}
	
	// Get financial script
	urlDomain = s.restEndpoint.ServiceUrlDomainPayFee + "/script/get/script.debit"
	script := "script.debit"
	res_script, err := s.restApiService.GetData(ctx, 
												urlDomain, 
												s.restEndpoint.XApigwIdPayFee,
												script)
	if err != nil {
		return nil, err
	}

	var script_parsed core.Script
	err = mapstructure.Decode(res_script, &script_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	childLogger.Debug().Interface("script_parsed:",script_parsed).Msg("")

	// Get the fees
	_, err = s.circuitBreaker.Execute(func() (interface{}, error) {
		for _, v := range script_parsed.Fee {
			childLogger.Debug().Interface("v:",v).Msg("")
	
			urlDomain = s.restEndpoint.ServiceUrlDomainPayFee + "/key/get/" + v
			res_fee, err := s.restApiService.GetData(ctx, urlDomain,s.restEndpoint.XApigwIdPayFee, v)
			if err != nil {
				return nil, err
			}
			childLogger.Debug().Interface("res_fee:",res_fee).Msg("")
	
			var fee_parsed core.Fee
			err = mapstructure.Decode(res_fee, &fee_parsed)
			if err != nil {
				childLogger.Error().Err(err).Msg("error parse interface")
				return nil, errors.New(err.Error())
			}
	
			accountStatementFee := core.AccountStatementFee{}
			accountStatementFee.FkAccountStatementID = res.ID
			accountStatementFee.TypeFee = fee_parsed.Name
			accountStatementFee.ValueFee = fee_parsed.Value
			accountStatementFee.ChargeAt = time.Now()
			accountStatementFee.Currency = debit.Currency
			accountStatementFee.Amount	 = (debit.Amount * (fee_parsed.Value/100))
			accountStatementFee.TenantID = debit.TenantID
	
			_, err = s.workerRepo.AddAccountStatementFee(ctx, tx, accountStatementFee)
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

	return &debit, nil
}

func (s WorkerService) List(ctx context.Context, debit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "service.List")	
    defer span.End()

	urlDomain := s.restEndpoint.ServiceUrlDomain + "/get/" + debit.AccountID
	rest_interface_data, err := s.restApiService.GetData(ctx, urlDomain, s.restEndpoint.XApigwId, debit.AccountID)
	if err != nil {
		return nil, err
	}

	var account_parsed core.Account
	err = mapstructure.Decode(rest_interface_data, &account_parsed)
    if err != nil {
		childLogger.Error().Err(err).Msg("error parse interface")
		return nil, errors.New(err.Error())
    }

	debit.FkAccountID = account_parsed.ID
	debit.Type = "DEBIT"

	res, err := s.workerRepo.List(ctx, debit)
	if err != nil {
		return nil, err
	}

	return res, nil
}