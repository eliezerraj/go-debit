package core

import (
	"time"

)

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
