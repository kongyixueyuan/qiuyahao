package BLC

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"fmt"
)

type QYH_Block struct {
	QYH_TimeStamp     int64
	QYH_Transactions   []*QYH_Transaction
	QYH_PrevBlockHash []byte
	QYH_Hash          []byte
	QYH_Nonce         int
	QYH_Height        int
}
// 生成新的区块
func QYH_NewBlock(transactions []*QYH_Transaction, prevBlockHash []byte, height int) *QYH_Block {
	// 生成新的区块对象
	block := &QYH_Block{
		time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte{},
		0,
		height,
	}
	// 挖矿

	pow := QYH_NewProofOfWork(block)
	nonce,hash :=pow.QYH_Run()

	block.QYH_Nonce = nonce
	block.QYH_Hash = hash[:]

	return block

}

// 将交易进行hash
func (b QYH_Block) QYH_HashTransactions() []byte {
	var transactions [][]byte
	// 获取交易真实内容
	for _,tx := range b.QYH_Transactions{
		transactions = append(transactions,tx.QYH_Serialize())
	}
	//txHash := sha256.Sum256(bytes.Join(transactions,[]byte{}))
	mTree := QYH_NewMerkelTree(transactions)
	return mTree.QYH_RootNode.QYH_Data
}

// 新建创世区块
func QYH_NewGenesisBlock(coinbase *QYH_Transaction) *QYH_Block  {
	return QYH_NewBlock([]*QYH_Transaction{coinbase},[]byte{},1)
}

// 序列化
func (b *QYH_Block) QYH_Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func QYH_DeserializeBlock(d []byte) *QYH_Block {
	var block QYH_Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
// 打印区块内容
func (block QYH_Block) QYH_String()  {
	fmt.Println("\n==============")
	fmt.Printf("Height:\t%d\n", block.QYH_Height)
	fmt.Printf("PrevBlockHash:\t%x\n", block.QYH_PrevBlockHash)
	fmt.Printf("Timestamp:\t%s\n", time.Unix(block.QYH_TimeStamp, 0).Format("2006-01-02 03:04:05 PM"))
	fmt.Printf("Hash:\t%x\n", block.QYH_Hash)
	fmt.Printf("Nonce:\t%d\n", block.QYH_Nonce)
	fmt.Println("Txs:")

	for _, tx := range block.QYH_Transactions {
		tx.QYH_String()
	}
	fmt.Println("==============")
}
