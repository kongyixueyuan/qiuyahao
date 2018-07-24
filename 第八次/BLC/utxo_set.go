package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"strings"
)

const utxoBucket = "chainstate"

type QYH_UTXOSet struct {
	QYH_Blockchain *QYH_Blockchain
}

// 查询可花费的交易输出
func (u QYH_UTXOSet) QYH_FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.QYH_Blockchain.QYH_db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := QYH_DeserializeOutputs(v)

			for outIdx, out := range outs.QYH_Outputs {
				if out.QYH_IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.QYH_Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

func (u QYH_UTXOSet) QYH_Reset() {
	db := u.QYH_Blockchain.QYH_db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		// 删除旧的bucket
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic()
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.QYH_Blockchain.QYH_FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(key, outs.QYH_Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

// 生成新区块的时候，更新UTXO数据库
func (u QYH_UTXOSet) QYH_Update(block *QYH_Block) {
	err := u.QYH_Blockchain.QYH_db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.QYH_Transactions {
			if !tx.QYH_IsCoinbase() {
				for _, vin := range tx.QYH_Vin {
					updatedOuts := QYH_TXOutputs{}
					outsBytes := b.Get(vin.QYH_Txid)
					outs := QYH_DeserializeOutputs(outsBytes)

					// 找出Vin对应的outputs,过滤掉花费的
					for outIndex, out := range outs.QYH_Outputs {
						if outIndex != vin.QYH_Vout {
							updatedOuts.QYH_Outputs = append(updatedOuts.QYH_Outputs, out)
						}
					}
					// 未花费的交易输出TXOutput为0
					if len(updatedOuts.QYH_Outputs) == 0 {
						err := b.Delete(vin.QYH_Txid)
						if err != nil {
							log.Panic(err)
						}
					} else { // 未花费的交易输出TXOutput>0
						err := b.Put(vin.QYH_Txid, updatedOuts.QYH_Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			// 将所有的交易输出TXOutput存入数据库中
			newOutputs := QYH_TXOutputs{}
			for _, out := range tx.QYH_Vout {
				newOutputs.QYH_Outputs = append(newOutputs.QYH_Outputs, out)
			}
			err := b.Put(tx.QYH_ID, newOutputs.QYH_Serialize())
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

// 打出某个公钥hash对应的所有未花费输出
func (u *QYH_UTXOSet) QYH_FindUTXO(pubKeyHash []byte) []QYH_TXOutput {
	var UTXOs []QYH_TXOutput

	err := u.QYH_Blockchain.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := QYH_DeserializeOutputs(v)

			for _, out := range outs.QYH_Outputs {
				if out.QYH_IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// 查询某个地址的余额
func (u *QYH_UTXOSet) QYH_GetBalance(address string) int {
	balance := 0
	pubKeyHash := QYH_Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := u.QYH_FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.QYH_Value
	}
	return balance
}

// 打印所有的UTXO
func (u *QYH_UTXOSet) QYH_String() {
	//outputs := make(map[string][]Yxh_TXOutput)

	var lines []string
	lines = append(lines, "---ALL UTXO:")
	err := u.QYH_Blockchain.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := QYH_DeserializeOutputs(v)

			lines = append(lines, fmt.Sprintf("     Key: %s", txID))
			for i, out := range outs.QYH_Outputs {
				//outputs[txID] = append(outputs[txID], out)
				lines = append(lines, fmt.Sprintf("     Output: %d", i))
				lines = append(lines, fmt.Sprintf("         value:  %d", out.QYH_Value))
				lines = append(lines, fmt.Sprintf("         PubKeyHash:  %x", out.QYH_PubKeyHash))
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(strings.Join(lines, "\n"))
}
