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
	flag.Parse() // Parses the command-line flags provided by flag.

	L.Init(*logLevel) // Initializes logging with the specified log level.
	// Logs a warning about the service's limitation to handle only block numbers that can be represented as int (64bit).
	L.L.Warn(`This service is (due to the requrements specification) 
			meant toonly work with block numbers that can be represented as int (64bit)`)

	// Initializes the service with the provided RPC URL, port, and parse interval.
	svc := parser_rest.Init(*port, *rpcURL, *parseInterval)
	// Starts the service.
	svc.Start()

	// Creates a channel to listen for an interrupt signal (e.g., Ctrl+C).
	shutdownCh := make(chan os.Signal, 1)
	// Registers the shutdown channel to receive interrupt signals.
	signal.Notify(shutdownCh, os.Interrupt)
	// Blocks until an interrupt signal is received.
	<-shutdownCh
	L.L.Info("Received interrupt signal. Shutting down gracefully...") // Logs a message when an interrupt signal is received.

	// Stops the service gracefully.
	svc.Stop()
}
