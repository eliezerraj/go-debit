#!/bin/bash

echo load DEBIT data

var_acc=0
genAcc(){
    var_acc=$(($RANDOM%($max-$min+1)+$min))
}

var_amount=0
genAmount(){
    var_amount=$(($RANDOM%($max_amount-$min_amount+1)+$min_amount * -1))
}

# --------------------Load n per 1-------------------------
domain=http://localhost:5002/add

min=499
max=510

max_amount=100
min_amount=50

for (( x=0; x<=10; x++ ))
do
    genAcc
    genAmount
    echo curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id": "ACC-'$var_acc'","type_charge": "DEBIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
    curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id": "ACC-'$var_acc'","type_charge": "DEBIT","amount":'$var_amount',"tenant_id": "TENANT-1"}'
done

