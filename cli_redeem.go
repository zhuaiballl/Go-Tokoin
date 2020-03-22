package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

func (cli *CLI) redeem(holder, owner, txId, nodeID, time, id, gps, temper string) {
	if !ValidateAddress(holder) {
		log.Panic("ERROR: Holder address is not valid")
	}
	if !ValidateAddress(owner) {
		log.Panic("ERROR: Owner address is not valid")
	}
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(owner)

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}

	cTime,_ := strconv.Atoi(time)
	cId := []byte(id)
	cGPS,_ := strconv.Atoi(gps)
	cTemper,_ := strconv.Atoi(temper)
	tx := RedeemTokoin(wallet, holder, &URPOSet, txID, &cTime, &cId, &cGPS, &cTemper)

	sendTx(knownNodes[0], tx)

	fmt.Println("Success!")
}
