package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/zhuaiballl/Go-Tokoin/utils"
	wlt "github.com/zhuaiballl/Go-Tokoin/wallet"
	"math/big"
	"strconv"
	"strings"

	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Time:         %d", output.Time))
		lines = append(lines, fmt.Sprintf("       ID:           %s", output.ID))
		lines = append(lines, fmt.Sprintf("       GPS:          %d", output.GPS))
		lines = append(lines, fmt.Sprintf("       Temperature:  %d", output.Temperature))
		lines = append(lines, fmt.Sprintf("       OwnerKey:     %x", output.PubKeyHash))
		lines = append(lines, fmt.Sprintf("       HolderKey:    %x", output.HolderKey))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Time, vout.ID, vout.GPS, vout.Temperature, vout.PubKeyHash, vout.HolderKey})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(addr, data string, time int, id []byte, gps int, temper int) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(time, id, gps, temper, addr)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// NewURPOTransaction creates a new transaction
func NewURPOTransaction(wallet *wlt.Wallet, to string, URPOSet *URPOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := wlt.HashPubKey(wallet.PublicKey)
	validOutputs := URPOSet.FindSpendableOutputs(pubKeyHash)

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *NewTXOutput(0, nil, 0, 0, to))

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	URPOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

// Deposit sets a holder for a tokoin
func Deposit(wallet *wlt.Wallet, holder string, URPOSet *URPOSet, txId string) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := wlt.HashPubKey(wallet.PublicKey)
	txID, err := hex.DecodeString(txId)
	if err != nil {
		log.Panic(err)
	}
	output := URPOSet.FindOutput(txID)
	if !(output.IsLockedWithKey(pubKeyHash)) {
		log.Panic("ERROR: Not locked with this key")
	}

	input := TXInput{txID, 0, nil, wallet.PublicKey}
	inputs = append(inputs, input)

	newOutput := output
	newOutput.Hold([]byte(holder))
	outputs = append(outputs, newOutput)

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	URPOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

// get a tokoin, edit it, and put the new one back in the blockchain
func EditPolicy(wallet wlt.Wallet, URPOSet *URPOSet, txId []byte, time, id, gps, temper string) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := wlt.HashPubKey(wallet.PublicKey)
	output := URPOSet.FindOutput(txId)
	if !(output.IsLockedWithKey(pubKeyHash)) {
		log.Panic("ERROR: Not locked with this key")
	}

	input := TXInput{txId, 0, nil, wallet.PublicKey}
	inputs = append(inputs, input)

	// Build a list of outputs
	//from := fmt.Sprintf("%s", wallet.GetAddress())
	newOutput := output
	// update if the transfered parameters are not empty
	if time != "" {
		newOutput.Time, _ = strconv.Atoi(time)
	}
	if id != "" {
		newOutput.ID = []byte(id)
	}
	if gps != "" {
		newOutput.GPS, _ = strconv.Atoi(gps)
	}
	if temper != "" {
		newOutput.Temperature, _ = strconv.Atoi(temper)
	}
	outputs = append(outputs, newOutput)

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	URPOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

func RevocatTokoin(wallet wlt.Wallet, URPOSet *URPOSet, txId []byte) *Transaction {
	var inputs []TXInput

	pubKeyHash := wlt.HashPubKey(wallet.PublicKey)
	output := URPOSet.FindOutput(txId)
	if !(output.IsLockedWithKey(pubKeyHash)) {
		log.Panic("ERROR: Not locked with this key")
	}

	input := TXInput{txId, 0, nil, wallet.PublicKey}
	inputs = append(inputs, input)

	tx := Transaction{nil, inputs, []TXOutput{}}
	tx.ID = tx.Hash()
	URPOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

func RedeemTokoin(wallet wlt.Wallet, holder string, URPOSet *URPOSet, txId []byte, time *int, id *[]byte, gps, temper *int) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	ownerKey := wlt.HashPubKey(wallet.PublicKey)
	output := URPOSet.FindOutput(txId)
	if !(output.IsLockedWithKey(ownerKey)) {
		log.Panic("ERROR: Wrong owner")
	}
	holderKey := utils.Base58Decode([]byte(holder))
	holderKey = holderKey[1 : len(holderKey)-4]
	if !(output.IsHeldWithKey(holderKey)) {
		log.Panic("ERROR: Wrong holder")
	}

	//Check the redeem condition
	if !(output.CheckCondition(time, id, gps, temper)) {
		log.Panic("ERROR: Condition not satisfied")
	}

	input := TXInput{txId, 0, nil, wallet.PublicKey}
	inputs = append(inputs, input)

	// Build a list of outputs
	newOutput := output
	// Remove the holderkey
	newOutput.HolderKey = nil
	outputs = append(outputs, newOutput)

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	URPOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
