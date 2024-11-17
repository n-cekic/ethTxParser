package parser_rest

import (
	"bytes"
	"encoding/json"
	"ethTx/cmd/util/logging"
	"ethTx/parser"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	observedAddrs map[string]struct{}
	transactions  map[string][]parser.Transaction
}

func (ms *mockStorage) StoreAddress(address string) error {
	if _, exists := ms.observedAddrs[address]; exists {
		return fmt.Errorf("%s already subscribed", address)
	}

	ms.observedAddrs[address] = struct{}{}
	return nil
}

func (ms *mockStorage) StoreTransactions(address string, tx parser.Transaction) {
	ms.transactions[address] = append(ms.transactions[address], tx)

}

func (ms *mockStorage) Transactions(address string) []parser.Transaction {
	return ms.transactions[address]
}

func (ms *mockStorage) IsObserved(address string) bool {
	return true
}

func TestGetItemHandler(t *testing.T) {
	logging.Init("info")
	myStorage := &mockStorage{observedAddrs: map[string]struct{}{
		"":    {},
		"0x1": {},
		"0x2": {},
		"0x3": {},
		"0x4": {},
	}}

	bp := parser.NewBlockParser("", 1).WithStorage(myStorage)

	srv := Server{bp: bp}

	req := httptest.NewRequest(http.MethodGet, "/block", nil)
	rec := httptest.NewRecorder()

	srv.getBlockHandler(rec, req)
	var resp blockNumberResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp.BlockNumber != -1 {
		t.Log("Expected -1, got:", resp.BlockNumber)
		t.Fail()
	}
}

func TestSubscribeHandler(t *testing.T) {
	logging.Init("info")
	myStorage := &mockStorage{observedAddrs: map[string]struct{}{
		"":    {},
		"0x1": {},
		"0x2": {},
		"0x3": {},
		"0x4": {},
	}}

	bp := parser.NewBlockParser("", 1).WithStorage(myStorage)

	srv := Server{bp: bp}

	requestBody := map[string]string{"address": "0x1A3F"}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	// ok
	req := httptest.NewRequest(http.MethodGet, "/subscribe", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	srv.subscribeHandler(rec, req)
	var resp string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp != "Address 0x1A3F has been subscribed." {
		// t.Log("Expected -1, got:", resp.BlockNumber)
		t.Fail()
	}

	// already subscribed
	req = httptest.NewRequest(http.MethodGet, "/subscribe", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	srv.subscribeHandler(rec, req)
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp != "That didn't work" {
		// t.Log("Expected -1, got:", resp.BlockNumber)
		t.Fail()
	}

	// wrong hex format
	requestBody = map[string]string{"address": "0xG"}
	body, err = json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewReader(body))
	rec = httptest.NewRecorder()

	srv.subscribeHandler(rec, req)
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp != "That didn't work" {
		// t.Log("Expected -1, got:", resp.BlockNumber)
		t.Fail()
	}
}

func TestGetTransactionsHandler(t *testing.T) {
	logging.Init("info")
	myStorage := &mockStorage{observedAddrs: map[string]struct{}{
		"":    {},
		"0x1": {},
		"0x2": {},
		"0x3": {},
		"0x4": {},
	},
		transactions: map[string][]parser.Transaction{
			"0x1": {
				parser.Transaction{
					Hash: "asd",
					From: "me",
					To:   "you",
				},
			},
		},
	}

	bp := parser.NewBlockParser("", 1).WithStorage(myStorage)

	srv := Server{bp: bp}

	req := httptest.NewRequest(http.MethodGet, "/address/0x1", nil)
	req.SetPathValue("id", "0x1")
	rec := httptest.NewRecorder()

	srv.getTransactionsHandler(rec, req)
	var resp getTransactionsForAddressResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	want := myStorage.transactions["0x1"]
	if want[0].From != resp.Transactions[0].From ||
		want[0].BlockNumber != resp.Transactions[0].BlockNumber ||
		want[0].Hash != resp.Transactions[0].Hash ||
		want[0].To != resp.Transactions[0].To {

		t.Logf("Expected %v, got: %v", want, resp)
		t.Fail()
	}
}
