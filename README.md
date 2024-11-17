# ethTxParser

Ethereum transaction parser

## How to run

Locate cmd/parserd/parserd.go

``` bash
go run .
```

## Flags

| flag name      | function                                 | default value                       |
| -------------- | ---------------------------------------- | ----------------------------------- |
| rpc.URL        | Node rpc URL                             | https://ethereum-rpc.publicnode.com |
| port           | Port to run the service on               | :8080                               |
| parse.interval | Interval on which to query for new block | 1s                                  |
| log.level      | Logging level: `info` OR `debug`         | info                                |

## Rest Endpoints

### GET /block - get last parsed block

Returns json containing latest parsed block number.

Response:

``` json
{
    "blockNumber" : 12321132
}
```

### POST /subscribe - subscribe new addres

Adds new addres to the list of addresses that are being observed

Request:
``` json
{
    "blockNumber" : "0x12321132"
}
```

Response:
 - 200 : Address 0x12321132 has been subscribed.
 - 400 : "that didn't work"

### GET /address/{address} - get transactions for address

Returns the list of transactions that happened on `{address}` address.

Response:
```json
{
    [
        {
            "hash": "0x123",
            "from": "0x342",
            "to": "0x32213",
            "value": "12312",
            "blockNumber": 1231,
        }
    ]
}
```