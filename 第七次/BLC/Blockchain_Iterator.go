package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

type QYH_BlockchainIterator struct {
	QYH_CurrentHash []byte
	QYH_DB  *bolt.DB
}

func (blockchainIterator *QYH_BlockchainIterator) QYH_Next() *QYH_Block {

	var block *QYH_Block

	err := blockchainIterator.QYH_DB.View(func(tx *bolt.Tx) error{

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			currentBloclBytes := b.Get(blockchainIterator.QYH_CurrentHash)
			//  获取到当前迭代器里面的currentHash所对应的区块
			block = QYH_DeserializeBlock(currentBloclBytes)

			// 更新迭代器里面CurrentHash
			blockchainIterator.QYH_CurrentHash = block.QYH_PrevBlockHash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}


	return block

}