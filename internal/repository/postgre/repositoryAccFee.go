package postgre

import (
	"context"
	"time"
	"errors"

	_ "github.com/lib/pq"
	"database/sql"

	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/core"
	"github.com/go-debit/internal/lib"
)

func (w WorkerRepository) AddAccountStatementFee(ctx context.Context, tx *sql.Tx ,accountStatementFee core.AccountStatementFee) (*core.AccountStatementFee, error){
	childLogger.Debug().Msg("AddAccountStatementFee")

	span := lib.Span(ctx, "repo.AddAccountStatementFee")	
    defer span.End()

	stmt, err := tx.Prepare(`INSERT INTO account_statement_fee ( 	fk_account_statement_id, 
																	charged_at,
																	type_fee,
																	value_fee,
																	currency,
																	amount,
																	tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6, $7) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}
	
	result, err := stmt.ExecContext(	ctx,
								accountStatementFee.FkAccountStatementID, 
								time.Now(),
								accountStatementFee.TypeFee,
								accountStatementFee.ValueFee,
								accountStatementFee.Currency,
								accountStatementFee.Amount,
								accountStatementFee.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}
	
	inserted, errRows := result.RowsAffected()
    if errRows != nil {
        return nil, errors.New(err.Error())
    } else if inserted == 0 {
		return nil, erro.ErrInsert
    }

	defer stmt.Close()
	return &accountStatementFee , nil
}
