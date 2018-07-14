package BLC

import (
	"time"
	"fmt"
	"bytes"
	"encoding/gob"
	"log"
)

type QYH_Block struct {
	//1. 区块高度
	QYH_Height int64
	//2. 上一个区块HASH
	QYH_PrevBlockHash []byte
	//3. 交易数据
	QYH_Txs []*QYH_Transaction
	//4. 时间戳
	QYH_Timestamp int64
	//5. Hash
	QYH_Hash []byte
	// 6. Nonce
	QYH_Nonce int64
}


// 需要将Txs转换成[]byte
func (block *QYH_Block) QYH_HashTransactions() []byte  {


	//var txHashes [][]byte
	//var txHash [32]byte
	//
	//for _, tx := range block.Txs {
	//	txHashes = append(txHashes, tx.TxHash)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//
	//return txHash[:]

	var transactions [][]byte

	for _, tx := range block.QYH_Txs {
		transactions = append(transactions, tx.QYH_Serialize())
	}
	mTree := QYH_NewMerkleTree(transactions)

	return mTree.QYH_RootNode.QYH_Data

}


// 将区块序列化成字节数组
func (block *QYH_Block) QYH_Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func QYH_DeserializeBlock(blockBytes []byte) *QYH_Block {

	var block QYH_Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}


//1. 创建新的区块
func QYH_NewBlock(txs []*QYH_Transaction,height int64,prevBlockHash []byte) *QYH_Block {

	//创建区块
	block := &QYH_Block{height,prevBlockHash,txs,time.Now().Unix(),nil,0}

	// 调用工作量证明的方法并且返回有效的Hash和Nonce
	pow := QYH_NewProofOfWork(block)

	// 挖矿验证
	hash,nonce := pow.QYH_Run()

	block.QYH_Hash = hash[:]
	block.QYH_Nonce = nonce

	fmt.Println()

	return block

}

//2. 单独写一个方法，生成创世区块

func QYH_CreateGenesisBlock(txs []*QYH_Transaction) *QYH_Block {


	return QYH_NewBlock(txs,1, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
}

