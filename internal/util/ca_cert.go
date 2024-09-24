package util

import(
	"os"
	"io/ioutil"
	"encoding/base64"

	"github.com/joho/godotenv"
	"github.com/go-debit/internal/core"
)

func GetCaCertEnv() core.Cert {
	childLogger.Debug().Msg("GetCaCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found !!!")
	}

	var cert		core.Cert
	var caPEM	[]byte

	if os.Getenv("TLS") !=  "false" {	
		cert.IsTLS = true
	}

	if (cert.IsTLS) {
		caPEM, err = ioutil.ReadFile("/var/pod/cert/ca.crt")
		if err != nil {
			childLogger.Info().Err(err).Msg("cert caPEM not found")
		} else {
			cert.CaPEM, err = base64.StdEncoding.DecodeString(string(caPEM))
			if err != nil {
				panic(err)
			}
		}
	}

	return cert
}