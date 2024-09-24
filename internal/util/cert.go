package util

import(
	"os"

	"github.com/joho/godotenv"
	"github.com/go-debit/internal/core"
)

func GetCertEnv() core.Cert {
	childLogger.Debug().Msg("GetCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found !!!")
	}

	var cert		core.Cert
	var certPEM, certPrivKeyPEM	[]byte

	if os.Getenv("TLS") !=  "false" {	
		cert.IsTLS = true
	}

	if (cert.IsTLS) {
		certPEM, err = os.ReadFile("/var/pod/cert/server_account_B64.crt") // server_account_B64.crt
		if err != nil {
			childLogger.Info().Err(err).Msg("cert certPEM not found")
		} else {
			cert.CertPEM = certPEM
		}
	
		certPrivKeyPEM, err = os.ReadFile("/var/pod/cert/server_account_B64.key") // server_account_B64.key
		if err != nil {
			childLogger.Info().Err(err).Msg("cert certPrivKeyPEM not found")
		} else {
			cert.CertPrivKeyPEM = certPrivKeyPEM
		}
	}

	return cert
}