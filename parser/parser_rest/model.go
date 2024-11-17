package parser_rest

import "ethTx/parser"

type blockNumberResponse struct {
	BlockNumber int `json:"blockNumber"`
}

type subscribeRequest struct {
	Address string `json:"address"`
}

type getTransactionsForAddressResponse struct {
	Transactions []parser.Transaction `json:"transactions"`
}
