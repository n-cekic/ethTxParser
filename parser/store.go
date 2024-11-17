package parser

import "fmt"

type Storage interface {
	StoreAddress(address string) error
	StoreTransactions(address string, tx Transaction)
	Transactions(address string) []Transaction
	IsObserved(address string) bool
}

type TransactionStorage struct {
	observedAddrs map[string]struct{}
	transactions  map[string][]Transaction
}

func (ts *TransactionStorage) StoreAddress(address string) error {
	if _, exists := ts.observedAddrs[address]; exists {
		return fmt.Errorf("%s already subscribed", address)
	}

	ts.observedAddrs[address] = struct{}{}
	return nil
}

func (ts *TransactionStorage) StoreTransactions(address string, tx Transaction) {
	ts.transactions[address] = append(ts.transactions[address], tx)
}

func (ts *TransactionStorage) Transactions(address string) []Transaction {
	return ts.transactions[address]
}

func (ts *TransactionStorage) IsObserved(address string) bool {
	_, observed := ts.observedAddrs[address]
	return observed
}
