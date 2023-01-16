package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/atotto/clipboard"
	bc "github.com/zhuaiballl/Go-Tokoin/blockchain"
	"github.com/zhuaiballl/Go-Tokoin/utils"
	"github.com/zhuaiballl/Go-Tokoin/wallet"
	"log"
	"strconv"
)

func (cli *CLI) test(nodeID, flag, owner, holder, txId, time, id, gps, temper string) {
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}
	output := URPOSet.FindOutput(txID)

	switch flag {
	case "id_check":
		{ //owner holder
			fmt.Print("owner check: ")
			if output.IsLockedWithKey(wallet.HashPubKey(wallets.GetWallet(owner).PublicKey)) {
				fmt.Println("pass")
			} else {
				fmt.Println("fail")
			}
			fmt.Print("holder check: ")
			if output.IsHeldWithKey(wallet.HashPubKey(wallets.GetWallet(holder).PublicKey)) {
				fmt.Println("pass")
			} else {
				fmt.Println("fail")
			}
		}
	case "ref_check":
		{
			cTime, _ := strconv.Atoi(time)
			cId := []byte(id)
			cGPS, _ := strconv.Atoi(gps)
			cTemper, _ := strconv.Atoi(temper)
			if output.CheckCondition(&cTime, &cId, &cGPS, &cTemper) {
				fmt.Println("pass")
			} else {
				fmt.Println("fail")
			}
		}
	case "modify_access_output":
		{
			tx := bc.EditPolicy(wallets.GetWallet(owner), &URPOSet, txID, time, id, gps, temper)
			bc.HandinTx(tx)
			fmt.Println("Success!")
		}
	default:
		fmt.Println("wrong flag!")
	}
}

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bchain := bc.CreateBlockchain(address, nodeID)
	defer bchain.CloseDB()

	URPOSet := bc.URPOSet{bchain}
	URPOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) createTokoin(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	//URPOSet := URPOSet{bc}
	defer bchain.CloseDB()

	//tx := NewURPOTransaction(&wallet, to, &URPOSet)

	cbTx := bc.NewCoinbaseTX(address, "", 0, nil, 0, 37)
	go bc.StartServer(nodeID, "")
	bc.HandinTx(cbTx)
	//txs := []*Transaction{cbTx}//, tx}
	//
	//newBlock := bc.MineBlock(txs)
	//bc.AddBlock(newBlock)
	//URPOSet.Update(newBlock)

	fmt.Println("Success!")
}

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	clipboard.WriteAll(address)
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) deposit(address, holder, txId string, nodeID string) { //wallet *Wallet, holder string, URPOSet *URPOSet,txId []byte, out int
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(holder) {
		log.Panic("ERROR: Holder address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(address)

	tx := bc.Deposit(&wallet, holder, &URPOSet, txId)

	bc.HandinTx(tx)

	fmt.Println("Success!")
}

func (cli *CLI) editPolicy(address, txId, nodeID, time, id, gps, temper string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Sender address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(address)

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}

	tx := bc.EditPolicy(wallet, &URPOSet, txID, time, id, gps, temper)
	bc.HandinTx(tx)
	fmt.Println("Success!")
}

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) listTokoins(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	txIDs := URPOSet.FindURPOIndexs(pubKeyHash)
	outs := URPOSet.FindURPO(pubKeyHash)

	for _, txID := range txIDs {
		fmt.Println(txID)
	}

	for _, out := range outs {
		out.Show()
	}
}

func (cli *CLI) printChain(nodeID string) {
	bchain := bc.NewBlockchain(nodeID)
	defer bchain.CloseDB()

	bci := bchain.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := bc.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) redeem(holder, owner, txId, nodeID, time, id, gps, temper string) {
	if !wallet.ValidateAddress(holder) {
		log.Panic("ERROR: Holder address is not valid")
	}
	if !wallet.ValidateAddress(owner) {
		log.Panic("ERROR: Owner address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(owner)

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}

	cTime, _ := strconv.Atoi(time)
	cId := []byte(id)
	cGPS, _ := strconv.Atoi(gps)
	cTemper, _ := strconv.Atoi(temper)
	tx := bc.RedeemTokoin(wallet, holder, &URPOSet, txID, &cTime, &cId, &cGPS, &cTemper)

	bc.HandinTx(tx)

	fmt.Println("Success!")
}

func (cli *CLI) reindexURPO(nodeID string) {
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	URPOSet.Reindex()

	count := URPOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the URPO set.\n", count)
}

func (cli *CLI) revocat(address, txId, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: The address is not valid")
	}
	bchain := bc.NewBlockchain(nodeID)
	URPOSet := bc.URPOSet{bchain}
	defer bchain.CloseDB()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(address)

	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}

	tx := bc.RevocatTokoin(wallet, &URPOSet, txID)

	bc.HandinTx(tx)

	fmt.Println("Success!")
}

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	bc.StartServer(nodeID, minerAddress)
}
