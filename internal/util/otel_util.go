package util

import(
	"os"

	"github.com/joho/godotenv"
	"github.com/go-debit/internal/core"
)

func GetOtelEnv() core.ConfigOTEL {
	childLogger.Debug().Msg("GetCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("No .env File !!!!")
	}

	var configOTEL	core.ConfigOTEL

	configOTEL.TimeInterval = 1
	configOTEL.TimeAliveIncrementer = 1
	configOTEL.TotalHeapSizeUpperBound = 100
	configOTEL.ThreadsActiveUpperBound = 10
	configOTEL.CpuUsageUpperBound = 100
	configOTEL.SampleAppPorts = []string{}

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") !=  "" {	
		configOTEL.OtelExportEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	return configOTEL
}