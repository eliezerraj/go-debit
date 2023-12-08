# go-credit

POC for test purposes.

CRUD a account_statement data.

## database

    CREATE TABLE account_statement (
        id              SERIAL PRIMARY KEY,
        fk_account_id   integer REFERENCES account(id),
        type_charge     varchar(200) NULL,
        charged_at      timestamptz NULL,
        currency        varchar(10) NULL,   
        amount          float8 NULL,
        tenant_id       varchar(200) NULL
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

+ GET /list/ACC-1

        curl svc01.domain.com/list/ACC-1 | jq

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   debit.domain.com

or

Add -host header in PostMan


## AWS

Create a public apigw