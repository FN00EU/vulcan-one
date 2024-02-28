package api

import (
	"log"
	"math/big"
	"net/http"

	"github.com/FN00EU/vulcan-one/internal/erc1155"
	"github.com/FN00EU/vulcan-one/internal/shared"
	"github.com/FN00EU/vulcan-one/internal/trn"
	"github.com/FN00EU/vulcan-one/internal/utils"
	"github.com/FN00EU/vulcan-one/internal/w3client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

type WalletRequest struct {
	Wallet  string   `json:"wallet"`
	Wallets []string `json:"wallets"`
}

type Wallets struct {
	WalletAddresses []string `json:"wallets"`
}

// TODO: Add command line flag for config
func Start() {

	configFile := "./configs/configuration.json"

	config, err := utils.LoadConfiguration(configFile)
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	shared.Config = config

	shared.Clients = w3client.SetupClients(shared.Config.EVMnetworks)
	defer w3client.CloseClients(shared.Clients)

	router := gin.Default()

	router.POST("/api/:network/:standard/:amount/:contract", func(c *gin.Context) {
		handleDynamicEndpoint(c, shared.Clients)
	})

	router.Run(shared.Config.Port)
}

func handleDynamicEndpoint(c *gin.Context, clients map[string]*w3.Client) {
	network := c.Param("network")
	standard := c.Param("standard")
	amountStr := c.Param("amount")
	contract := c.Param("contract")

	shared.ClientMutex.Lock()
	client, exists := clients[network]
	shared.ClientMutex.Unlock()
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid network"})
		return
	}

	amount, ok := utils.StrToBigInt(amountStr)
	log.Println(&amount)
	if !ok {
		erc1155Format := utils.IsValidERC1155Format(amountStr)
		if !erc1155Format {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}
	}

	isValidStandard := false
	for _, s := range shared.Config.ValidStandards {
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

	if req.Wallet != "" || len(req.Wallets) > 0 {
		validateOwnership(c, network, client, contract, req, amountStr, standard)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": shared.ErrInvalidRequest})
	}
}

func validateOwnership(c *gin.Context, network string, client *w3.Client, contractAddress string, wr WalletRequest, amount string, contractStandard string) {
	var addresses []string
	var callRequests []w3types.Caller
	var success bool
	var erc20decimals *uint8
	var fetchBalances []*big.Int
	var erc1155TokenAmounts []*big.Int
	var amountBigInt *big.Int
	var err error
	decimalMultiplier := new(big.Int).SetInt64(1)
	if wr.Wallet != "" {
		addresses = append(addresses, wr.Wallet)
	}
	if len(wr.Wallets) > 0 {
		addresses = append(addresses, wr.Wallets...)
	}

	// Support for AA operating EOAs later, right now, only FuturePass is supported on TRN.
	switch network {
	case "trn", "porcini":
		addresses = trn.AddFuturePasses(addresses, *client)
	}

	funcBalanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	funcDecimals := w3.MustNewFunc("decimals()", "uint8")

	switch contractStandard {
	case "erc20", "token":
		fetchBalances = make([]*big.Int, len(addresses))
		callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcDecimals).Returns(&erc20decimals))
		for i, address := range addresses {
			callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcBalanceOf, w3.A(address)).Returns(&fetchBalances[i]))
		}

	case "nft", "erc721":
		fetchBalances = make([]*big.Int, len(addresses))
		for i, address := range addresses {
			callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcBalanceOf, w3.A(address)).Returns(&fetchBalances[i]))
		}
	case "sft", "erc1155":
		var erc1155TokenIds []*big.Int
		var erc1155AddressList []common.Address
		var erc1155IDList []*big.Int
		funcBalanceOfBatchSFT := w3.MustNewFunc("balanceOfBatch(address[],uint256[])", "uint256[]")
		erc1155TokenIds, erc1155TokenAmounts, err = erc1155.ParseERC1155(amount)
		erc1155AddressList, erc1155IDList, erc1155TokenAmounts = erc1155.GenerateCombinations(addresses, erc1155TokenIds, erc1155TokenAmounts)

		if err != nil {
			log.Println(err)
		}
		fetchBalances = make([]*big.Int, len(addresses)*len(erc1155TokenIds))
		callRequests = append(callRequests, eth.CallFunc(w3.A(contractAddress), funcBalanceOfBatchSFT, erc1155AddressList, erc1155IDList).Returns(&fetchBalances))
	}

	err = client.Call(callRequests...)
	if err != nil {
		if callErr, ok := err.(w3.CallErrors); ok {
			log.Println("w3 error:", callErr)
			log.Println("Other Error:", err)
			c.JSON(500, gin.H{"w3 error": err.(w3.CallErrors)})

		} else {
			log.Println("Other Error:", err)
			c.JSON(500, gin.H{"error": err})
		}
	}

	for i, balance := range fetchBalances {

		switch contractStandard {
		case "erc20", "token":
			decimalMultiplier = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(*erc20decimals)), nil)
			amountBigInt, _ = utils.StrToBigInt(amount)

		case "erc1155", "sft":
			amountBigInt = new(big.Int).Set(erc1155TokenAmounts[i])

		default:
			amountBigInt, _ = utils.StrToBigInt(amount)
		}

		adjustedAmount := new(big.Int).Set(amountBigInt)
		adjustedAmount.Mul(adjustedAmount, decimalMultiplier)

		if balance.Cmp(adjustedAmount) >= 0 {
			success = true
			log.Printf("balance met in element - %d", i)
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
