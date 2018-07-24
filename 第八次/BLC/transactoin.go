package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"strings"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
)

// 创世区块，Token数量
const subsidy  = 10

type QYH_Transaction struct {
	QYH_ID   []byte
	QYH_Vin  []QYH_TXInput
	QYH_Vout []QYH_TXOutput
}

// 是否是创世区块交易
func (tx QYH_Transaction) QYH_IsCoinbase() bool {
	// Vin 只有一条
	// Vin 第一条数据的Txid 为 0
	// Vin 第一条数据的Vout 为 -1
	return len(tx.QYH_Vin) == 1 && len(tx.QYH_Vin[0].QYH_Txid) == 0 && tx.QYH_Vin[0].QYH_Vout == -1
}


// 将交易进行Hash
func (tx *QYH_Transaction) QYH_Hash() []byte  {
	var hash [32]byte

	txCopy := *tx
	txCopy.QYH_ID = []byte{}

	hash = sha256.Sum256(txCopy.QYH_Serialize())
	return hash[:]
}
// 新建创世区块的交易
func QYH_NewCoinbaseTX(to ,data string) *QYH_Transaction  {
	if data == ""{
		//如果数据为空，可以随机给默认数据,用于挖矿奖励
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}
	txin := QYH_TXInput{[]byte{},-1,nil,[]byte(data)}
	txout := QYH_NewTXOutput(subsidy,to)

	tx := QYH_Transaction{nil,[]QYH_TXInput{txin},[]QYH_TXOutput{*txout}}
	tx.QYH_ID = tx.QYH_Hash()
	return &tx
}

// 转帐时生成交易
func QYH_NewUTXOTransaction(wallet *QYH_Wallet,to string,amount int,UTXOSet *QYH_UTXOSet,txs []*QYH_Transaction) *QYH_Transaction   {

	// 如果本区块中，多笔转账
	/**
	第一种情况：
	  A:10
	  A->B 2
	  A->C 4

	  tx1:
	      Vin:
	           ATxID  out ...
	      Vout:
	           A : 8
	           B : 2
	  tx1:
	      Vin:
	           ATxID  out ...
	      Vout:
	           A : 4
	           C : 4
	第二种情况：
	  A:10+10
	  A->B 4
	  A->C 8
	**/

	pubKeyHash := QYH_HashPubKey(wallet.QYH_PublicKey)
	if len(txs) > 0 {
		// 查的txs中的UTXO
		utxo := QYH_FindUTXOFromTransactions(txs)

		// 找出当前钱包已经花费的
		unspentOutputs := make(map[string][]int)
		acc := 0
		for txID,outs := range utxo {
			for outIdx, out := range outs.QYH_Outputs {
				if out.QYH_IsLockedWithKey(pubKeyHash) && acc < amount {
					acc += out.QYH_Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		if acc >= amount { // 当前交易中的剩余余额可以支付
			fmt.Println("txs>0 && acc >= amount")
			return QYH_NewUTXOTransactionEnd(wallet,to,amount,UTXOSet,acc,unspentOutputs,txs)
		}else{
			fmt.Println("txs>0 && acc < amount")
			accLeft, validOutputs := UTXOSet.QYH_FindSpendableOutputs(pubKeyHash,  amount - acc)
			for k,v := range unspentOutputs{
				validOutputs[k] = v
			}
			return QYH_NewUTXOTransactionEnd(wallet,to,amount,UTXOSet,acc + accLeft,validOutputs,txs)
		}
	} else { //只是当前一笔交易
		fmt.Println("txs==0")
		acc, validOutputs := UTXOSet.QYH_FindSpendableOutputs(pubKeyHash, amount)

		return QYH_NewUTXOTransactionEnd(wallet,to,amount,UTXOSet,acc,validOutputs,txs)
	}
}

func QYH_NewUTXOTransactionEnd(wallet *QYH_Wallet,to string,amount int,UTXOSet *QYH_UTXOSet,acc int, UTXO map[string][]int,txs []*QYH_Transaction) *QYH_Transaction {

	if acc < amount {
		log.Panic("账户余额不足")
	}

	var inputs []QYH_TXInput
	var outputs []QYH_TXOutput
	// 构造input
	for txid, outs := range UTXO {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := QYH_TXInput{txID, out, nil, wallet.QYH_PublicKey}
			inputs = append(inputs, input)
		}
	}
	// 生成交易输出
	outputs = append(outputs, *QYH_NewTXOutput(amount, to))
	// 生成余额
	if acc > amount {
		outputs = append(outputs, *QYH_NewTXOutput(acc - amount, string(wallet.QYH_GetAddress())))
	}

	tx := QYH_Transaction{nil, inputs, outputs}
	tx.QYH_ID = tx.QYH_Hash()
	// 签名

	//tx.String()
	UTXOSet.QYH_Blockchain.QYH_SignTransaction(&tx, wallet.QYH_PrivateKey,txs)

	return &tx
}


// 找出交易中的utxo
func QYH_FindUTXOFromTransactions(txs []*QYH_Transaction) map[string]QYH_TXOutputs {
	UTXO := make(map[string]QYH_TXOutputs)
	// 已经花费的交易txID : TXOutputs.index
	spentTXOs := make(map[string][]int)
	// 循环区块中的交易
	for _, tx := range txs {
		// 将区块中的交易hash，转为字符串
		txID := hex.EncodeToString(tx.QYH_ID)

	Outputs:
		for outIdx, out := range tx.QYH_Vout { // 循环交易中的 TXOutputs
			// Was the output spent?
			// 如果已经花费的交易输出中，有此输出，证明已经花费
			if spentTXOs[txID] != nil {
				for _, spentOutIdx := range spentTXOs[txID] {
					if spentOutIdx == outIdx { // 如果花费的正好是此笔输出
						continue Outputs // 继续下一次循环
					}
				}
			}

			outs := UTXO[txID] // 获取UTXO指定txID对应的TXOutputs
			outs.QYH_Outputs = append(outs.QYH_Outputs, out)
			UTXO[txID] = outs
		}

		if tx.QYH_IsCoinbase() == false { // 非创世区块
			for _, in := range tx.QYH_Vin {
				inTxID := hex.EncodeToString(in.QYH_Txid)
				spentTXOs[inTxID] = append(spentTXOs[inTxID], in.QYH_Vout)
			}
		}
	}
	return UTXO

}

// 签名
func (tx *QYH_Transaction) QYH_Sign(privateKey ecdsa.PrivateKey,prevTXs map[string]QYH_Transaction)  {
	if tx.QYH_IsCoinbase() { // 创世区块不需要签名
		return
	}

	// 检查交易的输入是否正确
	for _,vin := range tx.QYH_Vin{
		if prevTXs[hex.EncodeToString(vin.QYH_Txid)].QYH_ID == nil{
			log.Panic("错误：之前的交易不正确")
		}
	}

	txCopy := tx.QYH_TrimmedCopy()

	for inID, vin := range txCopy.QYH_Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.QYH_Txid)]
		txCopy.QYH_Vin[inID].QYH_Signature = nil
		txCopy.QYH_Vin[inID].QYH_PubKey = prevTx.QYH_Vout[vin.QYH_Vout].QYH_PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.QYH_Vin[inID].QYH_Signature = signature
		txCopy.QYH_Vin[inID].QYH_PubKey = nil
	}
}
// 验证签名
func (tx *QYH_Transaction) QYH_Verify(prevTXs map[string]QYH_Transaction) bool {
	if tx.QYH_IsCoinbase() {
		return true
	}

	for _, vin := range tx.QYH_Vin {
		if prevTXs[hex.EncodeToString(vin.QYH_Txid)].QYH_ID == nil {
			log.Panic("错误：之前的交易不正确")
		}
	}

	txCopy := tx.QYH_TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.QYH_Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.QYH_Txid)]
		txCopy.QYH_Vin[inID].QYH_Signature = nil
		txCopy.QYH_Vin[inID].QYH_PubKey = prevTx.QYH_Vout[vin.QYH_Vout].QYH_PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.QYH_Signature)
		r.SetBytes(vin.QYH_Signature[:(sigLen / 2)])
		s.SetBytes(vin.QYH_Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.QYH_PubKey)
		x.SetBytes(vin.QYH_PubKey[:(keyLen / 2)])
		y.SetBytes(vin.QYH_PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.QYH_Vin[inID].QYH_PubKey = nil
	}

	return true
}

