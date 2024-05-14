package util

import(
	"os"
	"io/ioutil"

	"github.com/joho/godotenv"
	"github.com/go-debit/internal/core"
)

func GetCertEnv() core.Cert {
	childLogger.Debug().Msg("GetCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("No .env File !!!!")
	}

	var cert		core.Cert
	var certPEM, certPrivKeyPEM	[]byte

	if os.Getenv("TLS") !=  "false" {	
		cert.IsTLS = true
	}

	if (cert.IsTLS) {
		certPEM, err = ioutil.ReadFile("/var/pod/cert/server_account_B64.crt") // server_account_B64.crt
		if err != nil {
			childLogger.Info().Err(err).Msg("Cert certPEM nao encontrado")
		} else {
			cert.CertPEM = certPEM
		}
	
		certPrivKeyPEM, err = ioutil.ReadFile("/var/pod/cert/server_account_B64.key") // server_account_B64.key
		if err != nil {
			childLogger.Info().Err(err).Msg("Cert CertPrivKeyPEM nao encontrado")
		} else {
			cert.CertPrivKeyPEM = certPrivKeyPEM
		}
	}

	return cert
}