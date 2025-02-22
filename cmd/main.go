package main

import(
	"time"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-debit/internal/infra/configuration"
	"github.com/go-debit/internal/core/model"
	"github.com/go-debit/internal/core/service"
	"github.com/go-debit/internal/infra/server"
	"github.com/go-debit/internal/adapter/api"
	"github.com/go-debit/internal/adapter/database"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
)

var(
	logLevel = 	zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv()
	apiService 	:= configuration.GetEndpointEnv() 

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ConfigOTEL = &configOTEL
	appServer.DatabaseConfig = &databaseConfig
	appServer.ApiService = apiService
}

func main (){
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx, cancel := context.WithTimeout(	context.Background(), 
										time.Duration( appServer.Server.ReadTimeout ) * time.Second)
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				log.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database, appServer.ApiService)
	httpRouters := api.NewHttpRouters(workerService)
	httpServer := server.NewHttpAppServer(appServer.Server)
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)
}