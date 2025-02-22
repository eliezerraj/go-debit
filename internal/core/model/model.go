package model

import (
	"time"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	DatabaseConfig	*go_core_pg.DatabaseConfig  `json:"database"`
	ApiService 		[]ApiService 			
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
	Obs				string  	`json:"obs,omitempty"`
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

type ApiService struct {
	Name			string `json:"name_service"`
	Url				string `json:"url"`
	Method			string `json:"method"`
	Header_x_apigw_api_id	string `json:"x-apigw-api-id"`
}