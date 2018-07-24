package BLC

import "bytes"

type QYH_TXOutput struct {
	QYH_Value  int
	QYH_PubKeyHash []byte
}
// 根据地址获取 PubKeyHash
func (out *QYH_TXOutput) QYH_Lock(address []byte) {
	pubKeyHash := QYH_Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.QYH_PubKeyHash = pubKeyHash
}

// 判断是否当前公钥对应的交易输出(是否是某个人的交易输出)
func (out *QYH_TXOutput) QYH_IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.QYH_PubKeyHash, pubKeyHash) == 0
}

func QYH_NewTXOutput(value int, address string) *QYH_TXOutput {
	txo := &QYH_TXOutput{value, nil}
	txo.QYH_Lock([]byte(address))
	return txo
}


