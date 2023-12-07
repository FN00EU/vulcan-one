package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	w3 "github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

var Config *Configuration

type Configuration struct {
	EVMnetworks    map[string]string `json:"evmNetworks"`
	Port           string            `json:"port"`
	ValidStandards []string          `json:"validStandards"`
}

type WalletRequest struct {
	Wallet  string   `json:"wallet"`
	Wallets []string `json:"wallets"`
}

type Wallets struct {
	WalletAddresses []string `json:"wallets"`
}

func validateOwnership(c *gin.Context, rpcUrl string, decimalMultiplier *big.Int, contractAddress string, wr WalletRequest, amount big.Int, contractStandard string) {
	var addresses []string
	var callRequests []w3types.Caller
	var success bool
	if wr.Wallet != "" {
		addresses = append(addresses, wr.Wallet)
	}

	if len(wr.Wallets) > 0 {
		addresses = append(addresses, wr.Wallets...)
	}

	var fetchBalances = make([]*big.Int, len(addresses))

	client := w3.MustDial(rpcUrl)
	defer client.Close()
	// todo: add ERC1155 handling later
	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")

	for i, address := range addresses {
		callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcBalanceOf, w3.A(address)).Returns(&fetchBalances[i]))
	}

	err := client.Call(callRequests...)
	if err != nil {
		fmt.Println("Error:", err)
	}

	for i, balance := range fetchBalances {
		fmt.Printf("Results for address %s: %s\n", addresses[i], balance.String())
		adjustedAmount := new(big.Int).Set(&amount)
		if contractStandard == "erc20" || contractStandard == "token" {
			adjustedAmount.Mul(adjustedAmount, decimalMultiplier)
		}
		if balance.Cmp(adjustedAmount) >= 0 {
			success = true
			break
		}
	}

	if success {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
		})
	}

}

func main() {
	configFile := "configuration.json"

	config, err := loadConfiguration(configFile)
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	Config = config

	router := gin.Default()

	// Wildcard route to capture dynamic endpoint
	router.POST("/api/:network/:standard/:amount/:contract", handleDynamicEndpoint)

	router.Run(Config.Port)
}

func loadConfiguration(filename string) (*Configuration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func strToBigInt(s string) (big.Int, bool) {
	amount, success := new(big.Int).SetString(s, 10)
	return *amount, success
}

func handleDynamicEndpoint(c *gin.Context) {
	// Extract the dynamic endpoint name from the URL
	network := c.Param("network")
	standard := c.Param("standard")
	amountStr := c.Param("amount")
	contract := c.Param("contract")

	amount, ok := strToBigInt(amountStr)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	isValidStandard := false
	for _, s := range Config.ValidStandards {
		if standard == s {
			isValidStandard = true
			break
		}
	}

	if !isValidStandard {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad API call: Invalid 'standard'"})
		return
	}

	var decimalMultiplier *big.Int
	switch network {
	case "trn":
		decimalMultiplier = new(big.Int).SetInt64(1e6)
	default:
		decimalMultiplier = new(big.Int).SetInt64(1e18)
	}
	var req WalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Wallet != "" {
		validateOwnership(c, Config.EVMnetworks[network], decimalMultiplier, contract, req, amount, standard)
	} else if len(req.Wallets) > 0 {
		validateOwnership(c, Config.EVMnetworks[network], decimalMultiplier, contract, req, amount, standard)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON structure"})
	}
}
