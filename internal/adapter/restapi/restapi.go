package restapi

import(
	"errors"
	"net/http"
	"time"
	"encoding/json"
	"bytes"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/go-debit/internal/erro"
	"github.com/go-debit/internal/lib"
	"github.com/go-debit/internal/core"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var childLogger = log.With().Str("adapter/restapi", "restApiService").Logger()
//-------------------------------------
type RestApiService struct {
	restApiConfig *core.AppServer
}

func NewRestApiService(restApiConfig *core.AppServer) (*RestApiService){
	childLogger.Debug().Msg("*** NewRestApiService")
	
	return &RestApiService {
		restApiConfig: restApiConfig,
	}
}

func (r *RestApiService) CallApiRest(ctx context.Context,
									restApiCallData	core.RestApiCallData,
									body interface{}) (interface{}, error){

	childLogger.Debug().Msg("CallApiRest")
	childLogger.Debug().Msg("--------------------------")
	childLogger.Debug().Interface("CallApiRest : ", restApiCallData).Msg("")
	childLogger.Debug().Msg("--------------------------")

	span, ctxSpan := lib.SpanCtx(ctx, "adapter.CallApiRest:" + restApiCallData.Url)	
    defer span.End()

	transportHttp := &http.Transport{}

	/*transport := &http.Transport{
		TLSClientConfig: r.ClientTLSConf,
	}
	client := &http.Client{Timeout: time.Second * 5 , Transport: transport}*/

	client := http.Client{
		Transport: otelhttp.NewTransport(transportHttp),
		Timeout: time.Second * 29,
	}

	payload := new(bytes.Buffer) 
	if body != nil{
		json.NewEncoder(payload).Encode(body)
	}

	req, err := http.NewRequestWithContext(ctxSpan, restApiCallData.Method, restApiCallData.Url, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error NewRequestWithContext")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	if (restApiCallData.X_Api_Id != nil){
		req.Header.Add("x-apigw-api-id", *restApiCallData.X_Api_Id )
	}

	req.Host = r.restApiConfig.RestEndpoint.ServerHost;

	resp, err := client.Do(req.WithContext(ctxSpan))
	if err != nil {
		childLogger.Error().Err(err).Msg("error client.Do")
		return false, errors.New(err.Error())
	}

	childLogger.Debug().Int("StatusCode :", resp.StatusCode).Msg("")
	
	switch (resp.StatusCode) {
		case 401:
			return false, erro.ErrHTTPForbiden
		case 403:
			return false, erro.ErrHTTPForbiden
		case 200:
		case 400:
			return false, erro.ErrNotFound
		case 404:
			return false, erro.ErrNotFound
		default:
			return false, erro.ErrHTTPForbiden
	}

	result := body
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}