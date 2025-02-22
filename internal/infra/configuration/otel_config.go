package configuration

import(
	"os"

	"github.com/joho/godotenv"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

func GetOtelEnv() go_core_observ.ConfigOTEL {
	childLogger.Debug().Msg("GetOtelEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found")
	}

	var configOTEL	go_core_observ.ConfigOTEL

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