package core

type DatabaseRDS struct {
    Host 				string `json:"host"`
    Port  				string `json:"port"`
	Schema				string `json:"schema"`
	DatabaseName		string `json:"databaseName"`
	User				string `json:"user"`
	Password			string `json:"password"`
	Db_timeout			int	`json:"db_timeout"`
	Postgres_Driver		string `json:"postgres_driver"`
}

type AppServer struct {
	InfoPod 		*InfoPod 		`json:"info_pod"`
	Server     		*Server     	`json:"server"`
	Database		*DatabaseRDS	`json:"database"`
	Cert			*Cert			`json:"cert"`
	RestEndpoint	*RestEndpoint	`json:"rest_endpoint"`
	ConfigOTEL		*ConfigOTEL		`json:"otel_config"`
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
	Cert			*Cert `json:"server_cert"`	
}

type RestEndpoint struct {
	ServiceUrlDomain 		string `json:"service_url_domain"`
	XApigwId				string `json:"xApigwId"`
	CaCert					*Cert `json:"ca_cert"`
	ServiceUrlDomainPayFee	string `json:"service_url_domain_pay_fee"`
	XApigwIdPayFee			string `json:"xApigwId_pay_fee"`
}

type Cert struct {
	IsTLS				bool
	CaPEM 				[]byte 	`json:"ca_cert"`	
	CertPEM 			[]byte 	`json:"server_cert"`		
	CertPrivKeyPEM	    []byte  `json:"server_key"`	   
}

type ConfigOTEL struct {
	OtelExportEndpoint		string
	TimeInterval            int64    `mapstructure:"TimeInterval"`
	TimeAliveIncrementer    int64    `mapstructure:"RandomTimeAliveIncrementer"`
	TotalHeapSizeUpperBound int64    `mapstructure:"RandomTotalHeapSizeUpperBound"`
	ThreadsActiveUpperBound int64    `mapstructure:"RandomThreadsActiveUpperBound"`
	CpuUsageUpperBound      int64    `mapstructure:"RandomCpuUsageUpperBound"`
	SampleAppPorts          []string `mapstructure:"SampleAppPorts"`
}