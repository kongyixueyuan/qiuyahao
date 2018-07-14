package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"math/big"
	"time"
	"os"
	"strconv"
	"encoding/hex"
	"crypto/ecdsa"
	"bytes"
)

// 数据库名字
const dbName = "blockchain_%s.db"

// 表的名字
const blockTableName = "blocks"

type QYH_Blockchain struct {
	QYH_Tip []byte //最新的区块的Hash
	QYH_DB  *bolt.DB
}

// 迭代器
func (blockchain *QYH_Blockchain) QYH_Iterator() *QYH_BlockchainIterator {

	return &QYH_BlockchainIterator{blockchain.QYH_Tip, blockchain.QYH_DB}
}

// 判断数据库是否存在
//3000
//blockchain_3000.db
func QYH_DBExists(dbName string) bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

// 遍历输出所有区块的信息
func (blc *QYH_Blockchain) QYH_Printchain() {

	fmt.Println("输出所有区块的信息....")
	blockchainIterator := blc.QYH_Iterator()

	for {
		fmt.Println("第一次进入for循环.....")
		block := blockchainIterator.QYH_Next()

		fmt.Printf("Height：%d\n", block.QYH_Height)
		fmt.Printf("PrevBlockHash：%x\n", block.QYH_PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.QYH_Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.QYH_Hash)
		fmt.Printf("Nonce：%d\n", block.QYH_Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.QYH_Txs {

			fmt.Printf("%x\n", tx.QYH_TxHash)
			fmt.Println("Vins:")
			for _, in := range tx.QYH_Vins {
				fmt.Printf("%x\n", in.QYH_TxHash)
				fmt.Printf("%d\n", in.QYH_Vout)
				fmt.Printf("%x\n", in.QYH_PublicKey)
			}

			fmt.Println("Vouts:")
			for _, out := range tx.QYH_Vouts {
				//fmt.Println(out.Value)
				fmt.Printf("%d\n",out.QYH_Value)
				//fmt.Println(out.Ripemd160Hash)
				fmt.Printf("%x\n",out.QYH_Ripemd160Hash)
			}
		}

		fmt.Println("------------------------------")

		var hashInt big.Int
		hashInt.SetBytes(block.QYH_PrevBlockHash)

		// Cmp compares x and y and returns:
		//
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break;
		}
	}

}

