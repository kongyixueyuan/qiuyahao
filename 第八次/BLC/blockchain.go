package BLC

import (
	"github.com/boltdb/bolt"
	"os"
	"fmt"
	"log"
	"encoding/hex"
	"strconv"
	"crypto/ecdsa"
	"bytes"
	"errors"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "genesis data 08/07/2018 by viky"

type QYH_Blockchain struct {
	QYH_tip []byte
	QYH_db  *bolt.DB
}

// 打印区块链内容
func (bc *QYH_Blockchain) QYH_Printchain() {
	bci := bc.QYH_Iterator()

	for {
		block := bci.QYH_Next()
		block.QYH_String()
		if len(block.QYH_PrevBlockHash) == 0 {
			break
		}
	}

}

// 通过交易hash,查找交易
func (bc *QYH_Blockchain) QYH_FindTransaction(ID []byte) (QYH_Transaction, error) {
	bci := bc.QYH_Iterator()
	for {
		block := bci.QYH_Next()
		for _, tx := range block.QYH_Transactions {
			if bytes.Compare(tx.QYH_ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.QYH_PrevBlockHash) == 0 {
			break
		}
	}
	fmt.Printf("查找%x的交易失败",ID)
	return QYH_Transaction{}, errors.New("未找到交易")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *QYH_Blockchain) QYH_FindUTXO() map[string]QYH_TXOutputs {
	// 未花费的交易输出
	// key:交易hash   txID
	UTXO := make(map[string]QYH_TXOutputs)
	// 已经花费的交易txID : TXOutputs.index
	spentTXOs := make(map[string][]int)
	bci := bc.QYH_Iterator()

	for {
		block := bci.QYH_Next()

		// 循环区块中的交易
		for _, tx := range block.QYH_Transactions {
			// 将区块中的交易hash，转为字符串
			txID := hex.EncodeToString(tx.QYH_ID)

		Outputs:
			for outIdx, out := range tx.QYH_Vout { // 循环交易中的 TXOutputs
				// Was the output spent?
				// 如果已经花费的交易输出中，有此输出，证明已经花费
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx { // 如果花费的正好是此笔输出
							continue Outputs // 继续下一次循环
						}
					}
				}

				outs := UTXO[txID] // 获取UTXO指定txID对应的TXOutputs
				outs.QYH_Outputs = append(outs.QYH_Outputs, out)
				UTXO[txID] = outs
			}

			if tx.QYH_IsCoinbase() == false { // 非创世区块
				for _, in := range tx.QYH_Vin {
					inTxID := hex.EncodeToString(in.QYH_Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.QYH_Vout)
				}
			}
		}
		// 如果上一区块的hash为0，代表已经到创世区块，循环结束
		if len(block.QYH_PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// 获取迭代器
func (bc *QYH_Blockchain) QYH_Iterator() *QYH_BlockchainIterator {
	return &QYH_BlockchainIterator{bc.QYH_tip, bc.QYH_db}
}

// 新建区块链(包含创世区块)
func QYH_CreateBlockchain(address string, nodeID string) *QYH_Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if QYH_dbExists(dbFile) {
		fmt.Println("blockchain数据库已经存在.")
		os.Exit(1)
	}

	var tip []byte
	cbtx := QYH_NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := QYH_NewGenesisBlock(cbtx)

	//genesis.String()

	// 打开数据库，如果不存在自动创建
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		// 新区块存入数据库
		err = b.Put(genesis.QYH_Hash, genesis.QYH_Serialize())
		if err != nil {
			log.Panic(err)
		}
		// 将创世区块的hash存入数据库
		err = b.Put([]byte("l"), genesis.QYH_Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.QYH_Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &QYH_Blockchain{tip, db}
}

// 获取blockchain对象
func QYH_NewBlockchain(nodeID string) *QYH_Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if !QYH_dbExists(dbFile) {
		log.Panic("区块链还未创建")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &QYH_Blockchain{tip, db}
}

// 生成新的区块(挖矿)
func (bc *QYH_Blockchain) QYH_MineNewBlock(from []string, to []string, amount []string,nodeID string , mineNow bool) *QYH_Block {
	UTXOSet := QYH_UTXOSet{bc}

	wallets, err := QYH_NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	var txs []*QYH_Transaction

	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		if value<=0 {
			log.Panic("错误：转账金额需要大于0")
		}
		wallet := wallets.QYH_GetWallet(address)
		tx := QYH_NewUTXOTransaction(&wallet, to[index], value, &UTXOSet, txs)
		txs = append(txs, tx)
	}

	if mineNow {
		// 挖矿奖励
		tx := QYH_NewCoinbaseTX(from[0], "")
		txs = append(txs, tx)

		//=====================================
		newBlock := bc.QYH_MineBlock(txs)
		UTXOSet.QYH_Update(newBlock)
		return newBlock
	}else{
		// 如果不立即挖矿，将交易写到内存中
		//var txs_all []Yxh_Transaction
		//for _,value := range txs{
		//	txs_all= append(txs_all, *value)
		//}
		// 当前节点的IP地址
		nodeAddress = fmt.Sprintf("localhost:%s",nodeID)
		QYH_sendTxs(knownNodes[0],txs)
		return nil
	}


}

// 挖矿
func (bc *QYH_Blockchain) QYH_MineBlock(txs []*QYH_Transaction) *QYH_Block  {
	var lashHash []byte
	var lastHeight int

	// 检查交易是否有效，验证签名
	for _, tx := range txs {
		if !bc.QYH_VerifyTransaction(tx, txs) {
			log.Panic("错误：无效的交易")
		}
	}
	// 获取最后一个区块的hash,然后获取最后一个区块的信息，进而获得height
	err := bc.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lashHash = b.Get([]byte("l"))
		blockData := b.Get(lashHash)
		block := QYH_DeserializeBlock(blockData)
		lastHeight = block.QYH_Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	// 生成新的区块
	newBlock := QYH_NewBlock(txs, lashHash, lastHeight+1)

	// 将新区块的内容更新到数据库中
	err = bc.QYH_db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.QYH_Hash, newBlock.QYH_Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.QYH_Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.QYH_tip = newBlock.QYH_Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return newBlock
}

// 签名
func (bc *QYH_Blockchain) QYH_SignTransaction(tx *QYH_Transaction, privKey ecdsa.PrivateKey,txs []*QYH_Transaction) {
	prevTXs := make(map[string]QYH_Transaction)

	// 找到交易输入中，之前的交易
	Vin:
	for _, vin := range tx.QYH_Vin {
		for _, tx := range txs {
			if bytes.Compare(tx.QYH_ID, vin.QYH_Txid) == 0 {
				prevTX := *tx
				prevTXs[hex.EncodeToString(prevTX.QYH_ID)] = prevTX
				continue Vin
			}
		}

		prevTX, err := bc.QYH_FindTransaction(vin.QYH_Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.QYH_ID)] = prevTX

	}

	tx.QYH_Sign(privKey, prevTXs)
}

// 验证签名
func (bc *QYH_Blockchain) QYH_VerifyTransaction(tx *QYH_Transaction,txs []*QYH_Transaction) bool {
	if tx.QYH_IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]QYH_Transaction)
	Vin:
	for _, vin := range tx.QYH_Vin {
		for _, tx := range txs {
			if bytes.Compare(tx.QYH_ID, vin.QYH_Txid) == 0 {
				prevTX := *tx
				prevTXs[hex.EncodeToString(prevTX.QYH_ID)] = prevTX
				continue Vin
			}
		}
		prevTX, err := bc.QYH_FindTransaction(vin.QYH_Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.QYH_ID)] = prevTX
	}

	return tx.QYH_Verify(prevTXs)
}

// 判断数据库是否存在
func QYH_dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// 获取BestHeight
func (bc *QYH_Blockchain) QYH_GetBestHeight() int {
	var lastBlock QYH_Block

	err := bc.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *QYH_DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.QYH_Height
}

// 获取所有区块的hash
func (bc *QYH_Blockchain) QYH_GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.QYH_Iterator()

	for {
		block := bci.QYH_Next()

		blocks = append(blocks, block.QYH_Hash)

		if len(block.QYH_PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// 根据hash获取某个区块的内容
func (bc *QYH_Blockchain) QYH_GetBlock(blockHash []byte) (QYH_Block, error) {
	var block QYH_Block

	err := bc.QYH_db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("未找到区块")
		}

		block = *QYH_DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// 将区块添加到链中
func (bc *QYH_Blockchain) QYH_AddBlock(block *QYH_Block) {
	err := bc.QYH_db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.QYH_Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.QYH_Serialize()
		err := b.Put(block.QYH_Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := QYH_DeserializeBlock(lastBlockData)

		if block.QYH_Height > lastBlock.QYH_Height {
			err = b.Put([]byte("l"), block.QYH_Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.QYH_tip = block.QYH_Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}