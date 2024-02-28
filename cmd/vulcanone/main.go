package main

import (
	"sync"

	"github.com/FN00EU/vulcan-one/internal/api"
	"github.com/FN00EU/vulcan-one/internal/shared"

	w3 "github.com/lmittmann/w3"
)

// TODO: add logging
// TODO: error 404 for non-existing contract
var (
	Config            *shared.Configuration
	clientMutex       sync.Mutex
	clients           = make(map[string]*w3.Client)
	fpContractAddress = w3.A("0x000000000000000000000000000000000000FFFF")
)

type WalletRequest struct {
	Wallet  string   `json:"wallet"`
	Wallets []string `json:"wallets"`
}

type Wallets struct {
	WalletAddresses []string `json:"wallets"`
}

// TODO: finalize all errors as constants for easier logging
const (
	errRpcUnavailable    = "RPC url %s cannot be reached"
	errIncorrectStandard = "Standard %s is not supported"
	errInvalidRequest    = "Invalid JSON request"
	errUnmarshalJSON     = "error unmarshalling configuration: %v"
	logConnected         = "Connected to %s. Current block number: %d\n"
	addressNull          = "0x0000000000000000000000000000000000000000"
)

// TODO: better handling, add defer
func main() {
	api.Start()
}
