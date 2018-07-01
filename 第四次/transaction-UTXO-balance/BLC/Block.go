package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
	"fmt"
	"crypto/sha256"
)

type Block struct {

	Height int64

	PreBlockHash []byte

	Txs []*Transaction

	Timestamp int64

	Hash []byte

	Nonce int64

}

// 需要将txs转换成[]byte
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Txs  {
		txHashes = append(txHashes, tx.TxHash)

	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// 将区块序列化成字节数组
func (block *Block) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func DeserializeBlock(blockBytes []byte) *Block {

	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// 创建新的区块
func NewBlock(txs []*Transaction, height int64, preBlockHash []byte) *Block {

	// 创建区块
	block := &Block{height, preBlockHash, txs, time.Now().Unix(), nil, 0}

	// 调用工作量证明的方法并且返回有效的Hash和Nonce
	pow := NewProofOfWork(block)

	// 挖矿验证
	hash, nonce := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Println()

	return block

}

// 单独写一个方法，生成创世区块
func CreateGenenisBlock(txs []*Transaction) *Block {

	return NewBlock(txs, 1, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})

}