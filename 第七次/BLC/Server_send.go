package BLC

import (
	"io"
	"bytes"
	"log"
	"net"
)


//COMMAND_VERSION
func QYH_sendVersion(toAddress string,bc *QYH_Blockchain)  {


	bestHeight := bc.QYH_GetBestHeight()

	payload := gobEncode(QYH_Version{NODE_VERSION, bestHeight, nodeAddress})

	//version
	request := append(commandToBytes(COMMAND_VERSION), payload...)

	QYH_sendData(toAddress,request)

}



//COMMAND_GETBLOCKS
func QYH_sendGetBlocks(toAddress string)  {

	payload := gobEncode(QYH_GetBlocks{nodeAddress})

	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)

	QYH_sendData(toAddress,request)

}

// 主节点将自己的所有的区块hash发送给钱包节点
//COMMAND_BLOCK
//
func QYH_sendInv(toAddress string, kind string, hashes [][]byte) {

	payload := gobEncode(QYH_Inv{nodeAddress,kind,hashes})

	request := append(commandToBytes(COMMAND_INV), payload...)

	QYH_sendData(toAddress,request)

}



func QYH_sendGetData(toAddress string, kind string ,blockHash []byte) {

	payload := gobEncode(QYH_GetData{nodeAddress,kind,blockHash})

	request := append(commandToBytes(COMMAND_GETDATA), payload...)

	QYH_sendData(toAddress,request)
}



func QYH_sendBlock(toAddress string, block []byte)  {


	payload := gobEncode(QYH_BlockData{nodeAddress,block})

	request := append(commandToBytes(COMMAND_BLOCK), payload...)

	QYH_sendData(toAddress,request)

}


func QYH_sendData(to string,data []byte)  {

	conn, err := net.Dial("tcp", to)
	if err != nil {
		panic("error")
	}
	defer conn.Close()

	// 附带要发送的数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}