package main

import (
	"ethTx/cmd/tools/logging"
	"ethTx/parser"
	"flag"
	"time"
)

var (
	rpcURL        = flag.String("rpc.URL", "https://ethereum-rpc.publicnode.com", "Node rpc URL")
	parseInterval = flag.Duration("parse.interval", time.Second, "Interval on which to query for new block")
	logLevel      = flag.String("log.level", "info", "Logging level: `info` OR `debug`")
)

func main() {
	logging.Init(*logLevel)
	// rpcURL := "https://ethereum-rpc.publicnode.com"
	parser := parser.NewBlockParser(*rpcURL, *parseInterval)

	// Subscribe to addresses
	parser.Subscribe("0x123")
	parser.Subscribe("0x456")

	// Synchronize blocks (runs indefinitely in this example)
	parser.SynchronizeBlocks()
}
