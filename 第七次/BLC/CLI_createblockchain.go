package BLC


// 创建创世区块
func (cli *QYH_CLI) QYH_createGenesisBlockchain(address string,nodeID string)  {

	blockchain := QYH_CreateBlockchainWithGenesisBlock(address,nodeID)
	defer blockchain.QYH_DB.Close()

	utxoSet := &QYH_UTXOSet{blockchain}

	utxoSet.QYH_ResetUTXOSet()
}

//blocks
//utxoTable