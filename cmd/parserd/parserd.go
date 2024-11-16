package main

import (
	"ethTx/cmd/tools/logging"
	"ethTx/parser"
)

func main() {
	logging.Init()
	rpcURL := "https://ethereum-rpc.publicnode.com"
	parser := parser.NewBlockParser(rpcURL)

	// Subscribe to addresses
	parser.Subscribe("0x123")
	parser.Subscribe("0x456")

	// Synchronize blocks (runs indefinitely in this example)
	parser.SynchronizeBlocks()
}
