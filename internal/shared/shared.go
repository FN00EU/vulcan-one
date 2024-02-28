package shared

import (
	"sync"

	"github.com/lmittmann/w3"
)

var (
	Config            *Configuration
	ClientMutex       sync.Mutex
	Clients           = make(map[string]*w3.Client)
	fpContractAddress = w3.A("0x000000000000000000000000000000000000FFFF")
)

type Configuration struct {
	EVMnetworks    map[string][]string `json:"evmNetworks"`
	Port           string              `json:"port"`
	ValidStandards []string            `json:"validStandards"`
}

const (
	errRpcUnavailable    = "RPC url %s cannot be reached"
	errIncorrectStandard = "Standard %s is not supported"
	ErrInvalidRequest    = "Invalid JSON request"
	ErrUnmarshalJSON     = "error unmarshalling configuration: %v"
	LogConnected         = "Connected to %s. Current block number: %d\n"
	addressNull          = "0x0000000000000000000000000000000000000000"
)
