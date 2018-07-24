package BLC

import (
	"fmt"
	"net"
	"log"
	"bytes"
	"encoding/gob"
	"io"
	"io/ioutil"
	"encoding/hex"
)

const protocol = "tcp"   // 节点协议
const nodeVersion = 1    // 节点版本
const commandLength = 12 // 命令长度

var nodeAddress string                         // 当前节点地址
var miningAddress string                       // 挖矿地址
var knownNodes = []string{"localhost:3000"}    // 存储所有已知节点
var blocksInTransit = [][]byte{}               // 存储接收到的区块hash
var mempool = make(map[string]QYH_Transaction) // 存储接收到的交易

type QYH_addr struct {
	QYH_AddrList []string
}

type QYH_block struct {
	QYH_AddrFrom string
	QYH_Block    []byte
}

type QYH_getblocks struct {
	QYH_AddrFrom string
}

type QYH_getdata struct {
	QYH_AddrFrom string
	QYH_Type     string
	QYH_ID       []byte
}

type QYH_inv struct {
	QYH_AddrFrom string
	QYH_Type     string
	QYH_Items    [][]byte
}

type QYH_txs struct {
	QYH_AddFrom     string
	QYH_Transactions [][]byte
}

type QYH_version struct {
	QYH_Version    int
	QYH_BestHeight int
	QYH_AddrFrom   string
}

//启动Server
func QYH_StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := QYH_NewBlockchain(nodeID)

	// 3000端口为：主节点
	// 3001端口为：钱包节点
	// 3002端口为：挖矿节点
	if nodeAddress != knownNodes[0] {
		// 此节点是钱包节点或者矿工节点，需要向主节点发送请求同步数据
		QYH_sendVersion(knownNodes[0], bc)
	}

	for { // 等待接收命令
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go QYH_handleConnecton(conn, bc)
	}
}

// 向中心节点发送 version 消息来查询是否自己的区块链已过时
func QYH_sendVersion(addr string, bc *QYH_Blockchain) {
	bestHeight := bc.QYH_GetBestHeight()
	payload := QYH_gobEncode(QYH_version{nodeVersion, bestHeight, nodeAddress})

	request := append(QYH_commandToBytes("version"), payload...)

	QYH_sendData(addr, request)
}

// 发送数据
func QYH_sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	// 如果连接失败，更新节点数据
	if err != nil {
		fmt.Sprintf("%s地址不可用\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

}

// 发送获取区块的的命令
func QYH_sendGetBlocks(address string) {
	payload := QYH_gobEncode(QYH_getblocks{nodeAddress})
	request := append(QYH_commandToBytes("getblocks"), payload...)

	QYH_sendData(address, request)
}

// 发送处理区块及交易hash列表请求
func QYH_sendInv(address, kind string, items [][]byte) {
	inventory := QYH_inv{nodeAddress, kind, items}
	payload := QYH_gobEncode(inventory)
	request := append(QYH_commandToBytes("inv"), payload...)

	QYH_sendData(address, request)
}

// 发送获取区块详细数据的命令
func QYH_sendGetData(address, kind string, id []byte) {
	payload := QYH_gobEncode(QYH_getdata{nodeAddress, kind, id})
	request := append(QYH_commandToBytes("getdata"), payload...)

	QYH_sendData(address, request)
}

// 发送区块内容命令
func QYH_sendBlock(addr string, b *QYH_Block) {
	data := QYH_block{nodeAddress, b.QYH_Serialize()}
	payload := QYH_gobEncode(data)
	request := append(QYH_commandToBytes("block"), payload...)

	QYH_sendData(addr, request)
}

// 发送交易内容命令
func QYH_sendTx(addr string, tx *QYH_Transaction) {
	txs := []*QYH_Transaction{tx}
	QYH_sendTxs(addr,txs)
}
// 发送交易内容命令
func QYH_sendTxs(addr string, txs []*QYH_Transaction) {

	data := QYH_txs{nodeAddress, QYH_SerializeTransactions(txs)}
	payload := QYH_gobEncode(data)
	request := append(QYH_commandToBytes("tx"), payload...)

	QYH_sendData(addr, request)
}

//================================================================
// 命令收集并分发
func QYH_handleConnecton(conn net.Conn, bc *QYH_Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := QYH_bytesToCommand(request[:commandLength])
	fmt.Printf("接收到命令：%s\n", command)

	switch command {
	case "addr": // 添加新节点
		QYH_handleAddr(request)
	case "block": // 添加新区块
		QYH_handleBlock(request, bc)
	case "inv": // 接收区块及交易hash列表 ，区块获取到内容到存储到 blocksInTransit ， 交易存储到 mempool
		QYH_handleInv(request, bc)
	case "getblocks": // 将当前节点区块链中的所有区块hash，返回给请求节点
		QYH_handleGetBlocks(request, bc)
	case "getdata": // 将单个交易或区块的内容 返回给请求节点
		QYH_handleGetData(request, bc)
	case "tx": // 添加新的交易,交易数量大于2，矿工节点挖矿,如果是主节点，进行分发交易
		QYH_handleTx(request, bc)
	case "version": // 检查是否需要同步数据，根据区块的height
		QYH_handleVersion(request, bc)
	default:
		fmt.Println("未知命令!")
	}

	conn.Close()

}

// 添加新节点
func QYH_handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload QYH_addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.QYH_AddrList...)
	fmt.Printf("有%d个节点加入\n", len(knownNodes))
	// 把新节点推送给其他所有节点
	for _, node := range knownNodes {
		QYH_sendGetBlocks(node)
	}
}

