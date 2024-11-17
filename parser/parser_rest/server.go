package parser_rest

import (
	"encoding/json"
	L "ethTx/cmd/util/logging"
	P "ethTx/parser"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	port   string
	bp     *P.BlockParser
	router *http.ServeMux
}

func Init(port, rpcURL string, parseInterval time.Duration) Server {
	L.L.Info("Initializing server...")
	bp := P.NewBlockParser(rpcURL, parseInterval)
	srv := Server{port: port, bp: bp}
	srv.registerRoutes()
	L.L.Info("Server Initialized...")
	return srv
}

func (srv *Server) Start() {
	go srv.bp.SynchronizeBlocks()
	L.L.Info("Server listening on port", srv.port)
	http.ListenAndServe(srv.port, srv.router)
}

func (srv *Server) Stop() {
	L.L.Info("Server shutting down...")
	srv.bp.StopSynchronisingBlocks()

}

func (srv *Server) registerRoutes() {
	srv.router = http.NewServeMux()

	srv.router.Handle("GET /block", http.HandlerFunc(srv.getBlockHandler))
	srv.router.Handle("POST /subscribe", http.HandlerFunc(srv.subscribeHandler))
	srv.router.Handle("GET /address/{id}", http.HandlerFunc(srv.getTransactionsHandler))
}

func (srv *Server) getBlockHandler(w http.ResponseWriter, r *http.Request) {
	resp := blockNumberResponse{srv.bp.GetCurrentBlock()}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (srv *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var req subscribeRequest
	err := d.Decode(&req)
	if err != nil {
		L.L.Error("Failed decoding request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := srv.bp.Subscribe(req.Address); !ok {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Already subscribed")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fmt.Sprintf("Address %s has been subscribed.", req.Address))
}

func (srv *Server) getTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	address := r.PathValue("id")

	transactions := srv.bp.GetTransactions(address)
	resp := getTransactionsForAddressResponse{Transactions: transactions}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
