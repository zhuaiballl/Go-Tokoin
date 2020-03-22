package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func (cli *CLI) revocat(address, txId, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: The address is not valid")
	}
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(address)

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}

	tx := RevocatTokoin(wallet, &URPOSet, txID)

	sendTx(knownNodes[0], tx)

	fmt.Println("Success!")
}