/*
当接收到一个新块时，我们把它放到区块链里面。
如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求。
当最后把所有块都下载完后，对 UTXO 集进行重新索引

TODO: 并非无条件信任，我们应该在将每个块加入到区块链之前对它们进行验证。
TODO: 并非运行 UTXOSet.Reindex()， 而是应该使用 UTXOSet.Update(block)，
TODO: 因为如果区块链很大，它将需要很多时间来对整个 UTXO 集重新索引。
 */
func QYH_handleBlock(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.QYH_Block
	block := QYH_DeserializeBlock(blockData)

	fmt.Println("收到一个新的区块!")
	bc.QYH_AddBlock(block)

	fmt.Printf("Added block %x\n", block.QYH_Hash)

	// 如果还有更多的区块需要下载，继续从上一个下载的块的那个节点继续请求
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		QYH_sendGetData(payload.QYH_AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := QYH_UTXOSet{bc}
		UTXOSet.QYH_Reset()
	}
}

// 向其他节点展示当前节点有什么块和交易,接收区块及交易列表
func QYH_handleInv(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("接收到列表,长度为：%d，类型： %s\n", len(payload.QYH_Items), payload.QYH_Type)

	// 如果数据是 区块
	if payload.QYH_Type == "block" {
		blocksInTransit = payload.QYH_Items

		blockHash := payload.QYH_Items[0]
		// 发送获取区块详细数据的命令
		QYH_sendGetData(payload.QYH_AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	// 如果数据是 交易
	// 转账时，未立即挖矿，将交易保存到内存池中
	if payload.QYH_Type == "tx" {
		txID := payload.QYH_Items[0]
		// 如果内存池中，交易内容为空
		if mempool[hex.EncodeToString(txID)].QYH_ID == nil {
			// 发送获取交易详细内容请求
			QYH_sendGetData(payload.QYH_AddrFrom, "tx", txID)
		}
	}
}

// 处理获取区块命令，返回区块链中的所有区块hash
func QYH_handleGetBlocks(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.QYH_GetBlockHashes()
	QYH_sendInv(payload.QYH_AddrFrom, "block", blocks)
}

//  将单个交易或区块的内容 返回给请求节点
func QYH_handleGetData(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.QYH_Type == "block" {
		block, err := bc.QYH_GetBlock([]byte(payload.QYH_ID))
		if err != nil {
			return
		}

		QYH_sendBlock(payload.QYH_AddrFrom, &block)
	}

	if payload.QYH_Type == "tx" {
		txID := hex.EncodeToString(payload.QYH_ID)
		tx := mempool[txID]

		QYH_sendTx(payload.QYH_AddrFrom, &tx)
		// delete(mempool, txID)
	}
}

// 处理交易
func QYH_handleTx(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_txs

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.QYH_Transactions
	txsDes := QYH_DeserializeTransactions(txData)

	for _,tx := range txsDes {
		mempool[hex.EncodeToString(tx.QYH_ID)] = tx

		// 如果是主节点
		if nodeAddress == knownNodes[0] {
			for _, node := range knownNodes {
				// 给其他节点分发，添加交易
				if node != nodeAddress && node != payload.QYH_AddFrom {
					QYH_sendInv(node, "tx", [][]byte{tx.QYH_ID})
				}
			}
		} else {
			// 如果交易池中有两条交易 并且当前是挖矿节点
			if len(mempool) >= 2 && len(miningAddress) > 0 {
			MineTransactions:
				var txs []*QYH_Transaction

				for id := range mempool {
					tx := mempool[id]
					if bc.QYH_VerifyTransaction(&tx, txs) {
						txs = append(txs, &tx)
					}
				}

				if len(txs) == 0 {
					fmt.Println("交易不可用...")
					break
				}

				cbTx := QYH_NewCoinbaseTX(miningAddress, "")
				txs = append(txs, cbTx)

				newBlock := bc.QYH_MineBlock(txs)
				UTXOSet := QYH_UTXOSet{bc}
				UTXOSet.QYH_Update(newBlock)

				fmt.Println("挖到新的区块!")

				for _, tx := range txs {
					txID := hex.EncodeToString(tx.QYH_ID)
					delete(mempool, txID)
				}

				for _, node := range knownNodes {
					if node != nodeAddress {
						QYH_sendInv(node, "block", [][]byte{newBlock.QYH_Hash})
					}
				}

				if len(mempool) > 0 {
					goto MineTransactions
				}
			}
		}
	}
}

// 检查是否需要同步数据
func QYH_handleVersion(request []byte, bc *QYH_Blockchain) {
	var buff bytes.Buffer
	var payload QYH_version
	// 获取数据
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	// 获取当前节点的最后height
	myBestHeight := bc.QYH_GetBestHeight()
	// 数据中的最后height
	foreignerBestHeight := payload.QYH_BestHeight

	// 节点将从消息中提取的 BestHeight 与自身进行比较
	// 当前的height比对方的小
	// 发送获取区块的的命令
	if myBestHeight < foreignerBestHeight {
		QYH_sendGetBlocks(payload.QYH_AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		// 当前的height比对方的大
		// 通知对方节点，同步数据
		QYH_sendVersion(payload.QYH_AddrFrom, bc)
	}

	// 如果不是已知节点，将节点添加到已知节点中
	if !QYH_nodeIsKnown(payload.QYH_AddrFrom) {
		knownNodes = append(knownNodes, payload.QYH_AddrFrom)
	}
}

// 是否是已知节点
func QYH_nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

//================================================================

// 命令转字节数组
func QYH_commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

// 将字节数组转字符串命令
func QYH_bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

// 加密
func QYH_gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
