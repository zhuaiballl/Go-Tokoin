package main

import (
	"fmt"
	"log"
)

func (cli *CLI) deposit(address, holder, txId string, nodeID string) {//wallet *Wallet, holder string, URPOSet *URPOSet,txId []byte, out int
	if !ValidateAddress(address) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(holder) {
		log.Panic("ERROR: Holder address is not valid")
	}
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(address)

	tx := Deposit(&wallet, holder, &URPOSet, txId)

	sendTx(knownNodes[0], tx)

	fmt.Println("Success!")
}
