package parser

import (
	"bytes"
	"encoding/json"
	L "ethTx/cmd/tools/logging"
	"fmt"
	"net/http"
	"sync"
	"time"
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

type Block struct {
	Hash         string   `json:"Hash,omitempty"`
	ParentHash   string   `json:"ParentHash,omitempty"`
	Transactions []string `json:"Transactions,omitempty"`
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
		currentBlock:  -1,
		observedAddrs: make(map[string]struct{}),
		transactions:  make(map[string][]Transaction),
		rpcURL:        rpcURL,
		mu:            sync.Mutex{},
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

func (bp *BlockParser) SynchronizeBlocks() {
	for {
		time.Sleep(time.Second)
		//get latest block data
		latestBlockData, err := bp.getLatestBlock()
		if err != nil {
			L.L.Error("Failed fetching latest block.", "Error:", err.Error())
			continue
		} else if latestBlockData == nil {
			// no new blocks...
			continue
		}

		latestBlockParentHash, ok := latestBlockData["parentHash"].(string)
		if !ok {
			L.L.Error("Failed casting parent hash to string", fmt.Sprintf("%v", latestBlockData))
			continue
		}

		//get present block data
		currentBlock := bp.GetCurrentBlock()
		if currentBlock != -1 {
			blockData, err := bp.getBlockByNumber(currentBlock)
			if err != nil {
				L.L.Error("Failed fetching block:", fmt.Sprintf("%d", currentBlock), "Error:", err.Error())
				continue
			}

			oldBlockHash, ok := blockData["hash"].(string)
			if !ok {
				L.L.Error("Failed casting hash to string", fmt.Sprintf("%v", blockData))
				continue
			}

			// Validate chain integrity (optional)
			if latestBlockParentHash != oldBlockHash {
				// unhandled reorganization happened
				L.L.Error("Unhandled block reorg happened. Shutting down...")
				return
			}
		}

		// Process block transactions
		bp.processBlockTransactions(latestBlockData)

		// Update the current block
		latestBlockNoHex, ok := latestBlockData["number"].(string)
		if !ok {
			L.L.Error("Failed casting latest block number to string", fmt.Sprintf("%v", latestBlockNoHex))
			continue
		}
		var latestBlockNo int
		fmt.Sscanf(latestBlockNoHex, "0x%x", &latestBlockNo)
		bp.mu.Lock()
		bp.currentBlock = latestBlockNo
		bp.mu.Unlock()
	}
}

// getLatestBlock returns latest block data by calling getBlockByNumber(getBlockNumber()).
//
// If blockNumber == currentBlocknumber the function returns (nil, nil)
func (bp *BlockParser) getLatestBlock() (map[string]interface{}, error) {
	blockNumber, err := bp.getBlockNumber()
	if err != nil {
		return nil, err
	}

	if blockNumber == bp.GetCurrentBlock() {
		L.L.Info("No new blocks...")
		return nil, nil
	}
	L.L.Info("New block", fmt.Sprintf("%d", blockNumber))

	blockData, err := bp.getBlockByNumber(blockNumber)
	if err != nil {
		return nil, err
	}

	return blockData, nil
}

// getBlockNumber returns latest block number
func (bp *BlockParser) getBlockNumber() (int, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(bp.rpcURL, "application/json", bytes.NewReader(requestData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		L.L.Error("fetchBlock: failed decoding response body", err.Error())
		return 0, err
	}

	blockHexNumber, ok := response["result"].(string)
	if !ok {
		return 0, fmt.Errorf("failed casting the 'result' from response: %v", response)
	}

	var blockNo int
	fmt.Sscanf(blockHexNumber, "0x%x", &blockNo)
	return blockNo, nil
}

// fetchBlock fetches block data using the eth_getBlockByNumber method.
func (bp *BlockParser) getBlockByNumber(blockNumber int) (map[string]interface{}, error) {
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
		L.L.Error("fetchBlock: failed decoding response body", err.Error())
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
