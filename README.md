# go-debit

POC for test purposes.

CRUD a account_statement data synchronoius (REST)

Get the fee script from payfee (to calc the fees over the debit transaction)

## Diagram

go-debit (post:add/fund) == (REST) ==> go-account (service.AddFundBalanceAccount) 

go-debit (get:/script/get/{id}) == (REST) ==> go-payfee (service.GetScript)

## database

See repo https://github.com/eliezerraj/go-account-migration-worker.git

## Endpoints

+ GET /header

+ GET /info

+ POST /add

        {
            "account_id": "ACC-1",
            "type_charge": "DEBIR",
            "currency": "BRL",
            "amount": -100.00,
            "tenant_id": "TENANT-200"
        }

+ GET /list/ACC-1

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   debit.domain.com

or

Add -host header in PostMan

## AWS

Create a public apigw