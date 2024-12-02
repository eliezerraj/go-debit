package core

import (
	"time"

)

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
	AwsServiceConfig 	*AwsServiceConfig	`json:"aws_service_config"`
	RestApiCallData 	*RestApiCallData `json:"rest_api_call_dsa_data"`
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
	ServerHost				string `json:"server_host_localhost,omitempty"`
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

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	CreateAt		time.Time 	`json:"create_at,omitempty"`
	UpdateAt		*time.Time 	`json:"update_at,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
}

type AccountStatement struct {
	ID				int			`json:"id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	ChargeAt		time.Time 	`json:"charged_at,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}

type Fee struct {
    Name 		string  `redis:"name" json:"name"`
	Value		float64  `redis:"value" json:"value"`
}

type ScriptData struct {
    Script		Script 	`redis:"script" json:"script"`
}

type Script struct {
    Name 		string  `redis:"name" json:"name"`
    Description string   `redis:"description" json:"description"`
	Fee		    []string `redis:"fee" json:"fee"`
}

type AccountStatementFee struct {
	ID				int			`json:"id,omitempty"`
	FkAccountStatementID		 int `json:"fk_account_statement_id,omitempty"`
	TypeFee			string  	`json:"type_fee,omitempty"`
	ValueFee		float64  	`json:"value_fee,omitempty"`
	ChargeAt		time.Time 	`json:"charged_at,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}

type AwsServiceConfig struct {
	AwsRegion				string	`json:"aws_region"`
	ServiceUrlJwtSA 		string	`json:"service_url_jwt_sa"`
	SecretJwtSACredential 	string	`json:"secret_jwt_credential"`
	UsernameJwtDA			string	`json:"username_jwt_sa"`
	PasswordJwtDA			string	`json:"password_jwt_sa"`
}

type TokenSA struct {
	Token string `json:"token,omitempty"`
	Err   error
}

type RestApiCallData struct {
	Url				string `json:"url"`
	Method			string `json:"method"`
	X_Api_Id		*string `json:"x-apigw-api-id"`
	UsernameAuth	string `json:"user"`
	PasswordAuth 	string `json:"password"`
}