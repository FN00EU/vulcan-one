package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
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

func validateOwnership(c *gin.Context, network string, contractAddress string, wr WalletRequest, amount big.Int, contractStandard string) {
	var addresses []string
	var callRequests []w3types.Caller
	var success bool
	var erc20decimals *uint8

	decimalMultiplier := new(big.Int).SetInt64(1)
	if wr.Wallet != "" {
		addresses = append(addresses, wr.Wallet)
	}

	if len(wr.Wallets) > 0 {
		addresses = append(addresses, wr.Wallets...)
	}

	rpcUrl := Config.EVMnetworks[network]

	// Support for AA operating EOAs later, right now, only FuturePass is supported on TRN.
	switch network {
	case "trn":
		addresses = addFuturePasses(addresses, rpcUrl)
	}

	var fetchBalances = make([]*big.Int, len(addresses))

	client := w3.MustDial(rpcUrl)
	defer client.Close()
	// todo: add ERC1155 handling later
	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	funcDecimals := w3.MustNewFunc("decimals()", "uint8")

	for i, address := range addresses {
		callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcBalanceOf, w3.A(address)).Returns(&fetchBalances[i]))
	}

	if contractStandard == "erc20" || contractStandard == "token" {
		callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcDecimals).Returns(&erc20decimals))
	}

	err := client.Call(callRequests...)
	if err != nil {
		fmt.Println("Error:", err)
		c.JSON(404, gin.H{"error": "Querying non-existing contract"})
	}

	for i, balance := range fetchBalances {
		fmt.Printf("Results for address %s: %s\n", addresses[i], balance.String())
		if contractStandard == "erc20" || contractStandard == "token" {
			decimalMultiplier = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(*erc20decimals)), nil)
		}
		adjustedAmount := new(big.Int).Set(&amount)
		adjustedAmount.Mul(adjustedAmount, decimalMultiplier)
		fmt.Println("adjustedamount", adjustedAmount)

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
		c.JSON(http.StatusOK, gin.H{
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

func addFuturePasses(addresses []string, rpcUrl string) []string {
	var callRequests []w3types.Caller
	fp := make([]*common.Address, len(addresses))
	client2 := w3.MustDial(rpcUrl)
	defer client2.Close()
	funcGetFuturePassOfEOA := w3.MustNewFunc("futurepassOf(address)", "address")
	var fpAddresses []string

	for i, address := range addresses {
		callRequests = append(callRequests, eth.CallFunc(w3.A("0x000000000000000000000000000000000000FFFF"), funcGetFuturePassOfEOA, w3.A(address)).Returns(&fp[i]))
	}
	err := client2.Call(callRequests...)
	if err != nil {
		fmt.Println("Error:", err)
	}

	for _, fpAddress := range fp {
		if fpAddress.Hex() != "0x0000000000000000000000000000000000000000" {
			fpAddresses = append(fpAddresses, fpAddress.Hex())
		}
	}
	addresses = append(addresses, fpAddresses...)

	return addresses
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

	var req WalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Wallet != "" {
		validateOwnership(c, network, contract, req, amount, standard)
	} else if len(req.Wallets) > 0 {
		validateOwnership(c, network, contract, req, amount, standard)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON structure"})
	}
}
