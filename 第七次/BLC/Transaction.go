package BLC

import (
	"bytes"
	"log"
	"encoding/gob"
	"crypto/sha256"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"

	"math/big"
	"crypto/elliptic"
	"time"
)

// UTXO
type QYH_Transaction struct {

	//1. 交易hash
	QYH_TxHash []byte

	//2. 输入
	QYH_Vins []*QYH_TXInput

	//3. 输出
	QYH_Vouts []*QYH_TXOutput
}

//[]byte{}

// 判断当前的交易是否是Coinbase交易
func (tx *QYH_Transaction) QYH_IsCoinbaseTransaction() bool {

	return len(tx.QYH_Vins[0].QYH_TxHash) == 0 && tx.QYH_Vins[0].QYH_Vout == -1
}



//1. Transaction 创建分两种情况
//1. 创世区块创建时的Transaction
func QYH_NewCoinbaseTransaction(address string) *QYH_Transaction {

	//代表消费
	txInput := &QYH_TXInput{[]byte{},-1,nil,[]byte{}}


	txOutput := QYH_NewTXOutput(10,address)

	txCoinbase := &QYH_Transaction{[]byte{},[]*QYH_TXInput{txInput},[]*QYH_TXOutput{txOutput}}

	//设置hash值
	txCoinbase.QYH_HashTransaction()


	return txCoinbase
}

func (tx *QYH_Transaction) QYH_HashTransaction()  {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()),result.Bytes()},[]byte{})

	hash := sha256.Sum256(resultBytes)

	tx.QYH_TxHash = hash[:]
}



//2. 转账时产生的Transaction

func QYH_NewSimpleTransaction(from string,to string,amount int64, utxoSet *QYH_UTXOSet,txs []*QYH_Transaction,nodeID string) *QYH_Transaction {

	//$ ./bc send -from '["juncheng"]' -to '["zhangqiang"]' -amount '["2"]'
	//	[juncheng]
	//	[zhangqiang]
	//	[2]

	wallets,_ := QYH_NewWallets(nodeID)
	wallet := wallets.QYH_WalletsMap[from]


	// 通过一个函数，返回
	money,spendableUTXODic := utxoSet.QYH_FindSpendableUTXOS(from,amount,txs)
	//
	//	{hash1:[0],hash2:[2,3]}

	var txIntputs []*QYH_TXInput
	var txOutputs []*QYH_TXOutput

	for txHash,indexArray := range spendableUTXODic  {

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {

			txInput := &QYH_TXInput{txHashBytes,index,nil,wallet.QYH_PublicKey}
			txIntputs = append(txIntputs,txInput)
		}

	}

	// 转账
	txOutput := QYH_NewTXOutput(int64(amount),to)
	txOutputs = append(txOutputs,txOutput)

	// 找零
	txOutput = QYH_NewTXOutput(int64(money) - int64(amount),from)
	txOutputs = append(txOutputs,txOutput)

	tx := &QYH_Transaction{[]byte{},txIntputs,txOutputs}

	//设置hash值
	tx.QYH_HashTransaction()

	//进行签名
	utxoSet.QYH_Blockchain.QYH_SignTransaction(tx, wallet.QYH_PrivateKey,txs)

	return tx

}

func (tx *QYH_Transaction) QYH_Hash() []byte {

	txCopy := tx

	txCopy.QYH_TxHash = []byte{}

	hash := sha256.Sum256(txCopy.QYH_Serialize())
	return hash[:]
}


func (tx *QYH_Transaction) QYH_Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}


func (tx *QYH_Transaction) QYH_Sign(privKey ecdsa.PrivateKey, prevTXs map[string]QYH_Transaction) {

	if tx.QYH_IsCoinbaseTransaction() {
		return
	}


	for _, vin := range tx.QYH_Vins {
		if prevTXs[hex.EncodeToString(vin.QYH_TxHash)].QYH_TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}


	txCopy := tx.QYH_TrimmedCopy()

	for inID, vin := range txCopy.QYH_Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.QYH_TxHash)]
		txCopy.QYH_Vins[inID].QYH_Signature = nil
		txCopy.QYH_Vins[inID].QYH_PublicKey = prevTx.QYH_Vouts[vin.QYH_Vout].QYH_Ripemd160Hash
		txCopy.QYH_TxHash = txCopy.QYH_Hash()
		txCopy.QYH_Vins[inID].QYH_PublicKey = nil

		// 签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.QYH_TxHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.QYH_Vins[inID].QYH_Signature = signature
	}
}


// 拷贝一份新的Transaction用于签名                                    T
func (tx *QYH_Transaction) QYH_TrimmedCopy() QYH_Transaction {
	var inputs []*QYH_TXInput
	var outputs []*QYH_TXOutput

	for _, vin := range tx.QYH_Vins {
		inputs = append(inputs, &QYH_TXInput{vin.QYH_TxHash, vin.QYH_Vout, nil, nil})
	}

	for _, vout := range tx.QYH_Vouts {
		outputs = append(outputs, &QYH_TXOutput{vout.QYH_Value, vout.QYH_Ripemd160Hash})
	}

	txCopy := QYH_Transaction{tx.QYH_TxHash, inputs, outputs}

	return txCopy
}


// 数字签名验证

func (tx *QYH_Transaction) QYH_Verify(prevTXs map[string]QYH_Transaction) bool {
	if tx.QYH_IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range tx.QYH_Vins {
		if prevTXs[hex.EncodeToString(vin.QYH_TxHash)].QYH_TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.QYH_TrimmedCopy()

	curve := elliptic.P256()

	for inID, vin := range tx.QYH_Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.QYH_TxHash)]
		txCopy.QYH_Vins[inID].QYH_Signature = nil
		txCopy.QYH_Vins[inID].QYH_PublicKey = prevTx.QYH_Vouts[vin.QYH_Vout].QYH_Ripemd160Hash
		txCopy.QYH_TxHash = txCopy.QYH_Hash()
		txCopy.QYH_Vins[inID].QYH_PublicKey = nil


		// 私钥 ID
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.QYH_Signature)
		r.SetBytes(vin.QYH_Signature[:(sigLen / 2)])
		s.SetBytes(vin.QYH_Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.QYH_PublicKey)
		x.SetBytes(vin.QYH_PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.QYH_PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.QYH_TxHash, &r, &s) == false {
			return false
		}
	}

	return true
}
