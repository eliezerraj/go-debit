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

func (r *RestApiService) CallRestApi(	ctx context.Context, 
										method	string,
										path string, 
										x_apigw_api_id *string, 
										body interface{}) (interface{}, error) {
	childLogger.Debug().Msg("CallRestApi")

	span, ctxSpan := lib.SpanCtx(ctx, "adapter.CallRestApi:" + path)	
    defer span.End()

	childLogger.Debug().Str("method : ", method).Msg("")
	childLogger.Debug().Str("path : ", path).Msg("")
	childLogger.Debug().Interface("x_apigw_api_id : ", x_apigw_api_id).Msg("")
	childLogger.Debug().Interface("body : ", body).Msg("")

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

	req, err := http.NewRequestWithContext(ctxSpan, method, path, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	if (x_apigw_api_id != nil){
		req.Header.Add("x-apigw-api-id", *x_apigw_api_id)
	}
	req.Host = r.restApiConfig.RestEndpoint.ServerHost;

	resp, err := client.Do(req.WithContext(ctxSpan))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
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










//------------------------------------------
func (r *RestApiService) GetData(ctx context.Context, 
								urlDomain string, 
								xApigwId string, 
								data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("GetData")
	
	span := lib.Span(ctx, "adapter.GetData")	
    defer span.End()

	data_interface, err := makeGet(ctx, urlDomain, xApigwId, data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiService) PostData(	ctx context.Context, 
									urlDomain string, 
									xApigwId string, 
									data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	span := lib.Span(ctx, "adapter.PostData")	
    defer span.End()

	data_interface, err := makePost(ctx, urlDomain, xApigwId, data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func makeGet(ctx context.Context, 
			url string, 
			xApigwId string,
			data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makeGet")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	span := lib.Span(ctx, "adapter.GetData.makeGet: " + url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()

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
			return false, erro.ErrServer
	}

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}

func makePost(	ctx context.Context, 
				url string, 
				xApigwId string,
				data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("makePost")
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

	span := lib.Span(ctx, "adapter.GetData.makeGet: " + url)	
    defer span.End()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout: time.Second * 10,
	}
	
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(data)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return false, errors.New(err.Error())
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8");
	req.Header.Add("x-apigw-api-id", xApigwId);

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		childLogger.Error().Err(err).Msg("error Do Request")
		return false, errors.New(err.Error())
	}
	defer resp.Body.Close()
	
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

	result := data
	err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
		childLogger.Error().Err(err).Msg("error no ErrUnmarshal")
		return false, errors.New(err.Error())
    }

	return result, nil
}
