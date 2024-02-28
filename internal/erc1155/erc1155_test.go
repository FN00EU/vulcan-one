package erc1155_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/FN00EU/vulcan-one/internal/erc1155"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCombinations(t *testing.T) {
	// Test case 1: Empty input lists
	addresses := []string{}
	tokenIDs := []*big.Int{}
	erc1155TokenAmounts := []*big.Int{}

	erc1155AddressList, erc1155IDList, multipliedTokenAmounts := erc1155.GenerateCombinations(addresses, tokenIDs, erc1155TokenAmounts)
	assert.Empty(t, erc1155AddressList, "Should be empty")
	assert.Empty(t, erc1155IDList, "Should be empty")
	assert.Empty(t, multipliedTokenAmounts, "Should be empty")

	// Test case 2: Non-empty input lists
	addresses = []string{"0x123", "0x456"}
	tokenIDs = []*big.Int{big.NewInt(1), big.NewInt(2)}
	erc1155TokenAmounts = []*big.Int{big.NewInt(10), big.NewInt(20)}

	erc1155AddressList, erc1155IDList, multipliedTokenAmounts = erc1155.GenerateCombinations(addresses, tokenIDs, erc1155TokenAmounts)
	expectedAddresses := []common.Address{
		common.HexToAddress("0x123"),
		common.HexToAddress("0x123"),
		common.HexToAddress("0x456"),
		common.HexToAddress("0x456"),
	}
	expectedTokenIDs := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(1), big.NewInt(2)}
	expectedTokenAmounts := []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(10), big.NewInt(20)}

	assert.Equal(t, expectedAddresses, erc1155AddressList, "Should be equal")
	assert.Equal(t, expectedTokenIDs, erc1155IDList, "Should be equal")
	assert.Equal(t, expectedTokenAmounts, multipliedTokenAmounts, "Should be equal")
}

func TestReturnValidERC1155Format(t *testing.T) {
	// Test case 1: Valid exact match
	str := "44_1&44_1"
	result, err := erc1155.ReturnValidERC1155Format(str)
	assert.NoError(t, err, "Should not return an error")
	assert.Equal(t, "aformat", result, "Should return 'aformat'")

	// Test case 2: Valid range match
	str = "1-10"
	result, err = erc1155.ReturnValidERC1155Format(str)
	assert.NoError(t, err, "Should not return an error")
	assert.Equal(t, "-format", result, "Should return '-format'")

	// Test case 3: Invalid range match

	// Test case 4: Invalid input
	str = "invalid_format"
	result, err = erc1155.ReturnValidERC1155Format(str)
	assert.Error(t, err, "Should return an error")
	assert.Empty(t, result, "Should return an empty string")

	// Test case 5: Error in regex compilation
	invalidRegexStr := "invalid[regex"
	result, err = erc1155.ReturnValidERC1155Format(invalidRegexStr)
	assert.Error(t, err, "Should return an error")
	assert.Empty(t, result, "Should return an empty string")
}

func TestParseERC1155(t *testing.T) {
	tests := []struct {
		input           string
		expectedIds     []*big.Int
		expectedAmounts []*big.Int
		expectError     bool
	}{
		// Test case for "aformat"
		{
			input:           "123_1&456_2&789_3",
			expectedIds:     []*big.Int{big.NewInt(123), big.NewInt(456), big.NewInt(789)},
			expectedAmounts: []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			expectError:     false,
		},
		// Test case for "-format"
		{
			input:           "100-105",
			expectedIds:     []*big.Int{big.NewInt(100), big.NewInt(101), big.NewInt(102), big.NewInt(103), big.NewInt(104), big.NewInt(105)},
			expectedAmounts: []*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)},
			expectError:     false,
		},
		// Add more test cases as needed
	}

	for _, test := range tests {
		actualIds, actualAmounts, err := erc1155.ParseERC1155(test.input)

		if (err != nil) != test.expectError {
			t.Errorf("Test case %q failed: expected error %v, but got %v", test.input, test.expectError, err)
			continue
		}

		if !reflect.DeepEqual(actualIds, test.expectedIds) {
			t.Errorf("Test case %q failed: expected ids %v, but got %v", test.input, test.expectedIds, actualIds)
		}

		if !reflect.DeepEqual(actualAmounts, test.expectedAmounts) {
			t.Errorf("Test case %q failed: expected amounts %v, but got %v", test.input, test.expectedAmounts, actualAmounts)
		}
	}
}
