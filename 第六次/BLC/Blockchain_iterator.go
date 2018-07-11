package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

type QYH_BlockchainIterator struct {
	qyh_currentHash []byte
	qyh_db          *bolt.DB
}

func (i *QYH_BlockchainIterator) QYH_Next() *QYH_Block {
	var block *QYH_Block

	err := i.qyh_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.qyh_currentHash)
		block = QYH_DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.qyh_currentHash = block.QYH_PrevBlockHash

	return block
}
