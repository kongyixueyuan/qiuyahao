package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

type QYH_BlockchainIterator struct {
	QYH_currentHash []byte
	QYH_db          *bolt.DB
}

func (i *QYH_BlockchainIterator) QYH_Next() *QYH_Block {
	var block *QYH_Block

	err := i.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.QYH_currentHash)
		block = QYH_DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.QYH_currentHash = block.QYH_PrevBlockHash

	return block
}