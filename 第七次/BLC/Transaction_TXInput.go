package BLC

import "bytes"

type QYH_TXInput struct {
	// 1. 交易的Hash
	QYH_TxHash      []byte
	// 2. 存储TXOutput在Vout里面的索引
	QYH_Vout      int

	QYH_Signature []byte // 数字签名

	QYH_PublicKey    []byte // 公钥，钱包里面
}



// 判断当前的消费是谁的钱
func (txInput *QYH_TXInput) QYH_UnLockRipemd160Hash(ripemd160Hash []byte) bool {

	publicKey := QYH_Ripemd160Hash(txInput.QYH_PublicKey)

	return bytes.Compare(publicKey,ripemd160Hash) == 0
}