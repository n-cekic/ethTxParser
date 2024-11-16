package parser

import (
	"bytes"
	"encoding/json"
	"ethTx/cmd/tools/logging"
	"fmt"
	"net/http"
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
	rpcURL        string // URL of the Ethereum JSON-RPC endpoint
}

// NewBlockParser creates a new instance of BlockParser.
func NewBlockParser(rpcURL string) *BlockParser {
	return &BlockParser{
		currentBlock:  0,
		observedAddrs: make(map[string]struct{}),
		transactions:  make(map[string][]Transaction),
		rpcURL:        rpcURL,
	}
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

// SynchronizeBlocks starts synchronizing from the current block.
func (bp *BlockParser) SynchronizeBlocks() {
	for {
		nextBlock := bp.GetCurrentBlock() + 1
		blockData, err := bp.fetchBlock(nextBlock)
		if err != nil {
			logging.L.Error(fmt.Sprintf("Error fetching block %d: ", nextBlock), err.Error())
			break
		}

		// Process block transactions
		bp.processBlockTransactions(blockData)

		// Update the current block
		bp.mu.Lock()
		bp.currentBlock = nextBlock
		bp.mu.Unlock()
	}
}

// fetchBlock fetches block data using the eth_getBlockByNumber method.
func (bp *BlockParser) fetchBlock(blockNumber int) (map[string]interface{}, error) {
	hexBlockNumber := fmt.Sprintf("0x%x", blockNumber) // Convert block number to hex
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{hexBlockNumber, true}, // true fetches full transaction objects
		"id":      1,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(bp.rpcURL, "application/json", bytes.NewReader(requestData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		logging.L.Error("fetchBlock: failed decoding response body", err.Error())
		return nil, err
	}

	// Check for errors in the response
	if result, ok := response["result"].(map[string]interface{}); ok {
		return result, nil
	}

	return nil, fmt.Errorf("invalid response: %v", response)
}

// processBlockTransactions processes transactions in a block and stores relevant ones.
func (bp *BlockParser) processBlockTransactions(blockData map[string]interface{}) {
	transactions, ok := blockData["transactions"].([]interface{})
	if !ok {
		return
	}

	for _, tx := range transactions {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			continue
		}

		from, _ := txMap["from"].(string)
		to, _ := txMap["to"].(string)
		value, _ := txMap["value"].(string)
		blockNumber, _ := txMap["blockNumber"].(string)

		// Convert block number from hex to int
		var blockNumberInt int
		fmt.Sscanf(blockNumber, "0x%x", &blockNumberInt)

		txObj := Transaction{
			Hash:        txMap["hash"].(string),
			From:        from,
			To:          to,
			Value:       value,
			BlockNumber: blockNumberInt,
		}

		bp.mu.Lock()
		if _, isObserved := bp.observedAddrs[from]; isObserved {
			bp.transactions[from] = append(bp.transactions[from], txObj)
		}
		if _, isObserved := bp.observedAddrs[to]; isObserved {
			bp.transactions[to] = append(bp.transactions[to], txObj)
		}
		bp.mu.Unlock()
	}
}
