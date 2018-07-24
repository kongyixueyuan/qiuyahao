package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
)

type QYH_TXOutputs struct {
	QYH_Outputs []QYH_TXOutput
}

//  序列化 TXOutputs
func (outs QYH_TXOutputs) QYH_Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// 反序列化 TXOutputs
func QYH_DeserializeOutputs(data []byte) QYH_TXOutputs {
	var outputs QYH_TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
