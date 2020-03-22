package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listTokoins(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	txIDs := URPOSet.FindURPOIndexs(pubKeyHash)
	outs := URPOSet.FindURPO(pubKeyHash)

	for _, txID := range txIDs {
		fmt.Println(txID)
	}

	for _, out := range outs {
		out.show()
	}
}
