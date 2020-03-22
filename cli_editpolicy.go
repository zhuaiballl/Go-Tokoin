package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func (cli *CLI) editPolicy(address, txId, nodeID, time, id, gps, temper string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Sender address is not valid")
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

	tx := EditPolicy(wallet, &URPOSet, txID, time, id, gps, temper)
	sendTx(knownNodes[0], tx)
	fmt.Println("Success!")
}
