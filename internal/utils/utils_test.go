package utils_test

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/FN00EU/vulcan-one/internal/utils"
	"github.com/stretchr/testify/assert"
)

func MockFile(content string) (string, func()) {
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temporary file: %v", err))
	}

	if _, err := tmpfile.WriteString(content); err != nil {
		panic(fmt.Sprintf("Failed to write to temporary file: %v", err))
	}

	return tmpfile.Name(), func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}
}

func TestStrToBigInt(t *testing.T) {
	// Test case 1: Valid input
	s := "123456789"
	amount, success := utils.StrToBigInt(s)
	expectedAmount := big.NewInt(123456789)

	assert.True(t, success, "Conversion should be successful")
	assert.Equal(t, expectedAmount, amount, "Should be equal")

	// Test case 2: Invalid input
	s = "not_a_number"
	amount, success = utils.StrToBigInt(s)

	assert.False(t, success, "Conversion should not be successful")
	assert.Nil(t, amount, "Result should be nil for unsuccessful conversion")
}

func TestLoadConfiguration(t *testing.T) {
	// Test case 1: Valid JSON file
	jsonContent := `{
		"evmNetworks": {
			"eth": ["wss://ethereum.publicnode.com", "https://eth.llamarpc.com"],
			"trn": ["https://root.rootnet.live/archive"],
			"arb": ["https://arb1.arbitrum.io/rpc"],
			"frame": ["https://rpc.testnet.frame.xyz/http"]
		},
		"validStandards": ["erc20", "token", "erc721", "nft", "sft", "erc1155"],
		"port": ":8080"
	}`

	filename, cleanup := MockFile(jsonContent)
	defer cleanup()

	config, err := utils.LoadConfiguration(filename)
	assert.NoError(t, err, "Should not return an error")
	assert.NotNil(t, config, "Configuration should not be nil")
	assert.Equal(t, []string{"erc20", "token", "erc721", "nft", "sft", "erc1155"}, config.ValidStandards, "ValidStandards should match")
	assert.Equal(t, ":8080", config.Port, "Port should match")
	assert.Len(t, config.EVMnetworks, 4, "EVMnetworks should have 4 items")

	// Test case 2: Invalid JSON file
	invalidJSONContent := `invalid json`
	invalidFilename, cleanupInvalid := MockFile(invalidJSONContent)
	defer cleanupInvalid()

	invalidConfig, err := utils.LoadConfiguration(invalidFilename)
	assert.Error(t, err, "Should return an error")
	assert.Nil(t, invalidConfig, "Invalid configuration should be nil")
}

func TestCountElements(t *testing.T) {
	// Test case 1: Empty lists
	emptyLists := [][]*big.Int{}
	emptyCounts := utils.CountElements(emptyLists)
	assert.NotNil(t, emptyCounts, "Counts should not be nil")
	assert.Empty(t, emptyCounts, "Counts should be empty")

	// Test case 2: Non-empty lists
	lists := [][]*big.Int{
		[]*big.Int{big.NewInt(955), big.NewInt(921), big.NewInt(921)},
		[]*big.Int{big.NewInt(900)},
	}
	counts := utils.CountElements(lists)
	assert.NotNil(t, counts, "Counts should not be nil")
	assert.Len(t, counts, 2, "Counts should have length 2")
	assert.Equal(t, big.NewInt(3), counts[0], "First count should be 3")
	assert.Equal(t, big.NewInt(1), counts[1], "Second count should be 1")
}
