package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

func (cli *CLI) test(nodeID, flag, owner, holder, txId, time, id, gps, temper string) {
	bc := NewBlockchain(nodeID)
	URPOSet := URPOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}
	output := URPOSet.FindOutput(txID)

	switch flag {
	case "id_check":{//owner holder
		fmt.Print("owner check: ")
		if output.IsLockedWithKey(HashPubKey(wallets.GetWallet(owner).PublicKey)) {
			fmt.Println("pass")
		}else {
			fmt.Println("fail")
		}
		fmt.Print("holder check: ")
		if output.IsHeldWithKey(HashPubKey(wallets.GetWallet(holder).PublicKey)) {
			fmt.Println("pass")
		}else {
			fmt.Println("fail")
		}
	}
	case "ref_check":{
		cTime,_ := strconv.Atoi(time)
		cId := []byte(id)
		cGPS,_ := strconv.Atoi(gps)
		cTemper,_ := strconv.Atoi(temper)
		if output.CheckCondition(&cTime, &cId, &cGPS, &cTemper) {
			fmt.Println("pass")
		}else {
			fmt.Println("fail")
		}
	}
	case "modify_access_output":{
		tx := EditPolicy(wallets.GetWallet(owner), &URPOSet, txID, time, id, gps, temper)
		sendTx(knownNodes[0], tx)
		fmt.Println("Success!")
	}
	default:
		fmt.Println("wrong flag!")
	}
}
