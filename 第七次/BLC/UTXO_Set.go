package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"bytes"
)



const utxoTableName  = "utxoTableName"

type QYH_UTXOSet struct {
	QYH_Blockchain *QYH_Blockchain
}

// 重置数据库表
func (utxoSet *QYH_UTXOSet) QYH_ResetUTXOSet()  {

	err := utxoSet.QYH_Blockchain.QYH_DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {


			err := tx.DeleteBucket([]byte(utxoTableName))

			if err!= nil {
				log.Panic(err)
			}

		}

		b ,_ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil {

			//[string]*TXOutputs
			txOutputsMap := utxoSet.QYH_Blockchain.QYH_FindUTXOMap()


			for keyHash,outs := range txOutputsMap {

				txHash,_ := hex.DecodeString(keyHash)

				b.Put(txHash,outs.QYH_Serialize())

			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("重置失败....")
		log.Panic(err)
	}

}

func (utxoSet *QYH_UTXOSet) QYH_findUTXOForAddress(address string) []*QYH_UTXO{


	var utxos []*QYH_UTXO

	utxoSet.QYH_Blockchain.QYH_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		// 游标
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			txOutputs := QYH_DeserializeTXOutputs(v)

			for _,utxo := range txOutputs.QYH_UTXOS  {

				if utxo.QYH_Output.QYH_UnLockScriptPubKeyWithAddress(address) {
					utxos = append(utxos,utxo)
				}
			}
		}

		return nil
	})

	return utxos
}




func (utxoSet *QYH_UTXOSet) QYH_GetBalance(address string) int64 {

	UTXOS := utxoSet.QYH_findUTXOForAddress(address)

	var amount int64

	for _,utxo := range UTXOS  {
		amount += utxo.QYH_Output.QYH_Value
	}

	return amount
}


// 返回要凑多少钱，对应TXOutput的TX的Hash和index
func (utxoSet *QYH_UTXOSet) QYH_FindUnPackageSpendableUTXOS(from string, txs []*QYH_Transaction) []*QYH_UTXO {

	var unUTXOs []*QYH_UTXO

	spentTXOutputs := make(map[string][]int)

	//{hash:[0]}

	for _,tx := range txs {

		if tx.QYH_IsCoinbaseTransaction() == false {
			for _, in := range tx.QYH_Vins {
				//是否能够解锁
				publicKeyHash := QYH_Base58Decode([]byte(from))

				ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]
				if in.QYH_UnLockRipemd160Hash(ripemd160Hash) {

					key := hex.EncodeToString(in.QYH_TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.QYH_Vout)
				}

			}
		}
	}


	for _,tx := range txs {

	Work1:
		for index, out := range tx.QYH_Vouts {

			if out.QYH_UnLockScriptPubKeyWithAddress(from) {
				fmt.Println("看看是否是俊诚...")
				fmt.Println(from)

				fmt.Println(spentTXOutputs)

				if len(spentTXOutputs) == 0 {
					utxo := &QYH_UTXO{tx.QYH_TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash,indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.QYH_TxHash)

						if hash == txHashStr {

							var isUnSpentUTXO bool

							for _,outIndex := range indexArray {

								if index == outIndex {
									isUnSpentUTXO = true
									continue Work1
								}

								if isUnSpentUTXO == false {
									utxo := &QYH_UTXO{tx.QYH_TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &QYH_UTXO{tx.QYH_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}

			}

		}

	}

	return unUTXOs

}

func (utxoSet *QYH_UTXOSet) QYH_FindSpendableUTXOS(from string,amount int64,txs []*QYH_Transaction) (int64,map[string][]int)  {

	unPackageUTXOS := utxoSet.QYH_FindUnPackageSpendableUTXOS(from,txs)

	spentableUTXO := make(map[string][]int)

	var money int64 = 0

	for _,UTXO := range unPackageUTXOS {

		money += UTXO.QYH_Output.QYH_Value;
		txHash := hex.EncodeToString(UTXO.QYH_TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash],UTXO.QYH_Index)
		if money >= amount{
			return  money,spentableUTXO
		}
	}


	// 钱还不够
	utxoSet.QYH_Blockchain.QYH_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {

			c := b.Cursor()
			UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {

				txOutputs := QYH_DeserializeTXOutputs(v)

				for _,utxo := range txOutputs.QYH_UTXOS {

					money += utxo.QYH_Output.QYH_Value
					txHash := hex.EncodeToString(utxo.QYH_TxHash)
					spentableUTXO[txHash] = append(spentableUTXO[txHash],utxo.QYH_Index)

					if money >= amount {
						 break UTXOBREAK;
					}
				}
			}

		}

		return nil
	})

	if money < amount{
		log.Panic("余额不足......")
	}


	return  money,spentableUTXO
}


// 更新
func (utxoSet *QYH_UTXOSet) QYH_Update()  {

	// blocks
	//


	// 最新的Block
	block := utxoSet.QYH_Blockchain.QYH_Iterator().QYH_Next()


	// utxoTable
	//

	ins := []*QYH_TXInput{}

	outsMap := make(map[string]*QYH_TXOutputs)

	// 找到所有我要删除的数据
	for _,tx := range block.QYH_Txs {

		for _,in := range tx.QYH_Vins {
			ins = append(ins,in)
		}
	}

	for _,tx := range block.QYH_Txs  {


		utxos := []*QYH_UTXO{}

		for index,out := range tx.QYH_Vouts  {

			isSpent := false

			for _,in := range ins  {

				if in.QYH_Vout == index && bytes.Compare(tx.QYH_TxHash ,in.QYH_TxHash) == 0 && bytes.Compare(out.QYH_Ripemd160Hash, QYH_Ripemd160Hash(in.QYH_PublicKey)) == 0 {

					isSpent = true
					continue
				}
			}

			if isSpent == false {
				utxo := &QYH_UTXO{tx.QYH_TxHash,index,out}
				utxos = append(utxos,utxo)
			}

		}

		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.QYH_TxHash)
			outsMap[txHash] = &QYH_TXOutputs{utxos}
		}

	}



	err := utxoSet.QYH_Blockchain.QYH_DB.Update(func(tx *bolt.Tx) error{

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {


			// 删除
			for _,in := range ins {

				txOutputsBytes := b.Get(in.QYH_TxHash)

				if len(txOutputsBytes) == 0 {
					continue
				}

				fmt.Println("DeserializeTXOutputs")
				fmt.Println(txOutputsBytes)

				txOutputs := QYH_DeserializeTXOutputs(txOutputsBytes)

				fmt.Println(txOutputs)

				UTXOS := []*QYH_UTXO{}

				// 判断是否需要
				isNeedDelete := false

				for _,utxo := range txOutputs.QYH_UTXOS  {

					if in.QYH_Vout == utxo.QYH_Index && bytes.Compare(utxo.QYH_Output.QYH_Ripemd160Hash,QYH_Ripemd160Hash(in.QYH_PublicKey)) == 0 {

						isNeedDelete = true
					} else {
						UTXOS = append(UTXOS,utxo)
					}
				}



				if isNeedDelete {
					b.Delete(in.QYH_TxHash)
					if len(UTXOS) > 0 {

						preTXOutputs := outsMap[hex.EncodeToString(in.QYH_TxHash)]

						preTXOutputs.QYH_UTXOS = append(preTXOutputs.QYH_UTXOS, UTXOS...)

						outsMap[hex.EncodeToString(in.QYH_TxHash)] = preTXOutputs

					}
				}

			}

			// 新增

			for keyHash,outPuts := range outsMap  {
				keyHashBytes,_ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes,outPuts.QYH_Serialize())
			}

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}




