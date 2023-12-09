package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"

	"github.com/mitchellh/mapstructure"
	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/adapter/restapi"
	"github.com/go-debit/internal/repository/postgre"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 		*postgre.WorkerRepository
	restapi					*restapi.RestApiSConfig
}

func NewWorkerService(	workerRepository 	*postgre.WorkerRepository,
						restapi				*restapi.RestApiSConfig) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
		restapi:			restapi,
	}
}

func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, debit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Service.Add")

	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		root.Close(nil)
	}()

	childLogger.Debug().Interface("debit:",debit).Msg("")

	debit.Type = "DEBIT"
	if debit.Amount > 0 {
		return nil, erro.ErrInvalidAmount
	}

	rest_interface_data, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain, "/get", debit.AccountID )
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

	debit.FkAccountID = account_parsed.ID
	res, err := s.workerRepository.Add(ctx, tx, debit)
	if err != nil {
		return nil, err
	}

	childLogger.Debug().Interface("debit:",debit).Msg("")

	_, err = s.restapi.PostData(ctx, s.restapi.ServerUrlDomain ,"/add/fund", debit)
	if err != nil {
		return nil, err
	}

	// Get financial script
	script := "script.debit"
	res_script, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain2 ,"/script/get", script)
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
	for _, v := range script_parsed.Fee {
		childLogger.Debug().Interface("v:",v).Msg("")

		res_fee, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain2 ,"/key/get", v)
		if err != nil {
			return nil, err
		}
		childLogger.Debug().Interface("res_fee:",res_fee).Msg("")
	}

	return res, nil
}
func (s WorkerService) List(ctx context.Context, debit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	_, root := xray.BeginSubsegment(ctx, "Service.List")
	defer root.Close(nil)

	rest_interface_data, err := s.restapi.GetData(ctx, s.restapi.ServerUrlDomain ,debit.AccountID, "/get")
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
	res, err := s.workerRepository.List(ctx, debit)
	if err != nil {
		return nil, err
	}

	return res, nil
}