package main

import (
	"fmt"
	"github.com/atotto/clipboard"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	clipboard.WriteAll(address)
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
