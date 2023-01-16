package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/zhuaiballl/Go-Tokoin/config"
	"log"
	"strings"
	"time"
)

const total_voting_power = 4

var faultNumber int
var curHeight, curRound int
var step string
var lockedValue []byte
var lockedRound int
var validValue []byte
var validRound int
var tmpBlock Block
var proposalPool = make(map[string]int)
var prevotePool = make(map[string]int)
var precommitPool = make(map[string]int)
var messagePool = make(map[string]int)
var hashToBlock = make(map[string]Block)
var inSchedulePropose bool
var inSchedulePrevote bool
var inSchedulePrecommit bool

type getProposal struct {
	AddrFrom string
	Hash     []byte
}

type proposal struct {
	AddrFrom   string
	Height     int
	Round      int
	BlockHash  []byte
	ValidRound int
}

type prevote struct {
	AddrFrom    string
	Height      int
	ValidRound  int
	HashedValue []byte
}

type precommit struct {
	AddrFrom    string
	Height      int
	Round       int
	HashedValue []byte
}

func (propo *proposal) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", propo.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", propo.Round))
	lines = append(lines, fmt.Sprintf("BlockHash:  %x", propo.BlockHash))
	lines = append(lines, fmt.Sprintf("ValidRound: %d", propo.ValidRound))
	return strings.Join(lines, "\n")
}

func (propo *proposal) height_round_value() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", propo.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", propo.Round))
	lines = append(lines, fmt.Sprintf("BlockHash:  %x", propo.BlockHash))
	return strings.Join(lines, "\n")
}

func (prevo *prevote) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", prevo.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", prevo.ValidRound))
	lines = append(lines, fmt.Sprintf("HashedValue:%x", prevo.HashedValue))
	return strings.Join(lines, "\n")
}

func (prevo *prevote) height_round() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", prevo.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", prevo.ValidRound))
	return strings.Join(lines, "\n")
}

func (preco *precommit) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", preco.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", preco.Round))
	lines = append(lines, fmt.Sprintf("HashedValue:%x", preco.HashedValue))
	return strings.Join(lines, "\n")
}

func (preco *precommit) height_round() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Height:     %d", preco.Height))
	lines = append(lines, fmt.Sprintf("Round:      %d", preco.Round))
	return strings.Join(lines, "\n")
}

func proposer(height, round int) string {
	//return knownNodes[(height + round) % len(knownNodes)]
	return fmt.Sprintf("localhost:300%d", (height+round)%4)
}

func getBlock() Block {
	return tmpBlock
}

func SetBlock(block *Block) {
	tmpBlock = *block
}

func SetHeight(height int) {
	curHeight = height
}

func getBlockById(id []byte) Block {
	return hashToBlock[fmt.Sprintf("%x", id)]
}

func initTendermint() {
	faultNumber = 1
	curHeight = 0
	curRound = 0
	lockedValue = nil
	lockedRound = -1
	validValue = nil
	validRound = -1
	inSchedulePropose = false
	inSchedulePrevote = false
	inSchedulePrecommit = false
}

func broadcastPropoBlock(b *Block) {
	data := BlockPayload{config.NodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(CommandToBytes("propoBlo"), payload...)

	fmt.Println("broadcasting proposal block")

	for i := 0; i < 4; i++ {
		node := fmt.Sprintf("localhost:300%d", i)
		SendData(node, request)
	}
}

func handlePropoBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload BlockPayload

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("received proposal block from %s\n", payload.AddrFrom)

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Printf("proposal block hash: %x\n", block.Hash)

	hashToBlock[fmt.Sprintf("%x", block.Hash)] = *block

	sendGetProposal(payload.AddrFrom, block.Hash)
}

func sendGetProposal(addr string, hash []byte) {
	data := getProposal{config.NodeAddress, hash}
	payload := GobEncode(data)
	request := append(CommandToBytes("getPropo"), payload...)

	fmt.Printf("sending getProposal %x to %s\n", hash, addr)
	SendData(addr, request)
}

func handleGetProposal(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getProposal

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("received getProposal message %x from %s\n", payload.Hash, payload.AddrFrom)

	sendProposal(payload.AddrFrom, payload.Hash)
}

func sendProposal(addr string, hash []byte) {
	proposal := proposal{config.NodeAddress, curHeight, curRound, hash, validRound}
	payload := GobEncode(proposal)
	request := append(CommandToBytes("proposal"), payload...)

	fmt.Printf("sending proposal message %x to %s\n", proposal.BlockHash, addr)
	SendData(addr, request)
}

