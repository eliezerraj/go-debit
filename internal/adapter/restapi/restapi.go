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
	"github.com/aws/aws-xray-sdk-go/xray"
)

var childLogger = log.With().Str("adapter/restapi", "restApiService").Logger()

type RestApiService struct {

}

func NewRestApiService() (*RestApiService){
	childLogger.Debug().Msg("*** NewRestApiService")
	
	return &RestApiService {
	}
}

func (r *RestApiService) GetData(ctx context.Context, 
								serverUrlDomain string, 
								xApigwId string, 
								path string, 
								id string,) (interface{}, error) {
	childLogger.Debug().Msg("GetData")

	domain := serverUrlDomain + path +"/" + id

	data_interface, err := makeGet(ctx, domain, xApigwId, id)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func (r *RestApiService) PostData(ctx context.Context, 
									serverUrlDomain string, 
									xApigwId string, 
									path string,
									data interface{}) (interface{}, error) {
	childLogger.Debug().Msg("PostData")

	domain := serverUrlDomain + path 

	data_interface, err := makePost(ctx, domain, xApigwId, data)
	if err != nil {
		childLogger.Error().Err(err).Msg("error Request")
		return nil, errors.New(err.Error())
	}
    
	return data_interface, nil
}

func makeGet(ctx context.Context, 
			url string, 
			xApigwId string,
			id interface{}) (interface{}, error) {

	childLogger.Debug().Msg("makeGet")
	client := xray.Client(&http.Client{Timeout: time.Second * 5})
	
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

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

	result := id
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
	client := xray.Client(&http.Client{Timeout: time.Second * 5})
	
	childLogger.Debug().Str("url : ", url).Msg("")
	childLogger.Debug().Str("xApigwId : ", xApigwId).Msg("")

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
