package postgre

import (
	"context"
	"errors"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"database/sql"

	"github.com/go-debit/internal/erro"
)

var childLogger = log.With().Str("repository", "WorkerRepository").Logger()

type WorkerRepository struct {
	databaseHelper DatabaseHelper
}

func NewWorkerRepository(databaseHelper DatabaseHelper) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databaseHelper: databaseHelper,
	}
}

func (w WorkerRepository) SetSessionVariable(ctx context.Context,userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()
	
	stmt, err := client.Prepare("SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()

	var res_balance string
	rows, err := client.QueryContext(ctx, "SELECT current_setting('sess.user_credential')" )
	if err != nil {
		childLogger.Error().Err(err).Msg("Prepare statement")
		return "", errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_balance )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return "", errors.New(err.Error())
        }
		return res_balance, nil
	}

	return "", erro.ErrNotFound
}

func (w WorkerRepository) StartTx(ctx context.Context) (*sql.Tx, error) {
	childLogger.Debug().Msg("StartTx")

	client := w.databaseHelper.GetConnection()

	tx, err := client.BeginTx(ctx, &sql.TxOptions{})
    if err != nil {
        return nil, errors.New(err.Error())
    }

	return tx, nil
}

func (w WorkerRepository) Ping(ctx context.Context) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("Ping")
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")

	client := w.databaseHelper.GetConnection()

	err := client.PingContext(ctx)
	if err != nil {
		return false, errors.New(err.Error())
	}

	return true, nil
}
