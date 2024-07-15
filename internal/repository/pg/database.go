package pg

import (
	"context"
	"fmt"
	"time"
	"errors"

	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/lib"
	"github.com/go-debit/internal/erro"

	"github.com/rs/zerolog/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var childLogger = log.With().Str("repository.pg", "WorkerRepo").Logger()

type DatabasePG interface {
	GetConnection() (*pgxpool.Pool)
}

type DatabasePGServer struct {
	connPool   	*pgxpool.Pool
}

func NewDatabasePGServer(ctx context.Context, databaseRDS *core.DatabaseRDS) (DatabasePG, error) {
	childLogger.Debug().Msg("NewDatabasePGServer")
	
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
							databaseRDS.User, 
							databaseRDS.Password, 
							databaseRDS.Host, 
							databaseRDS.Port, 
							databaseRDS.DatabaseName) 
							
	connPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return DatabasePGServer{}, err
	}
	
	err = connPool.Ping(ctx)
	if err != nil {
		return DatabasePGServer{}, err
	}

	return DatabasePGServer{
		connPool: connPool,
	}, nil
}

func (d DatabasePGServer) GetConnection() (*pgxpool.Pool) {
	childLogger.Debug().Msg("GetConnection")
	return d.connPool
}

func (d DatabasePGServer) CloseConnection() {
	childLogger.Debug().Msg("CloseConnection")
	defer d.connPool.Close()
}

type WorkerRepository struct {
	databasePG DatabasePG
}

func NewWorkerRepository(databasePG DatabasePG) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databasePG: databasePG,
	}
}

func (w WorkerRepository) SetSessionVariable(ctx context.Context, userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")

	connPool := w.databasePG.GetConnection()
	
	_, err := connPool.Query(ctx, "SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")

	connPool := w.databasePG.GetConnection()

	var res_balance string
	rows, err := connPool.Query(ctx, "SELECT current_setting('sess.user_credential')" )
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

func (w WorkerRepository) StartTx(ctx context.Context) (pgx.Tx, error) {
	childLogger.Debug().Msg("StartTx")

	conn := w.databasePG.GetConnection()
	tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, errors.New(err.Error())
    }

	return tx, nil
}

func (w WorkerRepository) Add(ctx context.Context, tx pgx.Tx, debit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "repo.Add")	
    defer span.End()

	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(ctx, query, debit.FkAccountID, debit.Type, time.Now(), debit.Currency, debit.Amount, debit.TenantID)								
	
	var id int

	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	res_debit := core.AccountStatement{}
	res_debit.ID = id
	res_debit.ChargeAt = time.Now()

	return &res_debit , nil
}

func (w WorkerRepository) List(ctx context.Context, debit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "repo.List")	
    defer span.End()

	conn:= w.databasePG.GetConnection()
	
	result_query := core.AccountStatement{}
	balance_list := []core.AccountStatement{}

	query := `SELECT id, 
					fk_account_id, 
					type_charge,
					charged_at,
					currency, 
					amount,																										
					tenant_id	
					FROM account_statement 
					WHERE fk_account_id =$1 and type_charge= $2 order by charged_at desc`

	rows, err := conn.Query(ctx, query, debit.FkAccountID, debit.Type )
		if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.FkAccountID, 
							&result_query.Type, 
							&result_query.ChargeAt,
							&result_query.Currency,
							&result_query.Amount,
							&result_query.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	defer rows.Close()
	return &balance_list , nil
}

func (w WorkerRepository) AddAccountStatementFee(ctx context.Context, tx pgx.Tx,accountStatementFee core.AccountStatementFee) (*core.AccountStatementFee, error){
	childLogger.Debug().Msg("AddAccountStatementFee")

	span := lib.Span(ctx, "repo.AddAccountStatementFee")	
    defer span.End()

	query := `INSERT INTO account_statement_fee (fk_account_statement_id, 
												charged_at,
												type_fee,
												value_fee,
												currency,
												amount,
												tenant_id) 
				VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	accountStatementFee.ChargeAt = time.Now()

	row := tx.QueryRow(ctx, query, accountStatementFee.FkAccountStatementID,
									accountStatementFee.ChargeAt,
									accountStatementFee.TypeFee,
									accountStatementFee.ValueFee,
									accountStatementFee.Currency,
									accountStatementFee.Amount,
									accountStatementFee.TenantID)								

	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}
	accountStatementFee.ID = id

	return &accountStatementFee , nil
}