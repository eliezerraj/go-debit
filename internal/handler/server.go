package handler

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"
	"fmt"

	"github.com/gorilla/mux"

	"github.com/go-debit/internal/service"
	"github.com/go-debit/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)
//--------------------------------------------------------
type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")

	return HttpWorkerAdapter{
		workerService: workerService,
	}
}
//--------------------------------------------------------
type HttpServer struct {
	httpServer	*core.Server
}

func NewHttpAppServer(httpServer *core.Server) HttpServer {
	childLogger.Debug().Msg("NewHttpAppServer")

	return HttpServer{httpServer: httpServer }
}
//--------------------------------------------------------
func (h HttpServer) StartHttpAppServer(ctx context.Context, 
										httpWorkerAdapter *HttpWorkerAdapter,
										appServer *core.AppServer) {
	childLogger.Info().Msg("StartHttpAppServer")
		
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(appServer)
	})

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/info")
		json.NewEncoder(rw).Encode(appServer)
	})

	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpWorkerAdapter.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpWorkerAdapter.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpWorkerAdapter.Header)

	addDebit := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addDebit.Handle("/add", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "debit:", appServer.InfoPod.AvailabilityZone, ".add")), 
						http.HandlerFunc(httpWorkerAdapter.Add),
						),
	)
	addDebit.Use(httpWorkerAdapter.DecoratorDB)

	listDebit := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	listDebit.Handle("/list/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "debit:", appServer.InfoPod.AvailabilityZone, ".list")),
						http.HandlerFunc(httpWorkerAdapter.List),
						),
	)
	listDebit.Use(MiddleWareHandlerHeader)
	
	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpServer.Port)).Msg("Service Port")

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("Cancel http mux server !!!")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("WARNING Dirty Shutdown !!!")
		return
	}
	childLogger.Info().Msg("Stop Done !!!!")
}