func handleProposal(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload proposal

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("received proposal message from %s, proposal value: %x\n", payload.AddrFrom, payload.BlockHash)

	// checking proposer
	if strings.Compare(payload.AddrFrom, proposer(payload.Height, payload.Round)) != 0 {
		fmt.Println("Proposal from wrong proposer!")
		return
	}
	// checking height
	if payload.Height != curHeight {
		fmt.Println("Proposal on wrong height!")
		return
	}
	// counting
	proposalPool[payload.String()] = 1
	proposalPool[payload.height_round_value()] = 1
	add_height_and_round(payload.Height, payload.Round)
	// triggering rule 1
	if payload.Height == curHeight && payload.Round == curRound && payload.ValidRound == -1 && step == "propose" {
		voteBlock := getBlockById(payload.BlockHash)
		if bc.VerifyBlock(&voteBlock) && (lockedRound == -1 || bytes.Compare(lockedValue, payload.BlockHash) == 0) {
			fmt.Println("good propo")
			broadcastPrevote(curHeight, curRound, payload.BlockHash)
		} else {
			fmt.Println("bad propo")
			broadcastPrevote(curHeight, curRound, nil)
		}
		step = "prevote"
		logStep()
	}
}

func broadcastPrevote(height, round int, hashedValue []byte) {
	prevote := prevote{config.NodeAddress, height, round, hashedValue}
	payload := GobEncode(prevote)
	request := append(CommandToBytes("prevote"), payload...)

	fmt.Println("broadcasting prevote message")

	for i := 0; i < 4; i++ {
		node := fmt.Sprintf("localhost:300%d", i)
		SendData(node, request)
	}
}

func handlePrevote(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload prevote

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// checking height
	if payload.Height != curHeight {
		fmt.Println("Prevote on wrong height!")
		return
	}
	// logging
	status := "negative"
	if payload.HashedValue != nil {
		status = "positive"
	}
	fmt.Printf("received %s prevote message from %s\n", status, payload.AddrFrom)

	payloadString := payload.String()
	// counting
	if cnt, fd := prevotePool[payloadString]; fd {
		prevotePool[payloadString] = cnt + 1
	} else {
		prevotePool[payloadString] = 1
	}
	add_height_and_round(payload.Height, payload.ValidRound)

	if prevotePool[payloadString] >= 2*faultNumber+1 {
		if payload.HashedValue != nil {
			voteBlock := getBlockById(payload.HashedValue)

			targetProposal := proposal{"", curHeight, curRound, payload.HashedValue, payload.ValidRound}
			_, fd := proposalPool[targetProposal.String()]
			fmt.Printf("finding %s in proposal pool\n", targetProposal.String())
			if fd {
				fmt.Println("proposal found")
			} else {
				fmt.Println("proposal not found")
			}

			if fd && step == "propose" && (payload.ValidRound >= 0 && payload.ValidRound < curRound) && payload.Height == curHeight {
				if bc.VerifyBlock(&voteBlock) && (lockedRound <= payload.ValidRound || bytes.Compare(lockedValue, voteBlock.Hash) == 0) {
					fmt.Println("good prevote")
					broadcastPrevote(curHeight, curRound, payload.HashedValue)
				} else {
					fmt.Println("bad prevote")
					broadcastPrevote(curHeight, curRound, nil)
				}
				step = "prevote"
				logStep()
			}

			if payload.ValidRound == curRound && payload.Height == curHeight {
				_, fd = proposalPool[targetProposal.height_round_value()]
				if fd && bc.VerifyBlock(&voteBlock) && (step == "prevote" || step == "precommit") { // TODO: this rule should be triggered only at the first time its condition is met.
					if step == "prevote" {
						lockedValue = voteBlock.Hash
						lockedRound = curRound
						broadcastPrecommit(curHeight, curRound, payload.HashedValue)
						step = "precommit"
						logStep()
					}
					validValue = voteBlock.Hash
					validRound = curRound
				}
			}
		} else {
			if step == "prevote" {
				broadcastPrecommit(curHeight, curRound, nil)
				step = "precommit"
				logStep()
			}
		}
	}

	height_round := payload.height_round()
	if cnt, fd := prevotePool[height_round]; fd {
		prevotePool[height_round] = cnt + 1
	} else {
		prevotePool[height_round] = 1
	}
	if prevotePool[height_round] >= 2*faultNumber+1 && step == "prevote" && payload.Height == curHeight {
		scheduleTimeoutPrevote()
	}
}

func broadcastPrecommit(height, round int, hashedValue []byte) {
	precommit := precommit{config.NodeAddress, height, round, hashedValue}
	payload := GobEncode(precommit)
	request := append(CommandToBytes("precommit"), payload...)

	fmt.Println("broadcasting precommit message")

	for i := 0; i < 4; i++ {
		node := fmt.Sprintf("localhost:300%d", i)
		SendData(node, request)
	}
}

