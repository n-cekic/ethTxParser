package main

import "ethTx/cmd/tools/logging"

func main() {
	logging.Init()
	logging.L.Info("asd")
	logging.L.Error("error")
}
