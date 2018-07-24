package BLC

import "bytes"

type QYH_TXInput struct {
	QYH_Txid      []byte
	QYH_Vout      int      // Vout的index
	QYH_Signature []byte   // 签名
	QYH_PubKey    []byte   // 公钥
}

func (in QYH_TXInput) QYH_UsesKey(pubKeyHash []byte) bool  {
	lockingHash := QYH_HashPubKey(in.QYH_PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
