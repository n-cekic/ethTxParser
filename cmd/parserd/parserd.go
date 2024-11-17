package main

import (
	L "ethTx/cmd/util/logging"
	"ethTx/parser/parser_rest"
	"flag"
	"os"
	"os/signal"
	"time"
)

var (
	rpcURL        = flag.String("rpc.URL", "https://ethereum-rpc.publicnode.com", "Node rpc URL")
	port          = flag.String("port", ":8080", "Port to run the service on")
	parseInterval = flag.Duration("parse.interval", time.Second, "Interval on which to query for new block")
	logLevel      = flag.String("log.level", "info", "Logging level: `info` OR `debug`")
)

func main() {
	flag.Parse()

	L.Init(*logLevel)
	L.L.Warn(`This service is (due to the requrements specification) 
			meant toonly work with block numbers that can be represented as int (64bit)`)

	svc := parser_rest.Init(*port, *rpcURL, *parseInterval)
	svc.Start()

	// shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)
	<-shutdownCh
	L.L.Info("Received interrupt signal. Shutting down gracefully...")

	svc.Stop()
}
