package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/zhuaiballl/Go-Tokoin/utils"
	"log"
)

// TXOutput represents a transaction output
type TXOutput struct {
	Time        int
	ID          []byte
	GPS         int
	Temperature int
	PubKeyHash  []byte
	HolderKey   []byte
}

// Lock signs the output
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// Hold signs the output with holder's key
func (out *TXOutput) Hold(address []byte) {
	holderKey := utils.Base58Decode(address)
	holderKey = holderKey[1 : len(holderKey)-4]
	out.HolderKey = holderKey
}

// IsHeldWithKey checks if the output is held by the owner of the holderkey
func (out *TXOutput) IsHeldWithKey(holderKey []byte) bool {
	return bytes.Compare(out.HolderKey, holderKey) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(time int, id []byte, gps int, temper int, address string) *TXOutput {
	txo := &TXOutput{time, id, gps, temper, nil, nil}
	txo.Lock([]byte(address))

	return txo
}

func (out *TXOutput) Show() {
	fmt.Printf("{\n")
	fmt.Printf("Time: %d\n", out.Time)
	fmt.Printf("GPS: %d\n", out.GPS)
	fmt.Printf("Temperature: %d\n", out.Temperature)
	fmt.Printf("}\n")
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

func (out *TXOutput) CheckCondition(time *int, ID *[]byte, GPS *int, temper *int) bool {
	//if !(out.checkTime(time)){return false}
	//if !(out.checkID(ID)){return false}
	//if !(out.checkGPS(GPS)){return false}
	//if !(out.checkTemper(temper)){return false}
	return *time == out.Time && bytes.Compare(*ID, out.ID) == 0 && *GPS == out.GPS && *temper == out.Temperature
}
