package postgre

import (
	"context"
	"fmt"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/go-debit/internal/core"
)

type DatabaseHelper interface {
	GetConnection() (*sql.DB)
}

type DatabaseHelperImplementacion struct {
	client   	*sql.DB
}

func NewDatabaseHelper(	ctx context.Context, 
						databaseRDS *core.DatabaseRDS) (DatabaseHelper, error) {
	childLogger.Debug().Msg("NewDatabaseHelper")
	
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
							databaseRDS.User, 
							databaseRDS.Password, 
							databaseRDS.Host, 
							databaseRDS.Port, 
							databaseRDS.DatabaseName) 
							
	client, err := sql.Open(databaseRDS.Postgres_Driver, connStr)
	if err != nil {
		return DatabaseHelperImplementacion{}, err
	}
	
	err = client.PingContext(ctx)
	if err != nil {
		return DatabaseHelperImplementacion{}, err
	}

	return DatabaseHelperImplementacion{
		client: client,
	}, nil
}

func (d DatabaseHelperImplementacion) GetConnection() (*sql.DB) {
	childLogger.Debug().Msg("GetConnection")
	return d.client
}
