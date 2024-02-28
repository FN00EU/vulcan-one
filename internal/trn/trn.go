package trn

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
)

var (
	fpContractAddress = w3.A("0x000000000000000000000000000000000000FFFF")
	addressNull       = "0x0000000000000000000000000000000000000000"
)

func AddFuturePasses(addresses []string, client w3.Client) []string {
	var callRequests []w3types.Caller
	fp := make([]*common.Address, len(addresses))
	funcGetFuturePassOfEOA := w3.MustNewFunc("futurepassOf(address)", "address")
	var fpAddresses []string

	for i, address := range addresses {
		callRequests = append(callRequests, eth.CallFunc(fpContractAddress, funcGetFuturePassOfEOA, w3.A(address)).Returns(&fp[i]))
	}
	err := client.Call(callRequests...)
	if err != nil {
		log.Println("Error:", err)
	}

	for i, fpAddress := range fp {
		log.Println(i)
		if fpAddress.Hex() != addressNull {
			fpAddresses = append(fpAddresses, fpAddress.Hex())
		}
	}

	addresses = append(addresses, fpAddresses...)

	return addresses
}

func AssetIdToERC20Address(assetId string) string {
	assetIdInHex := new(big.Int)
	assetIdInHex.SetString(assetId, 10)
	assetIdHex := fmt.Sprintf("%08s", assetIdInHex.Text(16))
	return common.HexToAddress(fmt.Sprintf("0xCCCCCCCC%s000000000000000000000000", strings.ToUpper(assetIdHex))).Hex()
}

func CollectionIdToERC721Address(collectionId string) string {
	collectionIdInHex := new(big.Int)
	collectionIdInHex.SetString(collectionId, 10)
	collectionIdHex := fmt.Sprintf("%08s", collectionIdInHex.Text(16))
	return common.HexToAddress(fmt.Sprintf("0xAAAAAAAA%s000000000000000000000000", strings.ToUpper(collectionIdHex))).Hex()
}

func CollectionIdToERC1155Address(collectionId string) string {
	collectionIdInHex := new(big.Int)
	collectionIdInHex.SetString(collectionId, 10)
	collectionIdHex := fmt.Sprintf("%08s", collectionIdInHex.Text(16))
	return common.HexToAddress(fmt.Sprintf("0xBBBBBBBB%s000000000000000000000000", strings.ToUpper(collectionIdHex))).Hex()
}
