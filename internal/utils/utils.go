package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"regexp"

	"github.com/FN00EU/vulcan-one/internal/shared"
)

func StrToBigInt(s string) (amount *big.Int, success bool) {
	amount, success = new(big.Int).SetString(s, 10)
	return
}

func LoadConfiguration(filename string) (*shared.Configuration, error) {
	filename = filepath.Clean(filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config shared.Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf(shared.ErrUnmarshalJSON, err)
	}

	return &config, nil
}

func IsValidERC1155Format(str string) bool {
	erc1155Format, err := regexp.Compile(`^\d+_\d+(&\d+_\d+)*$|^\d+-\d+(&\d+-\d+)*$`)
	if err != nil {
		log.Printf("invalid regex")
	}

	if erc1155Format.MatchString(str) {
		return true
	}

	return false
}

func CountElements(lists [][]*big.Int) []*big.Int {
	counts := make([]*big.Int, len(lists))
	for i, list := range lists {
		counts[i] = big.NewInt(int64(len(list)))
	}
	return counts
}