func handlePrecommit(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload prevote

	buff.Write(request[config.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// checking height
	if payload.Height != curHeight {
		fmt.Println("Precommit on wrong height!")
		return
	}
	// logging
	status := "negative"

	if payload.HashedValue != nil {
		status = "positive"
	}
	fmt.Printf("received %s precommit message from %s\n", status, payload.AddrFrom)

	if payload.HashedValue != nil {
		//counting
		payloadString := payload.String()
		if cnt, fd := precommitPool[payloadString]; fd {
			precommitPool[payloadString] = cnt + 1
		} else {
			precommitPool[payloadString] = 1
		}

		targetProposal := proposal{"", curHeight, payload.ValidRound, payload.HashedValue, -1}
		_, fd := proposalPool[targetProposal.height_round_value()]
		if fd && precommitPool[payloadString] >= 2*faultNumber+1 && bc.GetBestHeight() <= curHeight && payload.Height == curHeight {
			curHeight++
			voteBlock := getBlockById(payload.HashedValue)
			if bc.VerifyBlock(&voteBlock) {
				fmt.Printf("added a new block! current height is %d, payload height is %d\n", curHeight, payload.Height)
				//curHeight++
				bc.AddBlock(&voteBlock)
				URPOSet := URPOSet{bc}
				URPOSet.Reindex()
				for _, tx := range voteBlock.Transactions {
					txID := hex.EncodeToString(tx.ID)
					DeleteMempoolTx(txID)
				}
				lockedRound = -1
				lockedValue = nil
				validRound = -1
				validValue = nil
				curRound = 0 //startRound(0)
			} else {
				curHeight--
			}
		}
	}

	add_height_and_round(payload.Height, payload.ValidRound)

	height_round := payload.height_round()
	if cnt, fd := precommitPool[height_round]; fd {
		precommitPool[height_round] = cnt + 1
	} else {
		precommitPool[height_round] = 1
	}
	if precommitPool[height_round] >= 2*faultNumber+1 && payload.Height == curHeight {
		scheduleTimeoutPrecommit()
	}
}

func StartRound(round int) {
	if round >= total_voting_power {
		return
	}
	fmt.Printf("start round %d...\n", round)
	curRound = round
	step = "propose"
	logStep()
	if strings.Compare(proposer(curHeight, curRound), config.NodeAddress) == 0 {
		var block Block
		if validValue != nil {
			block = getBlockById(validValue)
		} else {
			block = getBlock()
		}
		broadcastPropoBlock(&block)
		//broadcastProposal(curHeight, curRound, &block, validRound)
	} else {
		scheduleTimeoutPropose()
	}
}

func scheduleTimeoutPropose() {
	if !inSchedulePropose {
		inSchedulePropose = true
		timer := time.NewTimer(time.Second * 5) //timeoutPropose(curRound)
		go func() {
			<-timer.C
			onTimeoutPropose(curHeight, curRound)
		}()
	}
}

func onTimeoutPropose(height, round int) {
	if height == curHeight && round == curRound && step == "propose" {
		broadcastPrevote(curHeight, curRound, nil)
		step = "prevote"
	}
	inSchedulePropose = false
}

func scheduleTimeoutPrevote() {
	if !inSchedulePrevote {
		inSchedulePrevote = true
		timer := time.NewTimer(time.Second * 5) //timeoutPrevote(curRound)
		go func() {
			<-timer.C
			onTimeoutPrevote(curHeight, curRound)
		}()
	}
}

func onTimeoutPrevote(height, round int) {
	if height == curHeight && round == curRound && step == "prevote" {
		broadcastPrecommit(curHeight, curRound, nil)
		step = "precommit"
	}
	inSchedulePrevote = false
}

func scheduleTimeoutPrecommit() {
	if !inSchedulePrecommit {
		inSchedulePrecommit = true
		timer := time.NewTimer(time.Second * 5) //timeoutPrecommit(curRound)
		go func() {
			<-timer.C
			onTimeoutPrecommit(curHeight, curRound)
		}()
	}
}

func onTimeoutPrecommit(height, round int) {
	if height == curHeight && round == curRound {
		StartRound(curRound + 1)
	}
	inSchedulePrecommit = false
}

func add_height_and_round(height, round int) {
	heightnround := fmt.Sprintf("height:%d,round:%d", height, round)
	if cnt, fd := messagePool[heightnround]; fd {
		messagePool[heightnround] = cnt + 1
	} else {
		messagePool[heightnround] = 1
	}
	if round > curRound && messagePool[heightnround] >= faultNumber+1 {
		StartRound(round)
	}
}

func logStep() {
	fmt.Printf("step changed to %s\n", step)
}
