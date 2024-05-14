package util

import(
	"os"
	"io/ioutil"

	"github.com/joho/godotenv"
	"github.com/go-debit/internal/core"
)

func GetDatabaseEnv() core.DatabaseRDS {
	childLogger.Debug().Msg("GetDatabaseEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("No .env File !!!!")
	}
	
	var databaseRDS	core.DatabaseRDS
	databaseRDS.Db_timeout = 90
	
	if os.Getenv("DB_HOST") !=  "" {
		databaseRDS.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		databaseRDS.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_NAME") !=  "" {	
		databaseRDS.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		databaseRDS.Schema = os.Getenv("DB_SCHEMA")
	}
	if os.Getenv("DB_DRIVER") !=  "" {	
		databaseRDS.Postgres_Driver = os.Getenv("DB_DRIVER")
	}

	// Get Database Secrets
	file_user, err := ioutil.ReadFile("/var/pod/secret/username")
	if err != nil {
		childLogger.Error().Err(err).Msg("Fatal erro get /var/pod/secret/username")
		os.Exit(3)
	}
	file_pass, err := ioutil.ReadFile("/var/pod/secret/password")
	if err != nil {
		childLogger.Error().Err(err).Msg("Fatal erro get /var/pod/secret/password")
		os.Exit(3)
	}
	
	databaseRDS.User = string(file_user)
	databaseRDS.Password = string(file_pass)

	return databaseRDS
}