package parser

import (
	"sync"
)

type Parser interface {
	// last parsed block
	GetCurrentBlock() int
	// add address to observer
	Subscribe(address string) bool
	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

// Transacton is minimal (shortened) structure describing single transaction
type Transaction struct {
	Hash        string `json:"hash,omitempty"`        // Transaction hash
	From        string `json:"from,omitempty"`        // Sender address
	To          string `json:"to,omitempty"`          // Recipient address
	Value       string `json:"value,omitempty"`       // Amount transferred in Wei (string for large values)
	BlockNumber int    `json:"blockNumber,omitempty"` // Block number in which the transaction was included
}

type BlockParser struct {
	currentBlock  int
	observedAddrs map[string]struct{}
	transactions  map[string][]Transaction
	mu            sync.Mutex
}

// GetCurrentBlock returns the last parsed block.
func (bp *BlockParser) GetCurrentBlock() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.currentBlock
}

// Subscribe adds an address to be observed.
func (bp *BlockParser) Subscribe(address string) bool {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	if _, exists := bp.observedAddrs[address]; exists {
		return false
	}
	bp.observedAddrs[address] = struct{}{}
	return true
}

// GetTransactions returns a list of inbound or outbound transactions for an address.
func (bp *BlockParser) GetTransactions(address string) []Transaction {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.transactions[address]
}