//// 增加区块到区块链里面
func (blc *QYH_Blockchain) QYH_AddBlockToBlockchain(txs []*QYH_Transaction) {

	err := blc.QYH_DB.Update(func(tx *bolt.Tx) error {

		//1. 获取表
		b := tx.Bucket([]byte(blockTableName))
		//2. 创建新区块
		if b != nil {

			// ⚠️，先获取最新区块
			blockBytes := b.Get(blc.QYH_Tip)
			// 反序列化
			block := QYH_DeserializeBlock(blockBytes)

			//3. 将区块序列化并且存储到数据库中
			newBlock := QYH_NewBlock(txs, block.QYH_Height+1, block.QYH_Hash)
			err := b.Put(newBlock.QYH_Hash, newBlock.QYH_Serialize())
			if err != nil {
				log.Panic(err)
			}
			//4. 更新数据库里面"l"对应的hash
			err = b.Put([]byte("l"), newBlock.QYH_Hash)
			if err != nil {
				log.Panic(err)
			}
			//5. 更新blockchain的Tip
			blc.QYH_Tip = newBlock.QYH_Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//1. 创建带有创世区块的区块链
func QYH_CreateBlockchainWithGenesisBlock(address string,nodeID string) *QYH_Blockchain {

	// 格式化数据库名字
	dbName := fmt.Sprintf(dbName,nodeID)


	// 判断数据库是否存在
	if QYH_DBExists(dbName) {
		fmt.Println("创世区块已经存在.......")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块.......")

	// 创建或者打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var genesisHash []byte

	// 关闭数据库
	err = db.Update(func(tx *bolt.Tx) error {

		// 创建数据库表
		b, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if b != nil {
			// 创建创世区块
			// 创建了一个coinbase Transaction
			txCoinbase := QYH_NewCoinbaseTransaction(address)

			genesisBlock := QYH_CreateGenesisBlock([]*QYH_Transaction{txCoinbase})
			// 将创世区块存储到表中
			err := b.Put(genesisBlock.QYH_Hash, genesisBlock.QYH_Serialize())
			if err != nil {
				log.Panic(err)
			}

			// 存储最新的区块的hash
			err = b.Put([]byte("l"), genesisBlock.QYH_Hash)
			if err != nil {
				log.Panic(err)
			}

			genesisHash = genesisBlock.QYH_Hash
		}

		return nil
	})

	return &QYH_Blockchain{genesisHash, db}

}

// 返回Blockchain对象
func QYH_BlockchainObject(nodeID string) *QYH_Blockchain {

	dbName := fmt.Sprintf(dbName,nodeID)

	// 判断数据库是否存在
	if QYH_DBExists(dbName) == false {
		fmt.Println("数据库不存在....")
		os.Exit(1)
	}

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			// 读取最新区块的Hash
			tip = b.Get([]byte("l"))

		}

		return nil
	})

	return &QYH_Blockchain{tip, db}
}

// 如果一个地址对应的TXOutput未花费，那么这个Transaction就应该添加到数组中返回
func (blockchain *QYH_Blockchain) QYH_UnUTXOs(address string,txs []*QYH_Transaction) []*QYH_UTXO {

	var unUTXOs []*QYH_UTXO

	spentTXOutputs := make(map[string][]int)

	//{hash:[0]}

	for _,tx := range txs {

		if tx.QYH_IsCoinbaseTransaction() == false {
			for _, in := range tx.QYH_Vins {
				//是否能够解锁
				publicKeyHash := QYH_Base58Decode([]byte(address))

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
		for index,out := range tx.QYH_Vouts {

			if out.QYH_UnLockScriptPubKeyWithAddress(address) {
				fmt.Println("看看是否是俊诚...")
				fmt.Println(address)

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


	blockIterator := blockchain.QYH_Iterator()

	for {

		block := blockIterator.QYH_Next()

		fmt.Println(block)
		fmt.Println()

		for i := len(block.QYH_Txs) - 1; i >= 0 ; i-- {

			tx := block.QYH_Txs[i]
			// txHash
			// Vins
			if tx.QYH_IsCoinbaseTransaction() == false {
				for _, in := range tx.QYH_Vins {
					//是否能够解锁
					publicKeyHash := QYH_Base58Decode([]byte(address))

					ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]

					if in.QYH_UnLockRipemd160Hash(ripemd160Hash) {

						key := hex.EncodeToString(in.QYH_TxHash)

						spentTXOutputs[key] = append(spentTXOutputs[key], in.QYH_Vout)
					}

				}
			}

			// Vouts

		work:
			for index, out := range tx.QYH_Vouts {

				if out.QYH_UnLockScriptPubKeyWithAddress(address) {

					fmt.Println(out)
					fmt.Println(spentTXOutputs)

					//&{2 zhangqiang}
					//map[]

					if spentTXOutputs != nil {

						//map[cea12d33b2e7083221bf3401764fb661fd6c34fab50f5460e77628c42ca0e92b:[0]]

						if len(spentTXOutputs) != 0 {

							var isSpentUTXO bool

							for txHash, indexArray := range spentTXOutputs {

								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.QYH_TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}

							if isSpentUTXO == false {

								utxo := &QYH_UTXO{tx.QYH_TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)

							}
						} else {
							utxo := &QYH_UTXO{tx.QYH_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}

					}
				}

			}

		}

		fmt.Println(spentTXOutputs)

		var hashInt big.Int
		hashInt.SetBytes(block.QYH_PrevBlockHash)

		// Cmp compares x and y and returns:
		//
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}

	}

	return unUTXOs
}

// 转账时查找可用的UTXO
func (blockchain *QYH_Blockchain) QYH_FindSpendableUTXOS(from string, amount int,txs []*QYH_Transaction) (int64, map[string][]int) {

	//1. 现获取所有的UTXO

	utxos := blockchain.QYH_UnUTXOs(from,txs)

	spendableUTXO := make(map[string][]int)

	//2. 遍历utxos

	var value int64

	for _, utxo := range utxos {

		value = value + utxo.QYH_Output.QYH_Value

		hash := hex.EncodeToString(utxo.QYH_TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.QYH_Index)

		if value >= int64(amount) {
			break
		}
	}

	if value < int64(amount) {

		fmt.Printf("%s's fund is 不足\n", from)
		os.Exit(1)
	}

	return value, spendableUTXO
}

// 挖掘新的区块
func (blockchain *QYH_Blockchain) QYH_MineNewBlock(from []string, to []string, amount []string, nodeID string) {

	//	$ ./bc send -from '["juncheng"]' -to '["zhangqiang"]' -amount '["2"]'
	//	[juncheng]
	//	[zhangqiang]
	//	[2]

	//1.建立一笔交易


	utxoSet := &QYH_UTXOSet{blockchain}

	var txs []*QYH_Transaction

	for index,address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := QYH_NewSimpleTransaction(address, to[index], int64(value), utxoSet, txs, nodeID)
		txs = append(txs, tx)
		//fmt.Println(tx)
	}

	//奖励
	tx := QYH_NewCoinbaseTransaction(from[0])
	txs = append(txs,tx)


	//1. 通过相关算法建立Transaction数组
	var block *QYH_Block

	blockchain.QYH_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte("l"))

			blockBytes := b.Get(hash)

			block = QYH_DeserializeBlock(blockBytes)

		}

		return nil
	})


	// 在建立新区块之前对txs进行签名验证

	_txs := []*QYH_Transaction{}

	for _,tx := range txs  {

		if blockchain.QYH_VerifyTransaction(tx,_txs) != true {
			log.Panic("ERROR: Invalid transaction")
		}

		_txs = append(_txs,tx)
	}


	//2. 建立新的区块
	block = QYH_NewBlock(txs, block.QYH_Height+1, block.QYH_Hash)

	//将新区块存储到数据库
	blockchain.QYH_DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			b.Put(block.QYH_Hash, block.QYH_Serialize())

			b.Put([]byte("l"), block.QYH_Hash)

			blockchain.QYH_Tip = block.QYH_Hash

		}
		return nil
	})

}

