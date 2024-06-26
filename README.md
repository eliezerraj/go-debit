# go-debit

POC for test purposes.

CRUD a account_statement data synchronoius (REST)

Get the fee script from payfee (to calc the fees over the debit transaction)

## Diagram

go-debit (post:add/fund) == (REST) ==> go-account (service.AddFundBalanceAccount) 

go-debit (get:/script/get/{id}) == (REST) ==> go-payfee (service.GetScript)

## database

    CREATE TABLE public.account_statement (
        id serial4 NOT NULL,
        fk_account_id int4 NULL,
        type_charge varchar(200) NULL,
        charged_at timestamptz NULL,
        currency varchar(10) NULL,
        amount float8 NULL,
        tenant_id varchar(200) NULL,
        CONSTRAINT account_statement_pkey PRIMARY KEY (id)
    );

    CREATE TABLE public.account_statement_fee (
        id serial4 NOT NULL,
        fk_account_statement_id int4 NULL,
        charged_at timestamptz NULL,
        type_fee varchar(200) NULL,
        value_fee float8 NULL,
        currency varchar(10) NULL,
        amount float8 NULL,
        tenant_id varchar(200) NULL,
        CONSTRAINT account_statement_fee_pkey PRIMARY KEY (id)
    );


## Endpoints

+ POST /add

        {
            "account_id": "ACC-1",
            "type_charge": "DEBIR",
            "currency": "BRL",
            "amount": -100.00,
            "tenant_id": "TENANT-200"
        }

+ GET /header

+ GET /info

+ GET /list/ACC-1

        curl svc01.domain.com/list/ACC-1 | jq

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   debit.domain.com

or

Add -host header in PostMan


## AWS

Create a public apigw