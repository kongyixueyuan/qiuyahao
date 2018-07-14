package BLC


func (cli *QYH_CLI) QYH_resetUTXOSet(nodeID string)  {

	blockchain := QYH_BlockchainObject(nodeID)

	defer blockchain.QYH_DB.Close()

	utxoSet := &QYH_UTXOSet{blockchain}

	utxoSet.QYH_ResetUTXOSet()

}