// 查询余额
func (blockchain *QYH_Blockchain) QYH_GetBalance(address string) int64 {

	utxos := blockchain.QYH_UnUTXOs(address,[]*QYH_Transaction{})

	var amount int64

	for _, utxo := range utxos {

		amount = amount + utxo.QYH_Output.QYH_Value
	}

	return amount
}

func (bclockchain *QYH_Blockchain) QYH_SignTransaction(tx *QYH_Transaction,privKey ecdsa.PrivateKey,txs []*QYH_Transaction)  {

	if tx.QYH_IsCoinbaseTransaction() {
		return
	}

	prevTXs := make(map[string]QYH_Transaction)

	for _, vin := range tx.QYH_Vins {
		prevTX, err := bclockchain.QYH_FindTransaction(vin.QYH_TxHash,txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.QYH_TxHash)] = prevTX
	}

	tx.QYH_Sign(privKey, prevTXs)

}


func (bc *QYH_Blockchain) QYH_FindTransaction(ID []byte,txs []*QYH_Transaction) (QYH_Transaction, error) {


	for _,tx := range txs  {
		if bytes.Compare(tx.QYH_TxHash, ID) == 0 {
			return *tx, nil
		}
	}


	bci := bc.QYH_Iterator()

	for {
		block := bci.QYH_Next()

		for _, tx := range block.QYH_Txs {
			if bytes.Compare(tx.QYH_TxHash, ID) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.QYH_PrevBlockHash)


		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break;
		}
	}

	return QYH_Transaction{},nil
}


// 验证数字签名
func (bc *QYH_Blockchain) QYH_VerifyTransaction(tx *QYH_Transaction,txs []*QYH_Transaction) bool {


	prevTXs := make(map[string]QYH_Transaction)

	for _, vin := range tx.QYH_Vins {
		prevTX, err := bc.QYH_FindTransaction(vin.QYH_TxHash,txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.QYH_TxHash)] = prevTX
	}

	return tx.QYH_Verify(prevTXs)
}


