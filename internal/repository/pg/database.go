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
	Acquire(context.Context) (*pgxpool.Conn, error)
	Release(*pgxpool.Conn)
}

type DatabasePGServer struct {
	connPool   	*pgxpool.Pool
}

func Config(database_url string) (*pgxpool.Config) {
	const defaultMaxConns = int32(10)
	const defaultMinConns = int32(5)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5
   
	dbConfig, err := pgxpool.ParseConfig(database_url)
	if err!=nil {
		childLogger.Error().Err(err).Msg("Failed to create a config")
	}
   
	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout
   
	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		childLogger.Debug().Msg("Before acquiring connection pool !")
	 	return true
	}
   
	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		childLogger.Debug().Msg("After releasing connection pool !")
	 	return true
	}
   
	dbConfig.BeforeClose = func(c *pgx.Conn) {
		childLogger.Debug().Msg("Closed connection pool !")
	}
   
	return dbConfig
}

func NewDatabasePGServer(ctx context.Context, databaseRDS *core.DatabaseRDS) (DatabasePG, error) {
	childLogger.Debug().Msg("NewDatabasePGServer")
	
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
							databaseRDS.User, 
							databaseRDS.Password, 
							databaseRDS.Host, 
							databaseRDS.Port, 
							databaseRDS.DatabaseName) 
							
	connPool, err := pgxpool.NewWithConfig(ctx, Config(connStr))
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

func (d DatabasePGServer) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	childLogger.Debug().Msg("Acquire")
	
	span := lib.Span(ctx, "repo.Acquire")
	defer span.End()

	connection, err := d.connPool.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Error while acquiring connection from the database pool!!")
		return nil, err
	} 
	return connection, nil
}

func (d DatabasePGServer) Release(connection *pgxpool.Conn) {
	childLogger.Debug().Msg("Release")
	defer connection.Release()
}

func (d DatabasePGServer) GetConnection() (*pgxpool.Pool) {
	childLogger.Debug().Msg("GetConnection")
	return d.connPool
}

func (d DatabasePGServer) CloseConnection() {
	childLogger.Debug().Msg("CloseConnection")
	defer d.connPool.Close()
}

//-----------------------------------------------
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

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return false, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)
	
	_, err = conn.Query(ctx, "SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (*string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)

	var res_balance string
	rows, err := conn.Query(ctx, "SELECT current_setting('sess.user_credential')" )
	if err != nil {
		childLogger.Error().Err(err).Msg("Prepare statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_balance )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &res_balance, nil
	}

	return nil, erro.ErrNotFound
}

func (w WorkerRepository) StartTx(ctx context.Context) (pgx.Tx, *pgxpool.Conn,error) {
	childLogger.Debug().Msg("StartTx")

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, nil, errors.New(err.Error())
	}

	tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, nil ,errors.New(err.Error())
    }

	return tx, conn, nil
}

func (w WorkerRepository) ReleaseTx(connection *pgxpool.Conn) {
	childLogger.Debug().Msg("ReleaseTx")

	defer connection.Release()
}
//-----------------------------------------------
func (w WorkerRepository) Add(ctx context.Context, tx pgx.Tx, debit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "repo.Add")	
    defer span.End()

	debit.ChargeAt = time.Now()

	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(ctx, query, debit.FkAccountID, debit.Type, debit.ChargeAt, debit.Currency, debit.Amount, debit.TenantID)								
	
	var id int

	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	debit.ID = id 

	return &debit , nil
}

func (w WorkerRepository) List(ctx context.Context, debit core.AccountStatement) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "repo.List")	
    defer span.End()

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)
	
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