package middleware

import (	
	"net/http"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("handler.utils", "middleware").Logger()

// Middleware v01
func MiddleWareHandlerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (INICIO)  --------------")
	
		/*if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			log.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			log.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}

		log.Debug().Str("Method : ", r.Method ).Msg("")
		log.Debug().Str("URL : ", r.URL.Path ).Msg("")*/
		//log.Println(r.Header.Get("Host"))
		//log.Println(r.Header.Get("User-Agent"))
		//log.Println(r.Header.Get("X-Forwarded-For"))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
		
		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (FIM) ----------------")

		next.ServeHTTP(w, r)
	})
}