// [string]*TXOutputs
func (blc *QYH_Blockchain) QYH_FindUTXOMap() map[string]*QYH_TXOutputs  {

	blcIterator := blc.QYH_Iterator()

	// 存储已花费的UTXO的信息
	spentableUTXOsMap := make(map[string][]*QYH_TXInput)


	utxoMaps := make(map[string]*QYH_TXOutputs)


	for {
		block := blcIterator.QYH_Next()

		for i := len(block.QYH_Txs) - 1; i >= 0 ;i-- {

			txOutputs := &QYH_TXOutputs{[]*QYH_UTXO{}}

			tx := block.QYH_Txs[i]

			// coinbase
			if tx.QYH_IsCoinbaseTransaction() == false {
				for _,txInput := range tx.QYH_Vins {

					txHash := hex.EncodeToString(txInput.QYH_TxHash)
					spentableUTXOsMap[txHash] = append(spentableUTXOsMap[txHash], txInput)

				}
			}

			txHash := hex.EncodeToString(tx.QYH_TxHash)

			txInputs := spentableUTXOsMap[txHash]

			if len(txInputs) > 0 {


			WorkOutLoop:
				for index,out := range tx.QYH_Vouts  {

					for _,in := range  txInputs {

						outPublicKey := out.QYH_Ripemd160Hash
						inPublicKey := in.QYH_PublicKey


						if bytes.Compare(outPublicKey, QYH_Ripemd160Hash(inPublicKey)) == 0 {
							if index == in.QYH_Vout {

								continue WorkOutLoop
							} else {

								utxo := &QYH_UTXO{tx.QYH_TxHash,index,out}
								txOutputs.QYH_UTXOS = append(txOutputs.QYH_UTXOS,utxo)
							}
						}
					}


				}

			} else {

				for index,out := range tx.QYH_Vouts {
					utxo := &QYH_UTXO{tx.QYH_TxHash,index,out}
					txOutputs.QYH_UTXOS = append(txOutputs.QYH_UTXOS,utxo)
				}
			}


			// 设置键值对
			utxoMaps[txHash] = txOutputs

		}


		// 找到创世区块时退出
		var hashInt big.Int
		hashInt.SetBytes(block.QYH_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}

	}

	return utxoMaps
}



//----------

func (bc *QYH_Blockchain) QYH_GetBestHeight() int64 {

	block := bc.QYH_Iterator().QYH_Next()

	return block.QYH_Height
}

func (bc *QYH_Blockchain) QYH_GetBlockHashes() [][]byte {

	blockIterator := bc.QYH_Iterator()

	var blockHashs [][]byte

	for {
		block := blockIterator.QYH_Next()

		blockHashs = append(blockHashs,block.QYH_Hash)

		var hashInt big.Int
		hashInt.SetBytes(block.QYH_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break;
		}
	}

	return blockHashs
}

func (bc *QYH_Blockchain) QYH_GetBlock(blockHash []byte) ([]byte ,error) {

	var blockBytes []byte

	err := bc.QYH_DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {

			blockBytes = b.Get(blockHash)

		}

		return nil
	})

	return blockBytes,err
}

func (bc *QYH_Blockchain) QYH_AddBlock(block *QYH_Block)  {

	err := bc.QYH_DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {

			blockExist := b.Get(block.QYH_Hash)

			if blockExist != nil {
				// 如果存在，不需要做任何过多的处理
				return nil
			}

			err := b.Put(block.QYH_Hash,block.QYH_Serialize())

			if err != nil {
				log.Panic(err)
			}

			// 最新的区块链的Hash
			blockHash := b.Get([]byte("l"))

			blockBytes := b.Get(blockHash)

			blockInDB := QYH_DeserializeBlock(blockBytes)

			if blockInDB.QYH_Height < block.QYH_Height {

				b.Put([]byte("l"),block.QYH_Hash)
				bc.QYH_Tip = block.QYH_Hash
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}