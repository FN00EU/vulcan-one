package erc1155

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strings"

	"github.com/FN00EU/vulcan-one/internal/utils"
	"github.com/ethereum/go-ethereum/common"
)

func GenerateCombinations(addresses []string, tokenIDs []*big.Int, erc1155TokenAmounts []*big.Int) ([]common.Address, []*big.Int, []*big.Int) {
	erc1155AddressList := make([]common.Address, 0, len(addresses)*len(tokenIDs))
	erc1155IDList := make([]*big.Int, 0, len(addresses)*len(tokenIDs))
	multipliedTokenAmounts := make([]*big.Int, 0, len(addresses)*len(tokenIDs))

	for _, addr := range addresses {
		address := common.HexToAddress(addr)
		multipliedTokenAmounts = append(multipliedTokenAmounts, erc1155TokenAmounts...)

		for _, id := range tokenIDs {
			erc1155AddressList = append(erc1155AddressList, address)
			erc1155IDList = append(erc1155IDList, id)
		}
	}

	return erc1155AddressList, erc1155IDList, multipliedTokenAmounts
}

// TODO: FIX REGEXPS TO ALLOW ONLY id&amount_id&amount or exact id-toid
// FIX ERRORS
func ReturnValidERC1155Format(str string) (string, error) {
	// Check for the exact pattern: number-number_number-number
	exactMatch, err := regexp.Compile(`^\d+_\d+(&\d+_\d+)*$`)
	if err != nil {
		return "", err
	}

	// Check for the range pattern: number-number&number-number&number-number (multiple repetitions)
	rangeMatch, err := regexp.Compile(`^\d+-\d+(&\d+-\d+)*$`)
	if err != nil {
		return "", err
	}

	if exactMatch.MatchString(str) {
		return "aformat", nil
	}

	if rangeMatch.MatchString(str) {
		return "-format", nil
	}

	return "", errors.New("invalid format")
}

func ParseERC1155(str string) ([]*big.Int, []*big.Int, error) {

	format, isValid := ReturnValidERC1155Format(str)
	if isValid != nil {
		return nil, nil, fmt.Errorf("invalid input format")
	}

	var parseIds []*big.Int
	var parseAmounts []*big.Int

	if format == "aformat" {
		parts := strings.Split(str, "&")

		parseIds = make([]*big.Int, 0, len(parts))
		parseAmounts = make([]*big.Int, 0, len(parts))

		for _, part := range parts {
			subparts := strings.Split(part, "_")

			id, ok := utils.StrToBigInt(subparts[0])
			if !ok {
				log.Println("failed to parse ID")
			}

			amount, ok := utils.StrToBigInt(subparts[1])
			if !ok {
				log.Println("failed to parse ID")
			}

			parseIds = append(parseIds, id)
			parseAmounts = append(parseAmounts, amount)
		}
	}

	if format == "-format" {
		parts := strings.Split(str, "-")

		endN, _ := utils.StrToBigInt(parts[1])
		startN, _ := utils.StrToBigInt(parts[0])

		count := new(big.Int).Sub(endN, startN)
		count = count.Add(count, big.NewInt(1))

		parseIds = make([]*big.Int, count.Int64())
		parseAmounts = make([]*big.Int, count.Int64())

		for i := 0; i < int(count.Int64()); i++ {
			parseIds[i] = new(big.Int).Add(startN, big.NewInt(int64(i)))
			parseAmounts[i] = big.NewInt(1) // Assuming a default value of 1 for any of the ERC1155 ranges
		}
	}

	return parseIds, parseAmounts, nil
}
