package BLC

import (
	"bytes"
	"log"
	"encoding/gob"
	"fmt"
)

func QYH_handleVersion(request []byte,bc *QYH_Blockchain)  {

	var buff bytes.Buffer
	var payload QYH_Version

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	//Version
	//1. Version
	//2. BestHeight
	//3. 节点地址

	bestHeight := bc.QYH_GetBestHeight() //3
	foreignerBestHeight := payload.QYH_BestHeight // 1

	if bestHeight > foreignerBestHeight {
		QYH_sendVersion(payload.QYH_AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		// 去向主节点要信息
		QYH_sendGetBlocks(payload.QYH_AddrFrom)
	}

}

func QYH_handleAddr(request []byte,bc *QYH_Blockchain)  {


}

func QYH_handleGetblocks(request []byte,bc *QYH_Blockchain)  {


	var buff bytes.Buffer
	var payload QYH_GetBlocks

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.QYH_GetBlockHashes()

	//
	QYH_sendInv(payload.QYH_AddrFrom, BLOCK_TYPE, blocks)


}

func QYH_handleGetData(request []byte,bc *QYH_Blockchain)  {

	var buff bytes.Buffer
	var payload QYH_GetData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.QYH_Type == BLOCK_TYPE {

		block, err := bc.QYH_GetBlock([]byte(payload.QYH_Hash))
		if err != nil {
			return
		}

		QYH_sendBlock(payload.QYH_AddrFrom, block)
	}

	if payload.QYH_Type == "tx" {

	}
}

func QYH_handleBlock(request []byte,bc *QYH_Blockchain)  {
	var buff bytes.Buffer
	var payload QYH_BlockData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockBytes := payload.QYH_Block

	block := QYH_DeserializeBlock(blockBytes)

	fmt.Println("Recevied a new block!")
	bc.QYH_AddBlock(block)

	fmt.Printf("Added block %x\n", block.QYH_Hash)

	if len(transactionArray) > 0 {
		blockHash := transactionArray[0]
		QYH_sendGetData(payload.QYH_AddrFrom, "block", blockHash)

		transactionArray = transactionArray[1:]
	} else {

		fmt.Println("数据库重置......")
		UTXOSet := &QYH_UTXOSet{bc}
		UTXOSet.QYH_ResetUTXOSet()

	}

}

func QYH_handleTx(request []byte,bc *QYH_Blockchain)  {

}


func QYH_handleInv(request []byte,bc *QYH_Blockchain)  {

	var buff bytes.Buffer
	var payload QYH_Inv

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Ivn 3000 block hashes [][]

	if payload.QYH_Type == BLOCK_TYPE {

		//tansactionArray = payload.Items

		//payload.Items

		blockHash := payload.QYH_Items[0]
		QYH_sendGetData(payload.QYH_AddrFrom, BLOCK_TYPE , blockHash)

		if len(payload.QYH_Items) >= 1 {
			transactionArray = payload.QYH_Items[1:]
		}
	}

	if payload.QYH_Type == TX_TYPE {

	}

}