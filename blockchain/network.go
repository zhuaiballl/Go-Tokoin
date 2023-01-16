package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/zhuaiballl/Go-Tokoin/config"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type AddrPayload struct {
	AddrList []string
}

type BlockPayload struct {
	AddrFrom string
	Block    []byte
}

type GetblocksPayload struct {
	AddrFrom string
}

type GetdataPayload struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type InvPayload struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type TxPayload struct {
	AddFrom     string
	Transaction []byte
}

type VersionPayload struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func CommandToBytes(command string) []byte {
	var bytes [config.CommandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func DeleteMempoolTx(txID string) {
	delete(mempool, txID)
}

func extractCommand(request []byte) []byte {
	return request[:config.CommandLength]
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := AddrPayload{knownNodes}
	nodes.AddrList = append(nodes.AddrList, config.NodeAddress)
	payload := GobEncode(nodes)
	request := append(CommandToBytes("addr"), payload...)

	SendData(address, request)
}

func sendBlock(addr string, b *Block) {
	data := BlockPayload{config.NodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(CommandToBytes("block"), payload...)

	SendData(addr, request)
}

func SendData(addr string, data []byte) {
	conn, err := net.Dial(config.Protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := InvPayload{config.NodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CommandToBytes("inv"), payload...)

	SendData(address, request)
}

func sendGetBlocks(address string) {
	payload := GobEncode(GetblocksPayload{config.NodeAddress})
	request := append(CommandToBytes("getblocks"), payload...)

	SendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetdataPayload{config.NodeAddress, kind, id})
	request := append(CommandToBytes("getdata"), payload...)

	SendData(address, request)
}

func SendTx(addr string, tnx *Transaction) {
	data := TxPayload{config.NodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(CommandToBytes("tx"), payload...)

	SendData(addr, request)
}

func HandinTx(tx *Transaction) {
	SendTx(knownNodes[0], tx)
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := GobEncode(VersionPayload{config.NodeVersion, bestHeight, config.NodeAddress})

	request := append(CommandToBytes("version"), payload...)

	SendData(addr, request)
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload AddrPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload BlockPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bchain.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		URPOSet := URPOSet{bchain}
		URPOSet.Reindex()
	}
}

func handleInv(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload InvPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload GetblocksPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bchain.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload GetdataPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bchain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}
}

func handleTx(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload TxPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if config.NodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != config.NodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	}
	if config.NodeAddress == proposer(bchain.GetBestHeight(), 0) {
		//if len(mempool) >= 1 && len(miningAddress) > 1 {
	MineTransactions:
		var txs []*Transaction

		for id := range mempool {
			tx := mempool[id]
			if bchain.VerifyTransaction(&tx) {
				txs = append(txs, &tx)
			} else {
				fmt.Println("bad transaction")
				fmt.Printf("%s\n", tx)
			}
		}

		if len(txs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")
			return
		}

		cbTx := NewCoinbaseTX(config.MiningAddress, "", 0, nil, 0, 0)
		txs = append(txs, cbTx)

		newBlock := bchain.MineBlock(txs)
		URPOSet := URPOSet{bchain}
		URPOSet.Reindex()

		fmt.Println("New block is mined!")

		for _, tx := range txs {
			txID := hex.EncodeToString(tx.ID)
			delete(mempool, txID)
		}

		//for _, node := range knownNodes {
		//	if node != nodeAddress {
		//		sendInv(node, "block", [][]byte{newBlock.Hash})
		//	}
		//}
		SetBlock(newBlock)

		if len(mempool) > 0 {
			goto MineTransactions
		}
		//}
	}
	SetHeight(bchain.GetBestHeight())
	StartRound(0)
}

func handleVersion(request []byte, bchain *Blockchain) {
	var buff bytes.Buffer
	var payload VersionPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bchain.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bchain)
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn, bchain *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:config.CommandLength])
	fmt.Println(time.Now())
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bchain)
	case "inv":
		handleInv(request, bchain)
	case "getblocks":
		handleGetBlocks(request, bchain)
	case "getdata":
		handleGetData(request, bchain)
	case "tx":
		handleTx(request, bchain)
	case "version":
		handleVersion(request, bchain)
	case "proposal":
		handleProposal(request, bchain)
	case "propoBlo":
		handlePropoBlock(request, bchain)
	case "getPropo":
		handleGetProposal(request, bchain)
	case "prevote":
		handlePrevote(request, bchain)
	case "precommit":
		handlePrecommit(request, bchain)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

// StartServer starts a node
func StartServer(nodeID, minerAddress string) {
	config.NodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	config.MiningAddress = minerAddress
	initTendermint()
	ln, err := net.Listen(config.Protocol, config.NodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockchain(nodeID)

	curHeight = bc.GetBestHeight()

	if config.NodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