// 复制交易（输入的签名和公钥置为空）
func (tx *QYH_Transaction) QYH_TrimmedCopy() QYH_Transaction {
	var inputs []QYH_TXInput
	var outputs []QYH_TXOutput

	for _, vin := range tx.QYH_Vin {
		inputs = append(inputs, QYH_TXInput{vin.QYH_Txid, vin.QYH_Vout, nil, nil})
	}

	for _, vout := range tx.QYH_Vout {
		outputs = append(outputs, QYH_TXOutput{vout.QYH_Value, vout.QYH_PubKeyHash})
	}

	txCopy := QYH_Transaction{tx.QYH_ID, inputs, outputs}

	return txCopy
}
// 打印交易内容
func (tx QYH_Transaction) QYH_String()  {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction ID: %x", tx.QYH_ID))

	for i, input := range tx.QYH_Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.QYH_Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.QYH_Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.QYH_Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.QYH_PubKey))
	}

	for i, output := range tx.QYH_Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.QYH_Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.QYH_PubKeyHash))
	}
	fmt.Println(strings.Join(lines, "\n"))
}


// 将交易序列化
func (tx QYH_Transaction) QYH_Serialize() []byte  {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)

	if err != nil{
		log.Panic(err)
	}
	return encoded.Bytes()
}
// 反序列化交易
func QYH_DeserializeTransaction(data []byte) QYH_Transaction {
	var transaction QYH_Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

// 将交易数组序列化
func QYH_SerializeTransactions(txs []*QYH_Transaction) [][]byte  {

	var txsHash [][]byte
	for _,tx := range txs{
		txsHash = append(txsHash, tx.QYH_Serialize())
	}
	return txsHash
}

// 反序列化交易数组
func QYH_DeserializeTransactions(data [][]byte) []QYH_Transaction {
	var txs []QYH_Transaction
	for _,tx := range data {
		var transaction QYH_Transaction
		decoder := gob.NewDecoder(bytes.NewReader(tx))
		err := decoder.Decode(&transaction)
		if err != nil {
			log.Panic(err)
		}

		txs = append(txs, transaction)
	}
	return txs
}
