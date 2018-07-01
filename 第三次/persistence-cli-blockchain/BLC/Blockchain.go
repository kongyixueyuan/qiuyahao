package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"math/big"
	"time"
	"os"
)

// 数据库名字
const dbName  = "blockchain.db"
const blockTableName = "blocks"

type Blockchain struct {
	Tip []byte //最新的区块hash
	DB *bolt.DB
}

// 遍历输出所有区块的信息
func (blc *Blockchain) Printchain() {

	blockchainIterator := blc.Iterator()

	for {
		block := blockchainIterator.Next()

		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("PreBlockHash: %x\n", block.PreBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		// 时间格式必须是"2006-01-02 03:04:05 PM"
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PreBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

}

// 增加区块到区块链里面
func (blc *Blockchain) AddBlockToBlockchain(data string) {

	err := blc.DB.Update(func(tx *bolt.Tx) error {

		// 获取表
		b := tx.Bucket([]byte(blockTableName))
		// 创建新区块
		if b != nil {

			// 取当前最新区块
			blockBytes := b.Get(blc.Tip)

			block := DeserializeBlock(blockBytes)

			// 将区块序列化并且存储到数据库
			newBlock := NewBlock(data, block.Height + 1, block.Hash)

			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			// 更新数据库里面"l"对于的hash
			err = b.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			// 更新blockchain的Tip
			blc.Tip = newBlock.Hash

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// 判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

// 创建带有创世区块的区块链
func CreateBlockchainWithGenenisBlock(data string) {

	if DBExists() {
		fmt.Println("创世区块已经存在")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块")

	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// 更新数据库
	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if b != nil {
			// 创建创世区块
			genesisBlock := CreateGenenisBlock(data)

			// 将创世区块存储到表中
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())

			if err != nil {
				log.Panic(err)
			}

			//存储最新的区块的Hash
			err = b.Put([]byte("l"), genesisBlock.Hash)

			if err != nil {
				log.Panic(err)
			}

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}

// 返回Blockchain对象
func BlockchainObject() *Blockchain {

	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			// 读取最新区块的hash
			tip = b.Get([]byte("l"))
		}

		return nil
	})
	return &Blockchain{tip, db}

}
