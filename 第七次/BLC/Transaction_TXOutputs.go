package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
)

type QYH_TXOutputs struct {
	QYH_UTXOS []*QYH_UTXO
}


// 将区块序列化成字节数组
func (txOutputs *QYH_TXOutputs) QYH_Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func QYH_DeserializeTXOutputs(txOutputsBytes []byte) *QYH_TXOutputs {

	var txOutputs QYH_TXOutputs

	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return &txOutputs
}