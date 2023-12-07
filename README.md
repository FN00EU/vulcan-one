# Vulcan ONE API

Vulcan ONE API is a Go project that provides dynamic handling of API endpoints for Ethereum-related functionalities, such as ERC-20, ERC-721, and ERC-1155 token operations. The API allows users to create webhooks on EVM compatible chain on standardized contracts by specifying the desired contract standard, amount, and address in the API call.

## Getting Started

- Configure configuration.json

### Prerequisites

- [Go](https://golang.org/doc/install) installed on your machine.

### Running

```
go get all
go run main.go 
```

### Example
Standards for ERC20 compatible call with integer units are keywords "erc20" and "token", for better eading compatibility, if you want to check amount of erc20 use only whole number

standards for ERC721 compatible call with integer units are keywords "erc721" and "nft"

```
yoururl/api/evmchainfromconfiguration/erc20/amount/contractaddress
```

paste the url to custom webhook in Vulcan admin and add role



## Plan
- Add logs, improve error handling
- Add backup rpc dialers in case of main rpc unavailability
- ERC1155 friendly api call finalization
- Substrate based networks support
- Performance testing and improvements
- Improve README and add tutorials
- Add CI/CD