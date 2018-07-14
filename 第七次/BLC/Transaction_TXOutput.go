package BLC

import "bytes"


type QYH_TXOutput struct {
	QYH_Value int64
	QYH_Ripemd160Hash []byte  //用户名
}

func (txOutput *QYH_TXOutput) QYH_Lock(address string)  {

	publicKeyHash := QYH_Base58Decode([]byte(address))

	txOutput.QYH_Ripemd160Hash = publicKeyHash[1:len(publicKeyHash) - 4]
}


func QYH_NewTXOutput(value int64,address string) *QYH_TXOutput {

	txOutput := &QYH_TXOutput{value,nil}

	// 设置Ripemd160Hash
	txOutput.QYH_Lock(address)

	return txOutput
}


// 解锁
func (txOutput *QYH_TXOutput) QYH_UnLockScriptPubKeyWithAddress(address string) bool {

	publicKeyHash := QYH_Base58Decode([]byte(address))
	hash160 := publicKeyHash[1:len(publicKeyHash) - 4]

	return bytes.Compare(txOutput.QYH_Ripemd160Hash,hash160) == 0
}



