package w3client

import (
	"log"
	"math/big"

	"github.com/FN00EU/vulcan-one/internal/shared"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func CreateClientWithPriority(rpcURLs []string) (*w3.Client, error) {
	var client *w3.Client
	var err error
	var blockNumber big.Int
	for _, rpcURL := range rpcURLs {
		client, err = w3.Dial(rpcURL)
		if err == nil {
			if err := client.Call(eth.BlockNumber().Returns(&blockNumber)); err == nil {
				log.Printf(shared.LogConnected, rpcURL, blockNumber.Int64())
				break
			} else {
				log.Printf("Error making initial call to %s: %v\n", rpcURL, err)
				client.Close()
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return client, nil
}

func SetupClients(networks map[string][]string) map[string]*w3.Client {
	clients := make(map[string]*w3.Client)
	for network, rpcURLs := range networks {
		client, err := CreateClientWithPriority(rpcURLs)
		if err != nil {
			log.Printf("Error creating client for %s: %v", network, err)
			continue
		}
		shared.ClientMutex.Lock()
		clients[network] = client
		shared.ClientMutex.Unlock()

	}
	return clients
}

func CloseClients(clients map[string]*w3.Client) {
	shared.ClientMutex.Lock()
	defer shared.ClientMutex.Unlock()

	for _, client := range clients {
		client.Close()
	}
}

func RedialClient(network string) error {
	shared.ClientMutex.Lock()
	defer shared.ClientMutex.Unlock()

	newClient, err := CreateClientWithPriority(shared.Config.EVMnetworks[network])
	if err != nil {
		return err
	}

	if existingClient, ok := shared.Clients[network]; ok {
		existingClient.Close()
	}

	shared.Clients[network] = newClient
	return nil
}
