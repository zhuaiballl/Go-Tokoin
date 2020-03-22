package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createTokoin(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	//tx := NewURPOTransaction(&wallet, to, &URPOSet)

	cbTx := NewCoinbaseTX(address, "", 0, nil, 0, 37)
	txs := []*Transaction{cbTx}//, tx}

	newBlock := bc.MineBlock(txs)
	bc.AddBlock(newBlock)
	URPOSet.Update(newBlock)

	fmt.Println("Success!")
}
